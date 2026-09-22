package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/adapters/keychain"
	"github.com/masonhuemmer/m365/internal/adapters/mailwatchstate"
	"github.com/masonhuemmer/m365/internal/config"
	"github.com/masonhuemmer/m365/internal/domain"
)

func testDeps() (Deps, *bytes.Buffer, *bytes.Buffer) {
	mem := graph.Seed()
	out, errw := &bytes.Buffer{}, &bytes.Buffer{}
	st := &keychain.Fake{}
	d := Deps{
		Config:         config.Config{ClientID: "x", TenantID: "y"},
		Store:          st,
		Mail:           graph.MailAPI{Memory: mem},
		MailChanges:    graph.MailDeltaAPI{Memory: mem},
		MailWatchState: &mailwatchstate.Memory{},
		Teams:          graph.TeamsAPI{Memory: mem},
		Calendar:       graph.CalendarAPI{Memory: mem},
		Files:          graph.FilesAPI{Memory: mem},
		Login:          graph.FakeLogin(true, true),
		Stdout:         out,
		Stderr:         errw,
	}
	return d, out, errw
}

func TestHelpNoSession(t *testing.T) {
	d, out, _ := testDeps()
	code := Run([]string{"m365", "--help"}, d)
	if code != 0 {
		t.Fatal(code)
	}
	s := out.String()
	for _, want := range []string{"auth", "mail", "teams", "JSON", "3", "4", "5", "6"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
}

func TestJSONHumanMutex(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"m365", "--json", "--human", "auth", "status"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw)
	}
}
