# Banking suspicious-transaction routing with Go and Jev

This is a complete project scaffold and worked example for routing suspicious card transactions. It uses the TypeSafe System One HTTP API directly so the application does not depend on a Go SDK.

## Scope and safety boundary

The service does **not** approve, decline, reverse, or block payments. It routes a transaction to one of four operational paths:

- `ALLOW_STANDARD_PROCESSING`
- `STEP_UP_AUTHENTICATION`
- `MANUAL_FRAUD_REVIEW`
- `DECLINE_BY_DETERMINISTIC_POLICY`

The Jev model supplies probabilistic assessments. Deterministic policy owns hard stops, final routing, and safe fallback behavior.

## Architecture

```text
HTTP request / transaction event
        |
        v
Input validation and normalization
        |
        +--> deterministic hard-stop rules ----> final route
        |
        v
Jev evaluator (typed questions in one request)
        |
        v
Policy orchestrator (thresholds, value limits, confidence)
        |
        v
Audit record + response / event
```

## Project tree

```text
jev-banking-routing-go/
├── cmd/api/main.go
├── internal/domain/transaction.go
├── internal/jev/client.go
├── internal/routing/policy.go
├── internal/service/router.go
├── internal/service/router_test.go
├── testdata/scenarios.json
├── .env.example
├── Dockerfile
├── Makefile
├── README.md
└── go.mod
```

## 1. Create the module

```bash
mkdir jev-banking-routing-go
cd jev-banking-routing-go
go mod init github.com/example/jev-banking-routing-go
mkdir -p cmd/api internal/domain internal/jev internal/routing internal/service testdata
```

Create the following files.

---

## `go.mod`

```go
module github.com/example/jev-banking-routing-go

go 1.23
```

No third-party packages are required. The standard library is enough for a small, inspectable starting point.

---

## `internal/domain/transaction.go`

```go
package domain

import "time"

type Transaction struct {
	TransactionID      string  `json:"transaction_id"`
	AmountINR          float64 `json:"amount_inr"`
	Currency           string  `json:"currency"`
	MerchantName       string  `json:"merchant_name"`
	MerchantCategory   string  `json:"merchant_category"`
	MerchantCountry    string  `json:"merchant_country"`
	Channel            string  `json:"channel"`
	OccurredAt         time.Time `json:"occurred_at"`
	CardPresent        bool    `json:"card_present"`
	SanctionsMatched   bool    `json:"sanctions_matched"`
	AvailableBalanceINR float64 `json:"available_balance_inr"`
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
	PasswordResetWithin24Hours bool  `json:"password_reset_within_24_hours"`
}

type RiskSignals struct {
	TransactionsLast10Minutes int  `json:"transactions_last_10_minutes"`
	MerchantFirstSeen         bool `json:"merchant_first_seen_for_customer"`
	AddressVerification       string `json:"address_verification"`
	BeneficiaryNovelty        bool `json:"beneficiary_novelty"`
}

type EvaluationRequest struct {
	Transaction     Transaction     `json:"transaction"`
	CustomerHistory CustomerHistory `json:"customer_history"`
	Session         Session         `json:"session"`
	RiskSignals     RiskSignals     `json:"risk_signals"`
}

type Route string

const (
	RouteAllow  Route = "ALLOW_STANDARD_PROCESSING"
	RouteStepUp Route = "STEP_UP_AUTHENTICATION"
	RouteReview Route = "MANUAL_FRAUD_REVIEW"
	RouteDecline Route = "DECLINE_BY_DETERMINISTIC_POLICY"
)

type RoutingDecision struct {
	TransactionID string             `json:"transaction_id"`
	Route         Route              `json:"route"`
	Reasons       []string           `json:"reasons"`
	Model         string             `json:"model,omitempty"`
	Assessments   map[string]Answer  `json:"assessments,omitempty"`
	FallbackUsed  bool               `json:"fallback_used"`
}

type Answer struct {
	Choice        string             `json:"choice,omitempty"`
	Confidence    float64            `json:"confidence,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Noul          *float64           `json:"noul,omitempty"`
}
```

---

## 2. Model the Jev HTTP contract

The TypeSafe API accepts `state`, `model`, and a map of typed `questions`; Choice answers include the selected option, a full probability distribution, and confidence. Noul answers return a probability from 0 to 1. [web:12]

## `internal/jev/client.go`

```go
package jev

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/example/jev-banking-routing-go/internal/domain"
)

