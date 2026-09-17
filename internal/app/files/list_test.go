package files

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

type stub struct {
	root domain.DriveRoot
	page domain.DrivePage
	item domain.DriveItem
	data []byte
	err  error
	mut  int
}

func (s *stub) Root(context.Context) (domain.DriveRoot, error) { return s.root, s.err }
func (s *stub) List(context.Context, ListQuery) (domain.DrivePage, error) {
	return s.page, s.err
}
func (s *stub) Get(context.Context, string) (domain.DriveItem, error) { return s.item, s.err }
func (s *stub) Download(context.Context, string) ([]byte, domain.DriveItem, error) {
	return s.data, s.item, s.err
}
func (s *stub) Upload(context.Context, UploadInput) (string, error) {
	s.mut++
	return "file-new", s.err
}
func (s *stub) CreateFolder(context.Context, FolderInput) (string, error) {
	s.mut++
	return "folder-new", s.err
}
func (s *stub) Delete(context.Context, string) error  { s.mut++; return s.err }
func (s *stub) Move(context.Context, MoveInput) error { s.mut++; return s.err }

func sessFiles() domain.Session {
	return domain.Session{SignedIn: true, SessionUsable: true, FilesConsented: true}
}

func TestListDefaults(t *testing.T) {
	st := &stub{page: domain.DrivePage{Count: 1, Items: []domain.DriveItem{{ID: "file-1"}}}}
	p, err := List(context.Background(), st, sessFiles(), ListQuery{})
	if err != nil || p.Count != 1 {
		t.Fatal(err, p)
	}
	_, err = List(context.Background(), st, sessFiles(), ListQuery{Top: 51})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	_, err = List(context.Background(), st, domain.SignedOut(), ListQuery{})
	if domain.ExitOf(err) != domain.ExitAuth {
		t.Fatal(err)
	}
}
