package graph

import (
	"context"
	"testing"
)

func TestDownloadSynthetic(t *testing.T) {
	m := Seed()
	b, a, err := m.Download(context.Background(), "msg-1", "att-1")
	if err != nil || string(b) != "synthetic-ok" || a.Name != "note.txt" {
		t.Fatalf("%s %+v %v", b, a, err)
	}
}
