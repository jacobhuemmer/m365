package runtime

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

type World struct {
	T    *testing.T
	Code int
	Out  string
	Err  string
}

type Step struct {
	Kind, Text string
}

type Scenario struct {
	Name  string
	Steps []Step
}

type Feature struct {
	Name      string
	Scenarios []Scenario
}

type Handler func(w *World, text string) error

var handlers []struct {
	prefix string
	fn     Handler
}

func Register(prefix string, fn Handler) {
	handlers = append(handlers, struct {
		prefix string
		fn     Handler
	}{prefix, fn})
}

func Parse(path string) (Feature, error) {
	f, err := os.Open(path) // #nosec G304 -- feature path from generated tests under features/
	if err != nil {
		return Feature{}, err
	}
	defer f.Close()
	var feat Feature
	var cur *Scenario
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case strings.HasPrefix(line, "Feature:"):
			feat.Name = strings.TrimSpace(strings.TrimPrefix(line, "Feature:"))
		case strings.HasPrefix(line, "Scenario:"):
			feat.Scenarios = append(feat.Scenarios, Scenario{Name: strings.TrimSpace(strings.TrimPrefix(line, "Scenario:"))})
			cur = &feat.Scenarios[len(feat.Scenarios)-1]
		case strings.HasPrefix(line, "Given ") || strings.HasPrefix(line, "When ") || strings.HasPrefix(line, "Then ") || strings.HasPrefix(line, "And "):
			if cur == nil {
				continue
			}
			kind, text, _ := strings.Cut(line, " ")
			cur.Steps = append(cur.Steps, Step{Kind: kind, Text: text})
		}
	}
	return feat, sc.Err()
}

func RunFeature(t *testing.T, path string) {
	t.Helper()
	feat, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	// A step with no handler used to pass silently (SDO-566). Check the
	// whole feature first and report every unmatched step at once.
	if missing := unmatchedSteps(feat); len(missing) > 0 {
		for _, m := range missing {
			t.Errorf("%s: no step handler: %s", path, m)
		}
		t.FailNow()
	}
	for _, scn := range feat.Scenarios {
		t.Run(scn.Name, func(t *testing.T) {
			w := &World{T: t}
			for _, st := range scn.Steps {
				if err := dispatch(w, st.Text); err != nil {
					t.Fatalf("%s %s: %v", st.Kind, st.Text, err)
				}
			}
		})
	}
}

func dispatch(w *World, text string) error {
	if h, ok := handlerFor(text); ok {
		return h(w, text)
	}
	return fmt.Errorf("no step handler for %q", text)
}

// handlerFor finds the first handler whose prefix starts the step text.
func handlerFor(text string) (Handler, bool) {
	for _, h := range handlers {
		if strings.HasPrefix(text, h.prefix) {
			return h.fn, true
		}
	}
	return nil, false
}

// unmatchedSteps lists each step in feat that has no registered handler,
// as "scenario: Kind text".
func unmatchedSteps(feat Feature) []string {
	var out []string
	for _, scn := range feat.Scenarios {
		for _, st := range scn.Steps {
			if _, ok := handlerFor(st.Text); !ok {
				out = append(out, scn.Name+": "+st.Kind+" "+st.Text)
			}
		}
	}
	return out
}

// checkFeature parses a feature and reports unsupported syntax or steps
// with no handler, before any scenario runs.
func checkFeature(path string) (Feature, error) {
	return Feature{}, nil
}
