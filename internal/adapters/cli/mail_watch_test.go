package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

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
	for _, expected := range []string{"--folder", "--include-existing", "mail.changed", "JSON"} {
		if !strings.Contains(out.String(), expected) {
			t.Fatalf("help missing %q: %s", expected, out.String())
		}
	}
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
