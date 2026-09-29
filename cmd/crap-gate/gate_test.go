package main

import (
	"reflect"
	"strings"
	"testing"
)

const module = "github.com/masonhuemmer/m365"

func TestCrapScore(t *testing.T) {
	for _, tc := range []struct {
		c    int
		cov  float64
		want float64
	}{
		{1, 1.0, 1}, {10, 0, 110}, {12, 0, 156}, {39, 0.903, 40.4}, {15, 1.0, 15},
	} {
		if got := crapScore(tc.c, tc.cov); got != tc.want {
			t.Errorf("crapScore(%d, %v) = %v, want %v", tc.c, tc.cov, got, tc.want)
		}
	}
}

func TestParseCover(t *testing.T) {
	in := module + "/internal/app/mail/watch.go:16:\tWatch\t\t90.3%\n" +
		module + "/internal/a.go:3:\t\t(*x).y\t0.0%\n" +
		"total:\t\t\t(statements)\t76.4%\n"
	got, err := parseCover(strings.NewReader(in), module)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]float64{"internal/app/mail/watch.go:16": 0.903, "internal/a.go:3": 0}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if _, err := parseCover(strings.NewReader("garbage line\n"), module); err == nil {
		t.Fatal("want an error for a malformed line")
	}
}

func TestParseCyclo(t *testing.T) {
	in := "39 mail Watch internal/app/mail/watch.go:16:1\n4 cli (*x).y internal/a.go:3:1\n"
	got, err := parseCyclo(strings.NewReader(in), map[string]float64{"internal/app/mail/watch.go:16": 0.903})
	if err != nil {
		t.Fatal(err)
	}
	want := []fn{
		{File: "internal/app/mail/watch.go", Name: "Watch", Complexity: 39, Coverage: 0.903, Score: 40.4},
		{File: "internal/a.go", Name: "(*x).y", Complexity: 4, Coverage: 0, Score: 20},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
	if _, err := parseCyclo(strings.NewReader("x y\n"), nil); err == nil {
		t.Fatal("want an error for a malformed line")
	}
}

func TestParseBaseline(t *testing.T) {
	in := "# waiver text\n40.4 internal/app/mail/watch.go Watch\n\n20.0 internal/a.go (*x).y\n"
	got, err := parseBaseline(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]float64{"internal/app/mail/watch.go Watch": 40.4, "internal/a.go (*x).y": 20}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if _, err := parseBaseline(strings.NewReader("notanumber f g\n")); err == nil {
		t.Fatal("want an error for a malformed line")
	}
}

func TestCheckRatchet(t *testing.T) {
	fns := []fn{
		{File: "a.go", Name: "Same", Score: 40.4},
		{File: "b.go", Name: "New", Score: 16},
		{File: "c.go", Name: "Rose", Score: 25},
		{File: "d.go", Name: "Fixed", Score: 10},
		{File: "e.go", Name: "AtLimit", Score: 15},
		{File: "f.go", Name: "Better", Score: 18},
	}
	baseline := map[string]float64{
		"a.go Same": 40.4, "c.go Rose": 20, "d.go Fixed": 30, "g.go Gone": 50, "f.go Better": 22,
	}
	got := check(fns, baseline, 15)
	want := []violation{
		{Kind: "new", File: "b.go", Name: "New", Score: 16},
		{Kind: "rose", File: "c.go", Name: "Rose", Score: 25, Baseline: 20},
		{Kind: "stale", File: "d.go", Name: "Fixed", Score: 10, Baseline: 30},
		{Kind: "stale", File: "g.go", Name: "Gone", Baseline: 50},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
}
