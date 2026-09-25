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
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
		Model:      "jev-latest",
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
					"LOW":    "Signals align with the customer's usual behavior and there is little indication of fraud.",
					"MEDIUM": "There are meaningful anomalies, but the evidence is incomplete or mixed.",
					"HIGH":   "Multiple strong indicators suggest likely fraud or account takeover.",
				},
			},
			"recommended_handling": {
				Type: "choice",
				Instructions: "Choose the safest fraud-operations handling route. This is a recommendation only; deterministic policy will make the final route.",
				Criteria: map[string]string{
					"ALLOW":     "No additional fraud friction appears warranted from the supplied context.",
					"CHALLENGE": "Ask the customer for step-up authentication before allowing the transaction to continue.",
					"REVIEW":    "Send the transaction to a fraud analyst because the case is suspicious, high impact, or ambiguous.",
					"DECLINE":   "The supplied signals strongly suggest the transaction should not proceed; final decline remains a policy decision.",
				},
			},
			"account_takeover_suspected": {
				Type: "noul",
				Instructions: "Is account takeover plausibly indicated by the supplied session and transaction context?",
				Criteria: map[string]string{
					"true":  "Recent credential changes, unrecognized device, anomalous behavior, or similar evidence suggests account takeover.",
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