const defaultBaseURL = "https://api.typesafe.ai/v1/systemone"

type Client struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

type question struct {
	Type         string      `json:"type"`
	Instructions interface{} `json:"instructions"`
	Criteria     interface{} `json:"criteria,omitempty"`
}

type requestBody struct {
	State     domain.EvaluationRequest `json:"state"`
	Model     string                   `json:"model"`
	Questions map[string]question      `json:"questions"`
}

type rawAnswer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice"`
	Noul          *float64           `json:"noul"`
	Confidence    float64            `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities"`
}

type responseBody struct {
	Model   string               `json:"model"`
	Answers map[string]rawAnswer `json:"answers"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func New(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: defaultBaseURL,
		Model:   "jev-latest",
		HTTPClient: &http.Client{Timeout: 3 * time.Second},
	}
}

func (c *Client) Evaluate(ctx context.Context, in domain.EvaluationRequest) (string, map[string]domain.Answer, error) {
	if c.APIKey == "" {
		return "", nil, errors.New("TYPESAFE_API_KEY is required")
	}

	body := requestBody{
		State: in,
		Model: c.Model,
		Questions: map[string]question{
			"risk_band": {
				Type: "choice",
				Instructions: "Classify the fraud risk of this card transaction using the transaction, customer history, session, and risk signals. This is a risk assessment only; do not make an authorization decision.",
				Criteria: map[string]string{
					"LOW": "Signals align with the customer's usual behavior and there is little indication of fraud.",
					"MEDIUM": "There are meaningful anomalies, but the evidence is incomplete or mixed.",
					"HIGH": "Multiple strong indicators suggest likely fraud or account takeover.",
				},
			},
			"recommended_handling": {
				Type: "choice",
				Instructions: "Choose the safest fraud-operations handling route. This is a recommendation only; deterministic policy will make the final route.",
				Criteria: map[string]string{
					"ALLOW": "No additional fraud friction appears warranted from the supplied context.",
					"CHALLENGE": "Ask the customer for step-up authentication before allowing the transaction to continue.",
					"REVIEW": "Send the transaction to a fraud analyst because the case is suspicious, high impact, or ambiguous.",
					"DECLINE": "The supplied signals strongly suggest the transaction should not proceed; final decline remains a policy decision.",
				},
			},
			"account_takeover_suspected": {
				Type: "noul",
				Instructions: "Is account takeover plausibly indicated by the supplied session and transaction context?",
				Criteria: map[string]string{
					"true": "Recent credential changes, unrecognized device, anomalous behavior, or similar evidence suggests account takeover.",
					"false": "There is no meaningful account-takeover indication in the supplied state.",
				},
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", nil, fmt.Errorf("marshal Jev request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(payload))
	if err != nil {
		return "", nil, fmt.Errorf("build Jev request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("call Jev: %w", err)
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", nil, fmt.Errorf("read Jev response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("Jev returned %s: %s", resp.Status, string(responseBytes))
	}

	var decoded responseBody
	if err := json.Unmarshal(responseBytes, &decoded); err != nil {
		return "", nil, fmt.Errorf("decode Jev response: %w", err)
	}

	answers := make(map[string]domain.Answer, len(decoded.Answers))
	for id, answer := range decoded.Answers {
		answers[id] = domain.Answer{
			Choice:        answer.Choice,
			Confidence:    answer.Confidence,
			Probabilities: answer.Probabilities,
			Noul:          answer.Noul,
		}
	}
	return decoded.Model, answers, nil
}
```

---

## 3. Add deterministic policy

This is the most important file. It ensures the model cannot override sanctions, insufficient balance, invalid input, or a deliberately conservative high-risk policy.

## `internal/routing/policy.go`

```go
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
```

---

## 4. Implement the service and fallback

## `internal/service/router.go`

```go
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
```

---

## 5. Expose a small HTTP API

## `cmd/api/main.go`

```go
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/example/jev-banking-routing-go/internal/domain"
	"github.com/example/jev-banking-routing-go/internal/jev"
	"github.com/example/jev-banking-routing-go/internal/routing"
	"github.com/example/jev-banking-routing-go/internal/service"
)

func main() {
	client := jev.New(os.Getenv("TYPESAFE_API_KEY"))
	router := service.Router{
		Jev:    client,
		Policy: routing.DefaultPolicy(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("POST /v1/transactions/route", func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		var in domain.EvaluationRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&in); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(req.Context(), 2500*time.Millisecond)
		defer cancel()
		out := router.Route(ctx, in)

		w.Header().Set("Content-Type", "application/json")
		if out.FallbackUsed {
			w.Header().Set("X-Routing-Fallback", "true")
		}
		_ = json.NewEncoder(w).Encode(out)
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("routing API listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
```

---

## 6. Worked input scenario

This scenario deliberately combines several suspicious indicators: a large transaction relative to customer history, a new device, a recent password reset, a new merchant, and high transaction velocity.

## `testdata/scenarios.json`

```json
{
  "transaction": {
    "transaction_id": "txn_01JEV_DEMO_001",
    "amount_inr": 48500,
    "currency": "INR",
    "merchant_name": "Example Electronics",
    "merchant_category": "ELECTRONICS",
    "merchant_country": "IN",
    "channel": "ECOMMERCE",
    "occurred_at": "2026-09-24T18:15:00Z",
    "card_present": false,
    "sanctions_matched": false,
    "available_balance_inr": 87500
  },
  "customer_history": {
    "customer_id": "cust_2032",
    "account_age_days": 1820,
    "average_transaction_amount_inr": 2100,
    "largest_transaction_last_90_days_inr": 16000,
    "chargebacks_last_12_months": 0
  },
  "session": {
    "device_recognized": false,
    "ip_country": "IN",
    "customer_home_country": "IN",
    "password_reset_within_24_hours": true
  },
  "risk_signals": {
    "transactions_last_10_minutes": 5,
    "merchant_first_seen_for_customer": true,
    "address_verification": "MATCH",
    "beneficiary_novelty": false
  }
}
```

Start the API:

```bash
export TYPESAFE_API_KEY="your-api-key"
go run ./cmd/api
```

Call it from a second terminal:

```bash
curl -sS http://localhost:8080/v1/transactions/route \
  -H 'Content-Type: application/json' \
  --data-binary @testdata/scenarios.json | jq
```

Expected **shape**, not a guaranteed model result:

```json
{
  "transaction_id": "txn_01JEV_DEMO_001",
  "route": "STEP_UP_AUTHENTICATION",
  "reasons": [
    "account-takeover probability exceeds step-up threshold"
  ],
  "model": "jev-...",
  "assessments": {
    "risk_band": {
      "choice": "HIGH",
      "confidence": 0.91,
      "probabilities": {
        "LOW": 0.01,
        "MEDIUM": 0.08,
        "HIGH": 0.91
      }
    },
    "recommended_handling": {
      "choice": "CHALLENGE",
      "confidence": 0.86,
      "probabilities": {
        "ALLOW": 0.01,
        "CHALLENGE": 0.86,
        "REVIEW": 0.11,
        "DECLINE": 0.02
      }
    },
    "account_takeover_suspected": {
      "noul": 0.82
    }
  },
  "fallback_used": false
}
```

The expected operational action is step-up authentication, not an immediate decline. Even a high-risk assessment is insufficient by itself to decline a ₹48,500 transaction under this policy because the configured auto-decline ceiling is ₹10,000.

---

## 7. Test the policy without calling Jev

The policy is deterministic and should be unit-tested independently of external inference.

## `internal/service/router_test.go`

```go
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
			TransactionID: "txn-test",
			AmountINR: 5000,
			Currency: "INR",
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
		"risk_band": {Choice: "LOW", Confidence: 0.61},
		"recommended_handling": {Choice: "ALLOW", Confidence: 0.92},
	}
	got := routing.DefaultPolicy().Decide(in, "jev-test", answers)
	if got.Route != domain.RouteReview {
		t.Fatalf("route = %s, want %s", got.Route, domain.RouteReview)
	}
}

func TestAccountTakeoverGetsStepUp(t *testing.T) {
	in := baseRequest()
	answers := map[string]domain.Answer{
		"risk_band": {Choice: "HIGH", Confidence: 0.95},
		"recommended_handling": {Choice: "CHALLENGE", Confidence: 0.91},
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
		"risk_band": {Choice: "HIGH", Confidence: 0.95},
		"recommended_handling": {Choice: "DECLINE", Confidence: 0.95},
		"account_takeover_suspected": {Noul: ptr(0.12)},
	}
	got := routing.DefaultPolicy().Decide(in, "jev-test", answers)
	if got.Route != domain.RouteReview {
		t.Fatalf("route = %s, want %s", got.Route, domain.RouteReview)
	}
}
```

Run tests:

```bash
go test ./...
```

---

## 8. Evaluation plan: prove the routing policy works

Do not evaluate the system only by asking whether Jev agrees with an analyst. Evaluate the end-to-end route and the eventual outcome.

Create an offline, de-identified dataset with fields such as:

```text
transaction_id
input_state_json
historical_final_outcome
historical_analyst_route
confirmed_fraud
chargeback_within_90_days
customer_abandonment_after_challenge
manual_review_minutes
```

For each historical case:

1. Run the same versioned Jev questions against the recorded state.
2. Apply the exact policy version used in production.
3. Compare the proposed route with the historical resolution and eventual fraud outcome.
4. Slice results by amount band, channel, merchant category, account age, device recognition, and geography.
5. Review false positives and false negatives with fraud operations before enabling any automation.

Recommended metrics:

| Metric | Why it matters |
|---|---|
| Fraud capture / recall | How many confirmed fraud cases were challenged, reviewed, or declined |
| False-positive challenge rate | How often good customers receive unnecessary friction |
| Precision of `MANUAL_FRAUD_REVIEW` | Whether review capacity is spent on genuinely suspicious cases |
| Step-up completion rate | Whether challenged legitimate customers can complete authentication |
| Net fraud loss | Financial outcome after fraud, recoveries, and operational cost |
| Manual-review rate | Capacity and cost impact |
| p95 inference plus routing latency | Whether the system is safe for the transaction path |
| Fallback rate | Reliability of the integration |
| Confidence calibration | Whether 0.9-confidence cases really produce the expected outcome rate |

### Calibration example

Bucket decisions by Jev confidence:

| Confidence bucket | Number of cases | Confirmed fraud rate | Interpretation |
|---|---:|---:|---|
| 0.50–0.69 | 2,000 | 8% | Keep in manual review or challenge path |
| 0.70–0.84 | 4,000 | 22% | Conservative step-up may be appropriate |
| 0.85–1.00 | 3,000 | 64% | Candidate for high-priority review; auto-decline needs separate policy approval |

Do not assume confidence is perfectly calibrated. Measure it against delayed labels such as confirmed fraud or chargeback outcomes, and re-evaluate after input-distribution changes.

---

## 9. Operational hardening before production

- Persist an audit record with a correlation ID, normalized input-feature version, policy version, model ID, questions version, full answer probabilities, selected route, and final downstream outcome.
- Never log PAN, CVV, raw authentication secrets, or unnecessary personal data. Use tokenized identifiers and data-minimization controls.
- Add retries only for transient `429` and `529` errors, with exponential backoff and a strict overall deadline. TypeSafe documents these as rate-limit and temporary-overload conditions. [web:12]
- Keep the fail-safe default as `MANUAL_FRAUD_REVIEW` or an existing deterministic fraud path—not “allow.”
- Run shadow mode before customer-impacting routing: write decisions to an audit topic but retain the incumbent route.
- Require approval and change control for schema, question, model, and threshold changes. A rubric change is a risk-policy change.
- Separate model evaluation from payment authorization. This example routes suspicious transactions; it does not execute payment actions.
- Add circuit breaking, request-size limits, per-tenant rate limiting, and metrics for timeout, error, fallback, and model-score distributions.

---

## `Dockerfile`

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/routing-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/routing-api /routing-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/routing-api"]
```

## `Makefile`

```makefile
.PHONY: test run build docker

test:
	go test ./...

run:
	go run ./cmd/api

build:
	go build ./cmd/api

docker:
	docker build -t jev-banking-routing:local .
```

## `.env.example`

```bash
TYPESAFE_API_KEY=replace-with-a-real-key
```

## Next implementation increments

1. Replace direct HTTP response exposure with a Kafka or SQS event consumer and a routed-decision output topic.
2. Add PostgreSQL audit persistence with a data-retention policy and a separate restricted-access evidence store.
3. Implement an offline evaluator that replays de-identified historical transactions and writes CSV metrics by decision slice.
4. Add OpenTelemetry traces and Prometheus metrics around Jev calls, policy decisions, and fallback events.
5. Add an analyst-feedback event that records the final disposition and creates delayed labels for calibration monitoring.

## Source notes

Jev is designed to evaluate state against typed questions and return structured, probabilistic answers. The native endpoint is `POST https://api.typesafe.ai/v1/systemone`; a request contains `state`, `model`, and `questions`. Choice provides a chosen option, a probability distribution, and confidence, while Noul provides a calibrated yes/no probability. [web:12]
