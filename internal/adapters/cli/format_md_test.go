package cli

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestMailFormatMarkdown(t *testing.T) {
	want := "<h1>Update</h1>\n<p><strong>Done</strong>, see <code>make</code>.</p>\n" +
		"<ul><li>one</li><li>two</li></ul>\n<p><a href=\"https://example.com\">ticket</a></p>"
	for _, args := range [][]string{
		{"mail", "send", "--to", "user@example.com", "--subject", "t", "--format", "md", "--body-file", "testdata/md/basic.md", "--dry-run"},
		{"mail", "reply", "msg-1", "--format", "md", "--body-file", "testdata/md/basic.md", "--dry-run"},
	} {
		d, out, errw := testDeps()
		loginAll(t, &d)
		out.Reset()
		if code := Run(append([]string{"m365"}, args...), d); code != 0 {
			t.Fatalf("%v: exit %d %s", args, code, errw.String())
		}
		var m map[string]any
		if err := json.Unmarshal(out.Bytes(), &m); err != nil {
			t.Fatal(err, out.String())
		}
		r, _ := m["rendered"].(map[string]any)
		if r["content"] != want {
			t.Fatalf("%v: rendered %q\nwant %q", args[:2], r["content"], want)
		}
		if !reflect.DeepEqual(m["format_problems"], []any{}) {
			t.Fatalf("%v: problems %v", args[:2], m["format_problems"])
		}
	}
}

func TestFormatFlagRejections(t *testing.T) {
	cmds := [][]string{
		{"mail", "send", "--to", "user@example.com", "--subject", "t", "--body", "b"},
		{"mail", "reply", "msg-1", "--body", "b"},
		{"teams", "send", "chat-1", "--text", "b"},
	}
	cases := []struct {
		extra []string
		msg   string
	}{
		{[]string{"--format", "html"}, `unsupported --format "html"; only md`},
		{[]string{"--html", "--format", "md"}, "use --html or --format md, not both"},
	}
	for _, c := range cmds {
		for _, tc := range cases {
			d, out, errw := testDeps()
			loginAll(t, &d)
			out.Reset()
			args := append(append(append([]string{"m365"}, c...), tc.extra...), "--dry-run")
			if code := Run(args, d); code != 3 {
				t.Fatalf("%v: exit %d want 3", args, code)
			}
			var got map[string]any
			if err := json.Unmarshal(errw.Bytes(), &got); err != nil {
				t.Fatalf("%v: stderr %q", args, errw.String())
			}
			if want := map[string]any{"class": "usage", "message": tc.msg}; !reflect.DeepEqual(got, want) {
				t.Fatalf("%v: stderr %#v want %#v", args, got, want)
			}
		}
	}
}

func TestSendHelpListsFormatMarkdown(t *testing.T) {
	for _, args := range [][]string{{"mail", "send"}, {"mail", "reply"}, {"teams", "send"}} {
		d, out, _ := testDeps()
		if code := Run(append(append([]string{"m365"}, args...), "--help"), d); code != 0 {
			t.Fatalf("%v help exit %d", args, code)
		}
		if !strings.Contains(out.String(), "--format md") {
			t.Fatalf("%v help lacks --format md:\n%s", args, out.String())
		}
	}
}
