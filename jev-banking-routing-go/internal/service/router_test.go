package service_test

import (
	"testing"

	"github.com/example/jev-banking-routing-go/internal/domain"
	"github.com/example/jev-banking-routing-go/internal/routing"
)

func ptr(v float64) *float64 { return &v }

func baseRequest() domain.EvaluationRequest {
	return domain.EvaluationRequest{
		Transaction: domain.Transaction{
			TransactionID:       "txn-test",
			AmountINR:           5000,
			Currency:            "INR",
			AvailableBalanceINR: 20000,
		},
		CustomerHistory: domain.CustomerHistory{CustomerID: "customer-test"},
	}
}

func TestSanctionsAlwaysDeclines(t *testing.T) {
	in := baseRequest()
	in.Transaction.SanctionsMatched = true
	got := routing.DefaultPolicy().Decide(in, "jev-test", nil)
	if got.Route != domain.RouteDecline {
		t.Fatalf("route = %s, want %s", got.Route, domain.RouteDecline)
	}
}

func TestLowConfidenceGoesToReview(t *testing.T) {
	in := baseRequest()
	answers := map[string]domain.Answer{
		"risk_band":              {Choice: "LOW", Confidence: 0.61},
		"recommended_handling":   {Choice: "ALLOW", Confidence: 0.92},
		"account_takeover_suspected": {},
	}
	got := routing.DefaultPolicy().Decide(in, "jev-test", answers)
	if got.Route != domain.RouteReview {
		t.Fatalf("route = %s, want %s", got.Route, domain.RouteReview)
	}
}

func TestAccountTakeoverGetsStepUp(t *testing.T) {
	in := baseRequest()
	answers := map[string]domain.Answer{
		"risk_band":              {Choice: "HIGH", Confidence: 0.95},
		"recommended_handling":   {Choice: "CHALLENGE", Confidence: 0.91},
		"account_takeover_suspected": {Noul: ptr(0.83)},
	}
	got := routing.DefaultPolicy().Decide(in, "jev-test", answers)
	if got.Route != domain.RouteStepUp {
		t.Fatalf("route = %s, want %s", got.Route, domain.RouteStepUp)
	}
}

func TestHighValueHighRiskRequiresReview(t *testing.T) {
	in := baseRequest()
	in.Transaction.AmountINR = 48500
	answers := map[string]domain.Answer{
		"risk_band":              {Choice: "HIGH", Confidence: 0.95},
		"recommended_handling":   {Choice: "DECLINE", Confidence: 0.95},
		"account_takeover_suspected": {Noul: ptr(0.12)},
	}
	got := routing.DefaultPolicy().Decide(in, "jev-test", answers)
	if got.Route != domain.RouteReview {
		t.Fatalf("route = %s, want %s", got.Route, domain.RouteReview)
	}
}
