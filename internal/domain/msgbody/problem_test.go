package msgbody

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestProblemsJSON(t *testing.T) {
	cases := []struct {
		name string
		in   Problems
		want string
	}{
		{"nil is empty list", nil, `[]`},
		{"empty is empty list", Problems{}, `[]`},
		{"one", Problems{{Rule: RuleBrokenHTML, Detail: "unclosed <p>"}}, `[{"rule":"broken-html","detail":"unclosed <p>"}]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The CLI encodes with HTML escaping off; match it.
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetEscapeHTML(false)
			if err := enc.Encode(tc.in); err != nil {
				t.Fatal(err)
			}
			if got := strings.TrimSuffix(buf.String(), "\n"); got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestProblemsErr(t *testing.T) {
	if err := (Problems{}).Err(); err != nil {
		t.Fatalf("no problems: got %v", err)
	}
	cases := []struct {
		name string
		in   Problems
		want string
	}{
		{
			"one",
			Problems{{Rule: RuleBrokenHTML, Detail: "unclosed <p>"}},
			"1 format problem: broken-html: unclosed <p>",
		},
		{
			"two",
			Problems{{Rule: RuleBrokenHTML, Detail: "unclosed <p>"}, {Rule: RuleLeftoverMarkdown, Detail: "**bold**"}},
			"2 format problems: broken-html: unclosed <p>; leftover-markdown: **bold**",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var de *domain.Error
			if !errors.As(tc.in.Err(), &de) {
				t.Fatalf("got %v, want *domain.Error", tc.in.Err())
			}
			want := domain.Error{Class: domain.ClassUsage, Message: tc.want, Hint: "run with --preview to see them"}
			if *de != want {
				t.Fatalf("got %#v\nwant %#v", *de, want)
			}
		})
	}
}
