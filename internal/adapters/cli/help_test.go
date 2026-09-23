package cli

import (
	"strings"
	"testing"
)

func TestHelpCaps(t *testing.T) {
	d, out, _ := testDeps()
	Run([]string{"m365", "mail", "send", "--help"}, d)
	s := out.String()
	if !strings.Contains(s, "10 MiB") || !strings.Contains(s, "10 files") {
		t.Fatal(s)
	}
	out.Reset()
	Run([]string{"m365", "mail", "list", "--help"}, d)
	s = out.String()
	if !strings.Contains(s, "default 10") || !strings.Contains(s, "max 50") {
		t.Fatal(s)
	}
	out.Reset()
	Run([]string{"m365", "teams", "list", "--help"}, d)
	if !strings.Contains(out.String(), "default 20") {
		t.Fatal(out.String())
	}
	out.Reset()
	Run([]string{"m365", "teams", "send", "--help"}, d)
	if !strings.Contains(out.String(), "10 MiB") {
		t.Fatal(out.String())
	}
}

func TestTeamsSendHelpNamesFormatMD(t *testing.T) {
	d, out, _ := testDeps()
	code := Run([]string{"m365", "teams", "send", "--help"}, d)
	if code != 0 {
		t.Fatal(code)
	}
	s := out.String()
	for _, want := range []string{"--format md", "--html", "converts", "already"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
}

func TestMailReplyHelpNamesHTML(t *testing.T) {
	d, out, _ := testDeps()
	code := Run([]string{"m365", "mail", "reply", "--help"}, d)
	if code != 0 {
		t.Fatal(code)
	}
	s := out.String()
	for _, want := range []string{"--html", "message.body", "comment"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
}
