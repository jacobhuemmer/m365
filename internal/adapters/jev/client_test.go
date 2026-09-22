package jev

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestClientSendsChoiceRequestAndMapsResponse(t *testing.T) {
	input := domain.ClassificationInput{
		Target: domain.ResponseTarget{
			Names:     []string{"Mason Huemmer"},
			Addresses: []string{"mason@example.com"},
		},
		Thread: []domain.ClassificationMessage{{
			From:     "alex@example.com",
			To:       []string{"mason@example.com"},
			CC:       []string{"ops@example.com"},
			Received: "2026-09-21T15:04:05Z",
			Subject:  "Approval needed",
			BodyText: "Can you approve this today?",
		}},
		Truncation: domain.ClassificationTruncation{MessageLimit: 10, BodyByteLimit: 65536},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/systemone" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-secret" {
			t.Fatalf("authorization = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content type = %q", got)
		}
		var request struct {
			State     domain.ClassificationInput `json:"state"`
			Model     string                     `json:"model"`
			Questions map[string]struct {
				Type         string            `json:"type"`
				Instructions string            `json:"instructions"`
				Criteria     map[string]string `json:"criteria"`
			} `json:"questions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.Model != "jev-latest" || !reflect.DeepEqual(request.State, input) {
			t.Fatalf("request = %+v", request)
		}
		question, ok := request.Questions["response_owner"]
		if !ok || len(request.Questions) != 1 || question.Type != "choice" ||
			question.Instructions != "Who, if anyone, is expected to provide the next substantive email response?" {
			t.Fatalf("question = %+v", request.Questions)
		}
		wantCriteria := map[string]string{
			"waiting_on_target":    "The specified target person is expected to respond.",
			"waiting_on_other":     "Someone other than the target is expected to respond.",
			"no_response_expected": "The exchange is complete or informational.",
			"unclear":              "The available thread does not support a reliable decision.",
		}
		if !reflect.DeepEqual(question.Criteria, wantCriteria) {
			t.Fatalf("criteria = %+v", question.Criteria)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"model":"jev-1.13.0",
			"answers":{"response_owner":{
				"type":"choice",
				"choice":"waiting_on_target",
				"probabilities":{"waiting_on_target":0.91,"waiting_on_other":0.03,"no_response_expected":0.04,"unclear":0.02},
				"confidence":0.86
			}},
			"usage":{"input_tokens":100,"output_tokens":20}
		}`))
	}))
	defer server.Close()

	client, err := NewClient("test-secret", "jev-latest")
	if err != nil {
		t.Fatal(err)
	}
	client.endpoint = server.URL + "/v1/systemone"
	client.httpClient = server.Client()

	got, err := client.Classify(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	confidence := 0.86
	want := domain.ResponseClassification{
		Status:            domain.ResponseWaitingOnTarget,
		TargetProbability: 0.91,
		Probabilities: map[domain.ResponseStatus]float64{
			domain.ResponseWaitingOnTarget:    0.91,
			domain.ResponseWaitingOnOther:     0.03,
			domain.ResponseNoResponseExpected: 0.04,
			domain.ResponseUnclear:            0.02,
		},
		Confidence: &confidence,
		Model:      "jev-1.13.0",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("classification = %+v, want %+v", got, want)
	}
}

func TestNewClientRequiresKeyAndUsesBoundedTimeout(t *testing.T) {
	client, err := NewClient("", "jev-latest")
	if client != nil || domain.ExitOf(err) != domain.ExitUsage || !strings.Contains(err.Error(), "TYPESAFE_API_KEY") {
		t.Fatalf("missing key client=%v err=%v", client, err)
	}

	client, err = NewClient("secret", "jev-latest")
	if err != nil {
		t.Fatal(err)
	}
	if client.httpClient == nil || client.httpClient.Timeout != 15*time.Second {
		t.Fatalf("HTTP timeout = %v", client.httpClient)
	}
}

