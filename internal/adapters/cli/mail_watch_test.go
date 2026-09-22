package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/m365/internal/adapters/mailclassifier"
	"github.com/masonhuemmer/m365/internal/config"
	"github.com/masonhuemmer/m365/internal/domain"
)

func TestMailWatchBaselineResumeAndIncludeExisting(t *testing.T) {
	t.Run("quiet baseline and resume", func(t *testing.T) {
		d, out, errw := testDeps()
		loginAll(t, &d)
		out.Reset()

		if code := Run([]string{"m365", "mail", "watch"}, d); code != domain.ExitOK {
			t.Fatalf("first watch exit %d: %s", code, errw.String())
		}
		if out.Len() != 0 {
			t.Fatalf("baseline output = %q", out.String())
		}
		if code := Run([]string{"m365", "mail", "watch"}, d); code != domain.ExitOK {
			t.Fatalf("second watch exit %d: %s", code, errw.String())
		}
		if out.Len() != 0 {
			t.Fatalf("resume output = %q", out.String())
		}
	})

	t.Run("include existing emits one body-free event per conversation", func(t *testing.T) {
		d, out, errw := testDeps()
		loginAll(t, &d)
		out.Reset()

		if code := Run([]string{"m365", "mail", "watch", "--include-existing"}, d); code != domain.ExitOK {
			t.Fatalf("watch exit %d: %s", code, errw.String())
		}
		lines := nonEmptyLines(out.String())
		if len(lines) != 1 {
			t.Fatalf("watch lines = %q", out.String())
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
			t.Fatal(err)
		}
		if event["event"] != domain.MailChangedEvent || event["message_id"] != "msg-2" {
			t.Fatalf("watch event = %+v", event)
		}
		for _, forbidden := range []string{"body", "attachments", "cursor", "token", "revision"} {
			if _, ok := event[forbidden]; ok {
				t.Fatalf("watch event exposes %s: %+v", forbidden, event)
			}
		}

		out.Reset()
		if code := Run([]string{"m365", "mail", "watch"}, d); code != domain.ExitOK {
			t.Fatalf("resume exit %d: %s", code, errw.String())
		}
		if out.Len() != 0 {
			t.Fatalf("resume output = %q", out.String())
		}
	})
}

func TestMailWatchOutputFailureDoesNotAdvanceCheckpoint(t *testing.T) {
	d, _, errw := testDeps()
	loginAll(t, &d)
	d.Stdout = failingWriter{}
	if code := Run([]string{"m365", "mail", "watch", "--include-existing"}, d); code == domain.ExitOK {
		t.Fatal("expected output failure")
	}

	out := &bytes.Buffer{}
	d.Stdout = out
	errw.Reset()
	if code := Run([]string{"m365", "mail", "watch", "--include-existing"}, d); code != domain.ExitOK {
		t.Fatalf("retry exit %d: %s", code, errw.String())
	}
	if len(nonEmptyLines(out.String())) != 1 {
		t.Fatalf("retry output = %q", out.String())
	}
}

func TestMCPRunMailWatchMatchesDirectJSONLines(t *testing.T) {
	direct, out, errw := testDeps()
	loginAll(t, &direct)
	out.Reset()
	if code := Run([]string{"m365", "mail", "watch", "--include-existing"}, direct); code != domain.ExitOK {
		t.Fatalf("direct exit %d: %s", code, errw.String())
	}
	want := strings.TrimSpace(out.String())

	throughMCP, _, _ := testDeps()
	loginAll(t, &throughMCP)
	client := connectMCP(t, throughMCP)
	result, err := client.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run",
		Arguments: runIn{
			Namespace: "mail",
			Verb:      "watch",
			Flags:     map[string]any{"include-existing": true},
		},
	})
	if err != nil || result.IsError {
		t.Fatal(err, toolText(t, result))
	}
	if got := toolText(t, result); got != want {
		t.Fatalf("MCP watch = %q, direct = %q", got, want)
	}
}

