package routing

import (
	"fmt"

	"github.com/example/jev-banking-routing-go/internal/domain"
)

type Policy struct {
	MinimumConfidenceForAutomation float64
	MaximumAutoDeclineAmountINR    float64
	AccountTakeoverThreshold       float64
}

func DefaultPolicy() Policy {
	return Policy{
		MinimumConfidenceForAutomation: 0.80,
		MaximumAutoDeclineAmountINR:    10000,
		AccountTakeoverThreshold:       0.75,
	}
}

func (p Policy) Decide(in domain.EvaluationRequest, model string, a map[string]domain.Answer) domain.RoutingDecision {
	decision := domain.RoutingDecision{
		TransactionID: in.Transaction.TransactionID,
		Model:         model,
		Assessments:   a,
	}

	if in.Transaction.SanctionsMatched {
		decision.Route = domain.RouteDecline
		decision.Reasons = []string{"deterministic hard stop: sanctions match"}
		return decision
	}
	if in.Transaction.AmountINR > in.Transaction.AvailableBalanceINR {
		decision.Route = domain.RouteDecline
		decision.Reasons = []string{"deterministic hard stop: insufficient available balance"}
		return decision
	}

	risk := a["risk_band"]
	handling := a["recommended_handling"]
	ato := a["account_takeover_suspected"]

	if risk.Confidence < p.MinimumConfidenceForAutomation || handling.Confidence < p.MinimumConfidenceForAutomation {
		decision.Route = domain.RouteReview
		decision.Reasons = []string{fmt.Sprintf("model confidence below automation threshold %.2f", p.MinimumConfidenceForAutomation)}
		return decision
	}

	if ato.Noul != nil && *ato.Noul >= p.AccountTakeoverThreshold {
		decision.Route = domain.RouteStepUp
		decision.Reasons = []string{"account-takeover probability exceeds step-up threshold"}
		return decision
	}

	if risk.Choice == "HIGH" {
		if in.Transaction.AmountINR <= p.MaximumAutoDeclineAmountINR && handling.Choice == "DECLINE" {
			decision.Route = domain.RouteDecline
			decision.Reasons = []string{"high model risk with strong decline recommendation within low-value auto-decline limit"}
			return decision
		}
		decision.Route = domain.RouteReview
		decision.Reasons = []string{"high-risk case is above auto-decline limit or needs analyst review"}
		return decision
	}

	if handling.Choice == "CHALLENGE" || risk.Choice == "MEDIUM" {
		decision.Route = domain.RouteStepUp
		decision.Reasons = []string{"medium risk or step-up recommendation"}
		return decision
	}

	decision.Route = domain.RouteAllow
	decision.Reasons = []string{"low-risk, high-confidence assessment and no deterministic hard stop"}
	return decision
}
