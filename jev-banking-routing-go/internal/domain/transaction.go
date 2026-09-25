package domain

import "time"

type Transaction struct {
	TransactionID      string    `json:"transaction_id"`
	AmountINR          float64   `json:"amount_inr"`
	Currency           string    `json:"currency"`
	MerchantName       string    `json:"merchant_name"`
	MerchantCategory   string    `json:"merchant_category"`
	MerchantCountry    string    `json:"merchant_country"`
	Channel            string    `json:"channel"`
	OccurredAt         time.Time `json:"occurred_at"`
	CardPresent        bool      `json:"card_present"`
	SanctionsMatched   bool      `json:"sanctions_matched"`
	AvailableBalanceINR float64   `json:"available_balance_inr"`
}

type CustomerHistory struct {
	CustomerID                     string  `json:"customer_id"`
	AccountAgeDays                 int     `json:"account_age_days"`
	AverageTransactionAmountINR    float64 `json:"average_transaction_amount_inr"`
	LargestTransactionLast90DaysINR float64 `json:"largest_transaction_last_90_days_inr"`
	ChargebacksLast12Months        int     `json:"chargebacks_last_12_months"`
}

type Session struct {
	DeviceRecognized          bool   `json:"device_recognized"`
	IPCountry                 string `json:"ip_country"`
	CustomerHomeCountry       string `json:"customer_home_country"`
	PasswordResetWithin24Hours bool   `json:"password_reset_within_24_hours"`
}

type RiskSignals struct {
	TransactionsLast10Minutes int    `json:"transactions_last_10_minutes"`
	MerchantFirstSeen         bool   `json:"merchant_first_seen_for_customer"`
	AddressVerification       string `json:"address_verification"`
	BeneficiaryNovelty        bool   `json:"beneficiary_novelty"`
}

type EvaluationRequest struct {
	Transaction     Transaction     `json:"transaction"`
	CustomerHistory CustomerHistory `json:"customer_history"`
	Session         Session         `json:"session"`
	RiskSignals     RiskSignals     `json:"risk_signals"`
}

type Route string

const (
	RouteAllow   Route = "ALLOW_STANDARD_PROCESSING"
	RouteStepUp  Route = "STEP_UP_AUTHENTICATION"
	RouteReview  Route = "MANUAL_FRAUD_REVIEW"
	RouteDecline Route = "DECLINE_BY_DETERMINISTIC_POLICY"
)

type RoutingDecision struct {
	TransactionID string            `json:"transaction_id"`
	Route         Route             `json:"route"`
	Reasons       []string          `json:"reasons"`
	Model         string            `json:"model,omitempty"`
	Assessments   map[string]Answer `json:"assessments,omitempty"`
	FallbackUsed  bool              `json:"fallback_used"`
}

type Answer struct {
	Choice        string             `json:"choice,omitempty"`
	Confidence    float64            `json:"confidence,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Noul          *float64           `json:"noul,omitempty"`
}
