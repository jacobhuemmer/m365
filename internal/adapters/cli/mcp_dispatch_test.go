package cli

import (
	"reflect"
	"testing"
)

func TestFlagMapToArgs(t *testing.T) {
	got, err := FlagMapToArgs("mail", "list", nil, map[string]any{"top": float64(10), "unread": true, "folder": "inbox"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"m365", "mail", "list", "--folder", "inbox", "--top", "10", "--unread"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v != %v", got, want)
	}
	got, err = FlagMapToArgs("mail", "send", nil, map[string]any{"to": []any{"a@b.c", "d@e.f"}, "unread": false})
	if err != nil {
		t.Fatal(err)
	}
	want = []string{"m365", "mail", "send", "--to", "a@b.c", "--to", "d@e.f"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
}

func TestWriteGateInjectsDryRun(t *testing.T) {
	args, err := buildRunArgs("mail", "send", nil, map[string]any{"dry-run": false, "to": "a@b.c"}, false)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range args {
		if a == "--dry-run" {
			found = true
		}
	}
	if !found {
		t.Fatal(args)
	}
	args, err = buildRunArgs("mail", "list", nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range args {
		if a == "--dry-run" {
			t.Fatal(args)
		}
	}
}