func TestClientAcceptsOmittedConfidence(t *testing.T) {
	client, closeServer := testClient(t, http.StatusOK, `{
		"model":"jev-1.13.0",
		"answers":{"response_owner":{
			"type":"choice",
			"choice":"unclear",
			"probabilities":{"waiting_on_target":0.1,"waiting_on_other":0.2,"no_response_expected":0.3,"unclear":0.4}
		}}
	}`)
	defer closeServer()

	got, err := client.Classify(context.Background(), domain.ClassificationInput{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Confidence != nil || got.Status != domain.ResponseUnclear || got.TargetProbability != 0.1 {
		t.Fatalf("classification = %+v", got)
	}
}

func TestClientRejectsMalformedChoiceResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "invalid JSON", body: `{"model":`},
		{name: "missing returned model", body: validResponse(`"model":""`)},
		{name: "wrong answer type", body: validResponse(`"type":"noul"`)},
		{name: "unknown choice", body: validResponse(`"choice":"other"`)},
		{name: "missing probability", body: `{"model":"jev-1.13.0","answers":{"response_owner":{"type":"choice","choice":"unclear","probabilities":{"waiting_on_target":0.1,"waiting_on_other":0.2,"unclear":0.7}}}}`},
		{name: "unknown probability", body: `{"model":"jev-1.13.0","answers":{"response_owner":{"type":"choice","choice":"unclear","probabilities":{"waiting_on_target":0.1,"waiting_on_other":0.2,"no_response_expected":0.2,"unclear":0.4,"other":0.1}}}}`},
		{name: "probability out of range", body: validResponse(`"probabilities":{"waiting_on_target":-0.1,"waiting_on_other":0.2,"no_response_expected":0.3,"unclear":0.6}`)},
		{name: "confidence out of range", body: validResponse(`"confidence":1.1`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, closeServer := testClient(t, http.StatusOK, tt.body)
			defer closeServer()
			if _, err := client.Classify(context.Background(), domain.ClassificationInput{}); domain.ExitOf(err) != domain.ExitService {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestClientMapsHTTPFailuresWithoutRetryOrSecretLeak(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusTooManyRequests, http.StatusInternalServerError, 529} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"error":"test-secret Authorization: Bearer test-secret"}`))
			}))
			defer server.Close()
			client, err := NewClient("test-secret", "jev-latest")
			if err != nil {
				t.Fatal(err)
			}
			client.endpoint = server.URL
			client.httpClient = server.Client()
			_, err = client.Classify(context.Background(), domain.ClassificationInput{})
			if domain.ExitOf(err) != domain.ExitService || calls.Load() != 1 {
				t.Fatalf("status %d calls=%d err=%v", status, calls.Load(), err)
			}
			if strings.Contains(err.Error(), "test-secret") || strings.Contains(strings.ToLower(err.Error()), "bearer") {
				t.Fatalf("secret leaked in %q", err.Error())
			}
		})
	}
}

func TestClientMapsTimeoutToServiceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	client, err := NewClient("test-secret", "jev-latest")
	if err != nil {
		t.Fatal(err)
	}
	client.endpoint = server.URL
	client.httpClient = &http.Client{Timeout: 10 * time.Millisecond}

	_, err = client.Classify(context.Background(), domain.ClassificationInput{})
	if domain.ExitOf(err) != domain.ExitService || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("timeout error = %v", err)
	}
}

func testClient(t *testing.T, status int, body string) (*Client, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	client, err := NewClient("test-secret", "jev-latest")
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	client.endpoint = server.URL
	client.httpClient = server.Client()
	return client, server.Close
}

func validResponse(replacement string) string {
	base := `{"model":"jev-1.13.0","answers":{"response_owner":{"type":"choice","choice":"waiting_on_target","probabilities":{"waiting_on_target":0.7,"waiting_on_other":0.1,"no_response_expected":0.1,"unclear":0.1},"confidence":0.6}}}`
	switch {
	case strings.HasPrefix(replacement, `"model"`):
		return strings.Replace(base, `"model":"jev-1.13.0"`, replacement, 1)
	case strings.HasPrefix(replacement, `"type"`):
		return strings.Replace(base, `"type":"choice"`, replacement, 1)
	case strings.HasPrefix(replacement, `"choice"`):
		return strings.Replace(base, `"choice":"waiting_on_target"`, replacement, 1)
	case strings.HasPrefix(replacement, `"probabilities"`):
		return strings.Replace(base, `"probabilities":{"waiting_on_target":0.7,"waiting_on_other":0.1,"no_response_expected":0.1,"unclear":0.1}`, replacement, 1)
	default:
		return strings.Replace(base, `"confidence":0.6`, replacement, 1)
	}
}
