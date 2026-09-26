package cli

import (
	"flag"
	"reflect"
	"testing"
)

func TestFlagMapToArgs(t *testing.T) {
	got, err := FlagMapToArgs("mail", "list", nil, map[string]any{"top": float64(10), "unread": true, "folder": "inbox"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"m365", "mail", "list", "--folder=inbox", "--top=10", "--unread"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v != %v", got, want)
	}
	got, err = FlagMapToArgs("mail", "send", nil, map[string]any{"to": []any{"a@b.c", "d@e.f"}, "unread": false})
	if err != nil {
		t.Fatal(err)
	}
	want = []string{"m365", "mail", "send", "--to=a@b.c", "--to=d@e.f"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
}

func TestFlagMapToArgsPreservesGlobalLookingValue(t *testing.T) {
	args, err := buildRunArgs("mail", "send", nil, map[string]any{"cc": "--json"}, false)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, _, rest := peelGlobals(args)
	fs := flag.NewFlagSet("mail send", flag.ContinueOnError)
	cc := fs.String("cc", "", "")
	dryRun := fs.Bool("dry-run", false, "")
	if err := fs.Parse(rest[2:]); err != nil {
		t.Fatal(err)
	}
	if *cc != "--json" || !*dryRun {
		t.Fatalf("cc=%q dry-run=%t args=%v", *cc, *dryRun, rest)
	}
}

func TestParseMixedExactRecipientBool(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	exact := fs.Bool("exact-recipient", false, "")
	to := fs.String("to", "", "")
	if err := parseMixed(fs, []string{"--exact-recipient", "recipient-id", "--to", "person@example.com"}); err != nil {
		t.Fatal(err)
	}
	if !*exact || *to != "person@example.com" || !reflect.DeepEqual(fs.Args(), []string{"recipient-id"}) {
		t.Fatalf("exact=%t to=%q args=%v", *exact, *to, fs.Args())
	}
}

func TestFlagMapToArgsRejectsInjectedKeys(t *testing.T) {
	for _, key := range []string{"dry-run=false", "exact-recipient=false", "to=evil@example.com", "to evil", "--to", "To", "to_"} {
		t.Run(key, func(t *testing.T) {
			if _, err := FlagMapToArgs("mail", "send", nil, map[string]any{key: true}); err == nil {
				t.Fatal("accepted injected flag key")
			}
		})
	}
}

func TestFlagMapToArgsRejectsPositionalFlag(t *testing.T) {
	if _, err := FlagMapToArgs("mail", "send", []string{"--to=evil@example.com"}, map[string]any{"to": "approved@example.com"}); err == nil {
		t.Fatal("accepted flag smuggled through positional arguments")
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
