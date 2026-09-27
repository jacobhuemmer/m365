package steps

import (
	"testing"

	"github.com/masonhuemmer/m365/acceptance/runtime"
)

// newFeatures lists the SDO-563 feature files. The runner still passes
// unmatched steps (SDO-566), so these must prove every step has a handler.
var newFeatures = []string{
	"../../features/mail/mail-rendered.feature",
	"../../features/teams/teams-rendered.feature",
	"../../features/mail/mail-format-check.feature",
	"../../features/teams/teams-format-check.feature",
}

func TestNewFeatureStepsMatchHandlers(t *testing.T) {
	for _, path := range newFeatures {
		feat, err := runtime.Parse(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(feat.Scenarios) == 0 {
			t.Fatalf("%s: no scenarios", path)
		}
		for _, sc := range feat.Scenarios {
			for _, st := range sc.Steps {
				if !runtime.Matches(st.Text) {
					t.Errorf("%s: %q: step %q has no handler", path, sc.Name, st.Text)
				}
			}
		}
	}
}
