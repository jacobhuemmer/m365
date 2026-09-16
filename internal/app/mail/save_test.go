package mail

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestSaveRequiresPathAndOverwrite(t *testing.T) {
	st := &stub{}
	_, err := Save(context.Background(), st, sessMail(), "msg-1", "att-1", "", false, nil)
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	wrote := false
	path, err := Save(context.Background(), st, sessMail(), "msg-1", "att-1", "/tmp/x", false, func(p string, d []byte, ov bool) error {
		if ov {
			t.Fatal("overwrite")
		}
		wrote = true
		if string(d) != "synthetic-ok" {
			t.Fatalf("bytes %q", d)
		}
		return nil
	})
	if err != nil || path != "/tmp/x" || !wrote {
		t.Fatalf("%s %v", path, err)
	}
	_, err = Save(context.Background(), st, sessMail(), "msg-1", "att-1", "/tmp/x", false, func(string, []byte, bool) error {
		return domain.Usage("destination exists; pass --overwrite")
	})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}
