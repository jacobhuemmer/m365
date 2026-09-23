package jev

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/masonhuemmer/m365/internal/domain"
)

const (
	defaultEndpoint    = "https://api.typesafe.ai/v1/systemone"
	defaultTimeout     = 15 * time.Second
	maxResponseBytes   = 1 << 20
	responseQuestionID = "response_owner"
)

type Client struct {
	apiKey     string
	model      string
	endpoint   string
	httpClient *http.Client
}

func NewClient(apiKey, model string) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, domain.Usage("TYPESAFE_API_KEY (or TYPESAFE_AI_TOKEN) is required when experimental mail response classification is enabled")
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, domain.Usage("mail response classification model is required")
	}
	return &Client{
		apiKey:     apiKey,
		model:      model,
		endpoint:   defaultEndpoint,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}, nil
}

func (c *Client) Classify(ctx context.Context, input domain.ClassificationInput) (domain.ResponseClassification, error) {
	if c == nil || c.apiKey == "" || c.model == "" || c.endpoint == "" {
		return domain.ResponseClassification{}, domain.Usage("mail response classification is unavailable")
	}
	payload, err := json.Marshal(systemOneRequest{
		State: input,
		Model: c.model,
		Questions: map[string]choiceQuestion{
			responseQuestionID: responseOwnerQuestion(),
		},
	})
	if err != nil {
		return domain.ResponseClassification{}, domain.Service("mail response classification request failed")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return domain.ResponseClassification{}, domain.Service("mail response classification request failed")
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	response, err := httpClient.Do(req)
	if err != nil {
		if isTimeout(err) {
			return domain.ResponseClassification{}, domain.Service("mail response classification request timed out")
		}
		return domain.ResponseClassification{}, domain.Service("mail response classification service error")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return domain.ResponseClassification{}, domain.Service("mail response classification service error")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil || len(body) > maxResponseBytes {
		return domain.ResponseClassification{}, domain.Service("invalid mail response classification response")
	}
	var decoded systemOneResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return domain.ResponseClassification{}, domain.Service("invalid mail response classification response")
	}
	return mapResponse(decoded)
}

type systemOneRequest struct {
	State     domain.ClassificationInput `json:"state"`
	Model     string                     `json:"model"`
	Questions map[string]choiceQuestion  `json:"questions"`
}

type choiceQuestion struct {
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria"`
}

func responseOwnerQuestion() choiceQuestion {
	return choiceQuestion{
		Type:         "choice",
		Instructions: "Who, if anyone, is expected to provide the next substantive email response?",
		Criteria: map[string]string{
			string(domain.ResponseWaitingOnTarget):    "The specified target person is expected to respond.",
			string(domain.ResponseWaitingOnOther):     "Someone other than the target is expected to respond.",
			string(domain.ResponseNoResponseExpected): "The exchange is complete or informational.",
			string(domain.ResponseUnclear):            "The available thread does not support a reliable decision.",
		},
	}
}

type systemOneResponse struct {
	Model   string                  `json:"model"`
	Answers map[string]choiceAnswer `json:"answers"`
}

type choiceAnswer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    *float64           `json:"confidence"`
}

func mapResponse(response systemOneResponse) (domain.ResponseClassification, error) {
	model := strings.TrimSpace(response.Model)
	answer, ok := response.Answers[responseQuestionID]
	if model == "" || !ok || answer.Type != "choice" {
		return domain.ResponseClassification{}, domain.Service("invalid mail response classification response")
	}
	status, ok := responseStatus(answer.Choice)
	if !ok || len(answer.Probabilities) != 4 {
		return domain.ResponseClassification{}, domain.Service("invalid mail response classification response")
	}
	probabilities := make(map[domain.ResponseStatus]float64, 4)
	for rawStatus, probability := range answer.Probabilities {
		mappedStatus, known := responseStatus(rawStatus)
		if !known || probability < 0 || probability > 1 {
			return domain.ResponseClassification{}, domain.Service("invalid mail response classification response")
		}
		probabilities[mappedStatus] = probability
	}
	for _, required := range responseStatuses() {
		if _, present := probabilities[required]; !present {
			return domain.ResponseClassification{}, domain.Service("invalid mail response classification response")
		}
	}
	if answer.Confidence != nil && (*answer.Confidence < 0 || *answer.Confidence > 1) {
		return domain.ResponseClassification{}, domain.Service("invalid mail response classification response")
	}
	return domain.ResponseClassification{
		Status:            status,
		TargetProbability: probabilities[domain.ResponseWaitingOnTarget],
		Probabilities:     probabilities,
		Confidence:        answer.Confidence,
		Model:             model,
	}, nil
}

func responseStatus(value string) (domain.ResponseStatus, bool) {
	status := domain.ResponseStatus(value)
	for _, known := range responseStatuses() {
		if status == known {
			return status, true
		}
	}
	return "", false
}

func responseStatuses() []domain.ResponseStatus {
	return []domain.ResponseStatus{
		domain.ResponseWaitingOnTarget,
		domain.ResponseWaitingOnOther,
		domain.ResponseNoResponseExpected,
		domain.ResponseUnclear,
	}
}

func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netError net.Error
	return errors.As(err, &netError) && netError.Timeout()
}
