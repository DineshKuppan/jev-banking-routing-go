package service

import (
	"context"
	"fmt"

	"github.com/example/jev-banking-routing-go/internal/domain"
	"github.com/example/jev-banking-routing-go/internal/jev"
	"github.com/example/jev-banking-routing-go/internal/routing"
)

type Router struct {
	Jev    *jev.Client
	Policy routing.Policy
}

func (r Router) Route(ctx context.Context, in domain.EvaluationRequest) domain.RoutingDecision {
	if err := validate(in); err != nil {
		return domain.RoutingDecision{
			TransactionID: in.Transaction.TransactionID,
			Route:         domain.RouteReview,
			FallbackUsed:  true,
			Reasons:       []string{"invalid input: " + err.Error()},
		}
	}

	model, answers, err := r.Jev.Evaluate(ctx, in)
	if err != nil {
		return domain.RoutingDecision{
			TransactionID: in.Transaction.TransactionID,
			Route:         domain.RouteReview,
			FallbackUsed:  true,
			Reasons:       []string{"Jev unavailable or invalid response: " + err.Error()},
		}
	}
	return r.Policy.Decide(in, model, answers)
}

func validate(in domain.EvaluationRequest) error {
	if in.Transaction.TransactionID == "" {
		return fmt.Errorf("transaction_id is required")
	}
	if in.Transaction.AmountINR <= 0 {
		return fmt.Errorf("amount_inr must be positive")
	}
	if in.Transaction.Currency != "INR" {
		return fmt.Errorf("this demo supports INR only")
	}
	if in.CustomerHistory.CustomerID == "" {
		return fmt.Errorf("customer_id is required")
	}
	return nil
}