func TestMailWatchRejectsAllFolder(t *testing.T) {
	d, _, errw := testDeps()
	loginAll(t, &d)
	errw.Reset()
	if code := Run([]string{"m365", "mail", "watch", "--folder", "all"}, d); code != domain.ExitUsage {
		t.Fatalf("exit %d: %s", code, errw.String())
	}
}

func TestMailWatchHelpDoesNotRequireSession(t *testing.T) {
	d, out, errw := testDeps()
	if code := Run([]string{"m365", "mail", "watch", "--help"}, d); code != domain.ExitOK {
		t.Fatalf("exit %d: %s", code, errw.String())
	}
	for _, expected := range []string{"--folder", "--include-existing", "--classify", "--target-address", "disabled by default", "mail.changed", "JSON"} {
		if !strings.Contains(out.String(), expected) {
			t.Fatalf("help missing %q: %s", expected, out.String())
		}
	}
}

func TestMailWatchClassificationIsDisabledWithoutExplicitConfig(t *testing.T) {
	d, out, errw := testDeps()
	loginAll(t, &d)
	out.Reset()

	code := Run([]string{"m365", "mail", "watch", "--classify", "--include-existing", "--target-address", "user@example.com"}, d)
	if code != domain.ExitUsage || !strings.Contains(errw.String(), "disabled") || out.Len() != 0 {
		t.Fatalf("disabled classification exit=%d stdout=%q stderr=%q", code, out.String(), errw.String())
	}

	out.Reset()
	errw.Reset()
	if code := Run([]string{"m365", "mail", "watch", "--include-existing"}, d); code != domain.ExitOK {
		t.Fatalf("classifier-free retry exit %d: %s", code, errw.String())
	}
	if len(nonEmptyLines(out.String())) != 1 {
		t.Fatalf("disabled classification advanced checkpoint: %q", out.String())
	}
}

func TestMailWatchMissingJevKeyFailsBeforeCheckpoint(t *testing.T) {
	d, out, errw := testDeps()
	d.Config.Experimental.MailResponseClassification = config.MailResponseClassification{
		Enabled: true, Provider: "jev", Model: "jev-latest", ActionableThreshold: 0.8,
	}
	d.MailClassifierError = domain.Usage("TYPESAFE_API_KEY is required when experimental mail response classification is enabled")
	loginAll(t, &d)
	out.Reset()

	args := []string{"m365", "mail", "watch", "--classify", "--include-existing", "--target-address", "user@example.com"}
	if code := Run(args, d); code != domain.ExitUsage || out.Len() != 0 || !strings.Contains(errw.String(), "TYPESAFE_API_KEY") {
		t.Fatalf("missing key exit=%d stdout=%q stderr=%q", code, out.String(), errw.String())
	}

	d.MailClassifierError = nil
	errw.Reset()
	if code := Run(args, d); code != domain.ExitOK {
		t.Fatalf("retry exit=%d stderr=%q", code, errw.String())
	}
	if len(nonEmptyLines(out.String())) != 1 {
		t.Fatalf("missing key advanced checkpoint: %q", out.String())
	}
}

func TestMailWatchClassifierFailureRedactsCredentials(t *testing.T) {
	d, out, errw := testDeps()
	d.Config.Experimental.MailResponseClassification = config.MailResponseClassification{
		Enabled: true, Provider: "jev", Model: "jev-latest", ActionableThreshold: 0.8,
	}
	d.MailClassifier = &mailclassifier.Fake{Err: domain.Service("TYPESAFE_API_KEY=top-secret Authorization: Bearer bearer-secret")}
	loginAll(t, &d)
	out.Reset()

	args := []string{"m365", "mail", "watch", "--classify", "--include-existing", "--target-address", "user@example.com"}
	if code := Run(args, d); code != domain.ExitService || out.Len() != 0 {
		t.Fatalf("classifier failure exit=%d stdout=%q stderr=%q", code, out.String(), errw.String())
	}
	for _, secret := range []string{"top-secret", "bearer-secret"} {
		if strings.Contains(errw.String(), secret) {
			t.Fatalf("credential %q leaked: %s", secret, errw.String())
		}
	}
}

