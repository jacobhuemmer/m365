package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Findings from the Codex review of PR #17.

func TestGenerateFailsOnCaseOnlyClash(t *testing.T) {
	dir := layout(t, map[string]string{
		"features/mail/A-b.feature": "Feature: a\n",
		"features/mail/a_b.feature": "Feature: b\n",
	})
	out := filepath.Join(dir, "acceptance", "generated")
	err := generate(filepath.Join(dir, "features"), out)
	if err == nil {
		t.Fatal("want an error: mail_A_b and mail_a_b are the same file on a case-insensitive disk")
	}
	for _, want := range []string{"mail/A-b.feature", "mail/a_b.feature"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q lacks %q", err, want)
		}
	}
}

func TestGenerateClashKeepsOldTests(t *testing.T) {
	old := "package generated\n// previous run\n"
	dir := layout(t, map[string]string{
		"features/mail/send-reply.feature":                        "Feature: a\n",
		"features/mail/send_reply.feature":                        "Feature: b\n",
		"acceptance/generated/mail_send_reply_acceptance_test.go": old,
	})
	out := filepath.Join(dir, "acceptance", "generated")
	if err := generate(filepath.Join(dir, "features"), out); err == nil {
		t.Fatal("want a clash error")
	}
	got, err := os.ReadFile(filepath.Join(out, "mail_send_reply_acceptance_test.go"))
	if err != nil || string(got) != old {
		t.Fatalf("old generated test must survive a clash; got %q, %v", got, err)
	}
}

func TestListFeaturesMatchesGenerate(t *testing.T) {
	dir := layout(t, map[string]string{
		"features/top.feature":      "Feature: top\n",
		"features/a/b/c.feature":    "Feature: deep\n",
		"features/cli/help.feature": "Feature: help\n",
		"features/cli/notes.txt":    "not a feature\n",
	})
	root := filepath.Join(dir, "features")
	list, err := listFeatures(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []feature{
		{Name: "a_b_c", Path: filepath.Join(root, "a", "b", "c.feature")},
		{Name: "cli_help", Path: filepath.Join(root, "cli", "help.feature")},
		{Name: "top", Path: filepath.Join(root, "top.feature")},
	}
	if !reflect.DeepEqual(list, want) {
		t.Fatalf("listFeatures\ngot  %v\nwant %v", list, want)
	}
	out := filepath.Join(dir, "acceptance", "generated")
	if err := generate(root, out); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range generated(t, out) {
		names = append(names, strings.TrimSuffix(f, generatedSuffix))
	}
	if strings.Join(names, ",") != "a_b_c,cli_help,top" {
		t.Fatalf("generate wrote %v, want the listed features", names)
	}
}
