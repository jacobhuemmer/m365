package graph

import (
	"context"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/files"
	"github.com/masonhuemmer/m365/internal/domain"
)

func TestHTTPFilesViaFakeServer(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPFiles{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-files"}}
	r, err := c.Root(context.Background())
	if err != nil || r.ID == "" {
		t.Fatal(err, r)
	}
	p, err := c.List(context.Background(), files.ListQuery{Top: 20})
	if err != nil || p.Count == 0 {
		t.Fatal(err, p)
	}
	if p.NextPage != nil && (strings.Contains(*p.NextPage, "http") || strings.Contains(*p.NextPage, "graph.microsoft.com")) {
		t.Fatalf("next_page %v", p.NextPage)
	}
	it, err := c.Get(context.Background(), "file-1")
	if err != nil || it.Name != "note.txt" {
		t.Fatal(err, it)
	}
	_, err = c.Get(context.Background(), "missing")
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatal(err)
	}
}
