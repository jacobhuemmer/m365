package runtime

import "testing"

// Doc strings and data tables are step arguments the runner can't pass to a
// handler; skipping them would let a step ignore its input.
func TestParseRejectsStepArguments(t *testing.T) {
	for name, body := range map[string]string{
		"doc string": "Feature: f\n  Scenario: s\n    Given a\n      \"\"\"\n      text\n      \"\"\"\n",
		"data table": "Feature: f\n  Scenario: s\n    Given a\n      | x | y |\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(writeFeature(t, body)); err == nil {
				t.Fatal("want an error: step arguments are not supported")
			}
		})
	}
}

func TestParseIgnoresTagsAndComments(t *testing.T) {
	p := writeFeature(t, "@smoke\nFeature: f\n  # a comment\n  @tag\n  Scenario: s\n    Given a\n")
	f, err := Parse(p)
	if err != nil || len(f.Scenarios) != 1 || len(f.Scenarios[0].Steps) != 1 {
		t.Fatalf("got %+v, %v", f, err)
	}
}