func TestMailWatchClassificationShapeAndSensitiveFieldAbsence(t *testing.T) {
	d, out, errw := testDeps()
	d.Config.Experimental.MailResponseClassification = config.MailResponseClassification{
		Enabled: true, Provider: "jev", Model: "jev-latest", ActionableThreshold: 0.8,
	}
	loginAll(t, &d)
	out.Reset()

	args := []string{
		"m365", "mail", "watch", "--classify", "--include-existing",
		"--target-address", " USER@example.com ", "--target-address", "alias@example.com",
		"--target-name", "Mason Huemmer",
	}
	if code := Run(args, d); code != domain.ExitOK {
		t.Fatalf("classified watch exit %d: %s", code, errw.String())
	}
	lines := nonEmptyLines(out.String())
	if len(lines) != 1 {
		t.Fatalf("classified lines = %q", out.String())
	}
	var event map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatal(err)
	}
	if event["event"] != domain.MailResponseClassifiedEvent || event["actionable"] != true {
		t.Fatalf("classified event = %+v", event)
	}
	target, _ := event["target"].(map[string]any)
	addresses, _ := target["addresses"].([]any)
	if len(addresses) != 2 || addresses[0] != "user@example.com" || addresses[1] != "alias@example.com" {
		t.Fatalf("classified target = %+v", target)
	}
	classification, _ := event["classification"].(map[string]any)
	probabilities, _ := classification["probabilities"].(map[string]any)
	if len(probabilities) != 4 || classification["status"] != string(domain.ResponseWaitingOnTarget) || classification["model"] == "" {
		t.Fatalf("classification = %+v", classification)
	}
	for _, forbidden := range []string{"body", "attachments", "cursor", "token", "revision"} {
		if containsJSONKey(event, forbidden) {
			t.Fatalf("classified event exposes %s: %+v", forbidden, event)
		}
	}
}

func TestMCPRunClassifiedMailWatchMatchesDirectJSONLines(t *testing.T) {
	direct, out, errw := testDeps()
	direct.Config.Experimental.MailResponseClassification = config.MailResponseClassification{
		Enabled: true, Provider: "jev", Model: "jev-latest", ActionableThreshold: 0.8,
	}
	loginAll(t, &direct)
	out.Reset()
	directArgs := []string{"m365", "mail", "watch", "--classify", "--include-existing", "--target-address", "user@example.com", "--target-name", "Mason"}
	if code := Run(directArgs, direct); code != domain.ExitOK {
		t.Fatalf("direct classified exit %d: %s", code, errw.String())
	}
	want := strings.TrimSpace(out.String())

	throughMCP, _, _ := testDeps()
	throughMCP.Config.Experimental.MailResponseClassification = direct.Config.Experimental.MailResponseClassification
	loginAll(t, &throughMCP)
	client := connectMCP(t, throughMCP)
	result, err := client.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run",
		Arguments: runIn{
			Namespace: "mail",
			Verb:      "watch",
			Flags: map[string]any{
				"classify":         true,
				"include-existing": true,
				"target-address":   []string{"user@example.com"},
				"target-name":      []string{"Mason"},
			},
		},
	})
	if err != nil || result.IsError {
		t.Fatal(err, toolText(t, result))
	}
	if got := toolText(t, result); got != want {
		t.Fatalf("MCP classified watch = %q, direct = %q", got, want)
	}
}

func containsJSONKey(value any, key string) bool {
	switch typed := value.(type) {
	case map[string]any:
		for current, child := range typed {
			if current == key || containsJSONKey(child, key) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsJSONKey(child, key) {
				return true
			}
		}
	}
	return false
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func nonEmptyLines(output string) []string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
