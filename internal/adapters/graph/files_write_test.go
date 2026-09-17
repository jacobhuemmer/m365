package graph

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/files"
	"github.com/masonhuemmer/m365/internal/domain"
)

func TestHTTPFilesDownloadViaFake(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPFiles{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-files"}}
	b, it, err := c.Download(context.Background(), "file-1")
	if err != nil || it.ID != "file-1" || string(b) != "synthetic-ok" {
		t.Fatalf("%s %+v %v", b, it, err)
	}
	p := filepath.Join(t.TempDir(), "n.txt")
	if err := os.WriteFile(p, []byte("synthetic-ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	id, err := c.Upload(context.Background(), files.UploadInput{Folder: "root", File: domain.OutboundFile{Path: p, Name: "n.txt", Size: 12}})
	if err != nil {
		t.Fatal(err)
	}
	_ = id
}
