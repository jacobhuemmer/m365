package cli

import (
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
)

func TestMCPDeleteUnknownFlagKeepsDryRun(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	mem := d.Calendar.(graph.CalendarAPI).Memory
	files := d.Files.(graph.FilesAPI).Memory
	events, items := len(mem.CalEvents), len(files.Items)
	for _, tc := range []struct{ ns, id string }{
		{"calendar", mem.CalEvents[0].ID},
		{"files", files.Items[0].ID},
	} {
		args, err := buildRunArgs(tc.ns, "delete", []string{tc.id}, map[string]any{"a": "x"}, false)
		if err != nil {
			t.Fatal(err)
		}
		if c := Run(args, d); c == 0 {
			t.Fatalf("%s delete with unknown flag succeeded", tc.ns)
		}
	}
	if len(mem.CalEvents) != events || len(files.Items) != items {
		t.Fatalf("real delete: events %d->%d items %d->%d", events, len(mem.CalEvents), items, len(files.Items))
	}
}
