package graph

import (
	"context"
	"strconv"

	"github.com/masonhuemmer/m365/internal/app/files"
	"github.com/masonhuemmer/m365/internal/domain"
)

func (m *Memory) Root(context.Context) (domain.DriveRoot, error) {
	if err := m.fail(); err != nil {
		return domain.DriveRoot{}, err
	}
	return m.Drive, nil
}

func (m *Memory) ListItems(_ context.Context, q files.ListQuery) (domain.DrivePage, error) {
	if err := m.fail(); err != nil {
		return domain.DrivePage{}, err
	}
	parent := q.Folder
	if parent == "" {
		parent = "root"
	}
	var items []domain.DriveItem
	for _, it := range m.Items {
		if it.ParentID == parent {
			items = append(items, it)
		}
	}
	var next *string
	if q.Top > 0 && len(items) > q.Top {
		tok := "p.more"
		next = &tok
		items = items[:q.Top]
	}
	return domain.DrivePage{Limit: q.Top, Count: len(items), NextPage: next, Items: items}, nil
}

func (m *Memory) GetItem(_ context.Context, id string) (domain.DriveItem, error) {
	if err := m.fail(); err != nil {
		return domain.DriveItem{}, err
	}
	if id == "root" {
		return domain.DriveItem{ID: m.Drive.ID, Name: m.Drive.Name, IsFolder: true}, nil
	}
	for _, it := range m.Items {
		if it.ID == id {
			return it, nil
		}
	}
	return domain.DriveItem{}, domain.NotFound("item not found")
}

func (m *Memory) DownloadItem(_ context.Context, id string) ([]byte, domain.DriveItem, error) {
	it, err := m.GetItem(context.Background(), id)
	if err != nil {
		return nil, domain.DriveItem{}, err
	}
	if it.IsFolder {
		return nil, domain.DriveItem{}, domain.Usage("cannot download a folder")
	}
	return m.FileBytes[id], it, nil
}

func (m *Memory) Upload(_ context.Context, in files.UploadInput) (string, error) {
	if err := m.fail(); err != nil {
		return "", err
	}
	id := "file-" + strconv.Itoa(len(m.Items)+1)
	m.mu.Lock()
	m.Items = append(m.Items, domain.DriveItem{ID: id, Name: in.File.Name, Size: in.File.Size, ParentID: in.Folder})
	if m.FileBytes == nil {
		m.FileBytes = map[string][]byte{}
	}
	m.FileBytes[id] = []byte("uploaded")
	m.mu.Unlock()
	return id, nil
}

func (m *Memory) CreateFolder(_ context.Context, in files.FolderInput) (string, error) {
	if err := m.fail(); err != nil {
		return "", err
	}
	id := "folder-" + strconv.Itoa(len(m.Items)+1)
	m.mu.Lock()
	m.Items = append(m.Items, domain.DriveItem{ID: id, Name: in.Name, IsFolder: true, ParentID: in.Parent})
	m.mu.Unlock()
	return id, nil
}

func (m *Memory) DeleteItem(_ context.Context, id string) error {
	if err := m.fail(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, it := range m.Items {
		if it.ID == id {
			m.Items = append(m.Items[:i], m.Items[i+1:]...)
			return nil
		}
	}
	return domain.NotFound("item not found")
}

func (m *Memory) Move(_ context.Context, in files.MoveInput) error {
	if err := m.fail(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, it := range m.Items {
		if it.ID == in.ID {
			if in.Folder != "" {
				it.ParentID = in.Folder
			}
			if in.Name != "" {
				it.Name = in.Name
			}
			m.Items[i] = it
			return nil
		}
	}
	return domain.NotFound("item not found")
}

type FilesAPI struct{ *Memory }

func (f FilesAPI) List(ctx context.Context, q files.ListQuery) (domain.DrivePage, error) {
	return f.ListItems(ctx, q)
}
func (f FilesAPI) Get(ctx context.Context, id string) (domain.DriveItem, error) {
	return f.GetItem(ctx, id)
}
func (f FilesAPI) Download(ctx context.Context, id string) ([]byte, domain.DriveItem, error) {
	return f.DownloadItem(ctx, id)
}
func (f FilesAPI) Delete(ctx context.Context, id string) error {
	return f.DeleteItem(ctx, id)
}
