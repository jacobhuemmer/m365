package files

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestDryRunUploadNoMutate(t *testing.T) {
	st := &stub{}
	id, err := Upload(context.Background(), st, sessFiles(), UploadInput{
		File: domain.OutboundFile{Name: "note.txt", Size: 12}, DryRun: true,
	})
	if err != nil || id != "" || st.mut != 0 {
		t.Fatalf("%q %v mut=%d", id, err, st.mut)
	}
}

func TestDownloadMissingPath(t *testing.T) {
	st := &stub{data: []byte("synthetic-ok"), item: domain.DriveItem{ID: "file-1"}}
	_, err := Download(context.Background(), st, sessFiles(), "file-1", "", false, nil)
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}

func TestUploadZeroByte(t *testing.T) {
	st := &stub{}
	_, err := Upload(context.Background(), st, sessFiles(), UploadInput{File: domain.OutboundFile{Name: "z", Size: 0}})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}

func TestMoveRequiresParentOrName(t *testing.T) {
	st := &stub{}
	if err := Move(context.Background(), st, sessFiles(), MoveInput{ID: "file-1"}); domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}
