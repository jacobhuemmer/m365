package runtime

import (
	"reflect"
	"strings"
	"testing"
)

// SDO-566: a step with no handler must fail, not pass silently.

func TestUnmatchedStepsListsEveryOne(t *testing.T) {
	Register("sdo566 known", func(*World, string) error { return nil })
	feat := Feature{Scenarios: []Scenario{
		{Name: "first", Steps: []Step{
			{Kind: "Given", Text: "sdo566 known step"},
			{Kind: "When", Text: "sdo566 missing one"},
		}},
		{Name: "second", Steps: []Step{
			{Kind: "Then", Text: "sdo566 missing two"},
			{Kind: "And", Text: "sdo566 known"},
		}},
	}}
	want := []string{"first: When sdo566 missing one", "second: Then sdo566 missing two"}
	if got := unmatchedSteps(feat); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestUnmatchedStepsNoneWhenAllMatch(t *testing.T) {
	Register("sdo566 all", func(*World, string) error { return nil })
	feat := Feature{Scenarios: []Scenario{{Name: "s", Steps: []Step{{Kind: "Given", Text: "sdo566 all good"}}}}}
	if got := unmatchedSteps(feat); len(got) != 0 {
		t.Fatalf("got %q want none", got)
	}
}

func TestDispatchUnmatchedStepIsAnError(t *testing.T) {
	err := dispatch(&World{T: t}, "sdo566 nothing handles this")
	if err == nil || !strings.Contains(err.Error(), `no step handler for "sdo566 nothing handles this"`) {
		t.Fatalf("got %v", err)
	}
}
