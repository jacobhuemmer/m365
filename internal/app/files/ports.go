package files

import (
	"context"
	"strings"

	"github.com/masonhuemmer/m365/internal/app/auth"
	"github.com/masonhuemmer/m365/internal/domain"
)

type ListQuery struct {
	Folder    string
	Top       int
	PageToken string
}

type UploadInput struct {
	Folder string
	File   domain.OutboundFile
	DryRun bool
}

type FolderInput struct {
	Parent string
	Name   string
	DryRun bool
}

type MoveInput struct {
	ID     string
	Folder string
	Name   string
	DryRun bool
}

type Store interface {
	Root(ctx context.Context) (domain.DriveRoot, error)
	List(ctx context.Context, q ListQuery) (domain.DrivePage, error)
	Get(ctx context.Context, id string) (domain.DriveItem, error)
	Download(ctx context.Context, id string) ([]byte, domain.DriveItem, error)
	Upload(ctx context.Context, in UploadInput) (string, error)
	CreateFolder(ctx context.Context, in FolderInput) (string, error)
	Delete(ctx context.Context, id string) error
	Move(ctx context.Context, in MoveInput) error
}

func Root(ctx context.Context, st Store, sess domain.Session) (domain.DriveRoot, error) {
	if err := auth.RequireFiles(sess); err != nil {
		return domain.DriveRoot{}, err
	}
	return st.Root(ctx)
}

func List(ctx context.Context, st Store, sess domain.Session, q ListQuery) (domain.DrivePage, error) {
	if err := auth.RequireFiles(sess); err != nil {
		return domain.DrivePage{}, err
	}
	n, err := domain.NormalizeTop(q.Top, domain.DefaultFilesTop)
	if err != nil {
		return domain.DrivePage{}, err
	}
	q.Top = n
	if q.Folder == "" {
		q.Folder = "root"
	}
	return st.List(ctx, q)
}

func Get(ctx context.Context, st Store, sess domain.Session, id string) (domain.DriveItem, error) {
	if err := auth.RequireFiles(sess); err != nil {
		return domain.DriveItem{}, err
	}
	if strings.TrimSpace(id) == "" {
		return domain.DriveItem{}, domain.Usage("item id is required")
	}
	return st.Get(ctx, id)
}

func Download(ctx context.Context, st Store, sess domain.Session, id, dest string, overwrite bool, write func(string, []byte, bool) error) (string, error) {
	if err := auth.RequireFiles(sess); err != nil {
		return "", err
	}
	if strings.TrimSpace(id) == "" {
		return "", domain.Usage("item id is required")
	}
	if dest == "" {
		return "", domain.Usage("destination path is required")
	}
	b, _, err := st.Download(ctx, id)
	if err != nil {
		return "", err
	}
	if write == nil {
		return "", domain.Usage("write not available")
	}
	if err := write(dest, b, overwrite); err != nil {
		return "", err
	}
	return dest, nil
}

func Upload(ctx context.Context, st Store, sess domain.Session, in UploadInput) (string, error) {
	if err := auth.RequireFiles(sess); err != nil {
		return "", err
	}
	if err := domain.ValidateUpload(in.File); err != nil {
		return "", err
	}
	if in.Folder == "" {
		in.Folder = "root"
	}
	if in.DryRun {
		return "", nil
	}
	return st.Upload(ctx, in)
}

func CreateFolder(ctx context.Context, st Store, sess domain.Session, in FolderInput) (string, error) {
	if err := auth.RequireFiles(sess); err != nil {
		return "", err
	}
	if strings.TrimSpace(in.Name) == "" {
		return "", domain.Usage("folder name is required")
	}
	if in.Parent == "" {
		in.Parent = "root"
	}
	if in.DryRun {
		return "", nil
	}
	return st.CreateFolder(ctx, in)
}

func Delete(ctx context.Context, st Store, sess domain.Session, id string, dry bool) error {
	if err := auth.RequireFiles(sess); err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" {
		return domain.Usage("item id is required")
	}
	if dry {
		return nil
	}
	return st.Delete(ctx, id)
}

func Move(ctx context.Context, st Store, sess domain.Session, in MoveInput) error {
	if err := auth.RequireFiles(sess); err != nil {
		return err
	}
	if strings.TrimSpace(in.ID) == "" {
		return domain.Usage("item id is required")
	}
	if in.Folder == "" && in.Name == "" {
		return domain.Usage("new parent or new name is required")
	}
	if in.DryRun {
		return nil
	}
	return st.Move(ctx, in)
}
