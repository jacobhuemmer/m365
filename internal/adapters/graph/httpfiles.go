package graph

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/masonhuemmer/m365/internal/app/files"
	"github.com/masonhuemmer/m365/internal/domain"
)

type HTTPFiles struct{ *HTTPClient }

func (c *HTTPFiles) Root(ctx context.Context) (domain.DriveRoot, error) {
	res, err := c.do(ctx, http.MethodGet, "/me/drive")
	if err != nil {
		return domain.DriveRoot{}, err
	}
	defer res.Body.Close()
	var g struct {
		ID, Name string
	}
	if err := json.NewDecoder(res.Body).Decode(&g); err != nil {
		return domain.DriveRoot{}, domain.Service("invalid drive")
	}
	return domain.DriveRoot{ID: g.ID, Name: g.Name}, nil
}

func (c *HTTPFiles) List(ctx context.Context, q files.ListQuery) (domain.DrivePage, error) {
	path := "/me/drive/root/children?$top=" + strconv.Itoa(q.Top)
	if q.Folder != "" && q.Folder != "root" {
		path = "/me/drive/items/" + url.PathEscape(q.Folder) + "/children?$top=" + strconv.Itoa(q.Top)
	}
	if q.PageToken != "" {
		if next, ok := decodeNext(q.PageToken); ok {
			path = next
		}
	}
	res, err := c.do(ctx, http.MethodGet, path)
	if err != nil {
		return domain.DrivePage{}, err
	}
	defer res.Body.Close()
	var raw struct {
		Value []graphDriveItem `json:"value"`
		Next  string           `json:"@odata.nextLink"`
	}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return domain.DrivePage{}, domain.Service("invalid folder")
	}
	items := make([]domain.DriveItem, 0, len(raw.Value))
	for _, g := range raw.Value {
		items = append(items, g.toItem())
	}
	p := domain.DrivePage{Limit: q.Top, Count: len(items), Items: items}
	if tok := encodeNext(raw.Next); tok != "" {
		p.NextPage = &tok
	}
	return p, nil
}

func (c *HTTPFiles) Get(ctx context.Context, id string) (domain.DriveItem, error) {
	res, err := c.do(ctx, http.MethodGet, "/me/drive/items/"+url.PathEscape(id))
	if err != nil {
		return domain.DriveItem{}, err
	}
	defer res.Body.Close()
	var g graphDriveItem
	if err := json.NewDecoder(res.Body).Decode(&g); err != nil {
		return domain.DriveItem{}, domain.Service("invalid item")
	}
	return g.toItem(), nil
}

func (c *HTTPFiles) Download(ctx context.Context, id string) ([]byte, domain.DriveItem, error) {
	it, err := c.Get(ctx, id)
	if err != nil {
		return nil, domain.DriveItem{}, err
	}
	res, err := c.do(ctx, http.MethodGet, "/me/drive/items/"+url.PathEscape(id)+"/content")
	if err != nil {
		return nil, domain.DriveItem{}, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	return b, it, err
}

func (c *HTTPFiles) Upload(ctx context.Context, in files.UploadInput) (string, error) {
	data, err := os.ReadFile(in.File.Path)
	if err != nil {
		return "", domain.Usagef("unreadable file: %s", in.File.Path)
	}
	parent := in.Folder
	if parent == "" {
		parent = "root"
	}
	path := "/me/drive/items/" + url.PathEscape(parent) + ":/" + url.PathEscape(in.File.Name) + ":/content"
	if in.File.Size > 4*1024*1024 {
		res, err := c.do(ctx, http.MethodPost, "/me/drive/items/"+url.PathEscape(parent)+":/"+url.PathEscape(in.File.Name)+":/createUploadSession")
		if err != nil {
			return "", err
		}
		_ = res.Body.Close()
	}
	res, err := c.doBody(ctx, http.MethodPut, path, data)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var g struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&g)
	return g.ID, nil
}

func (c *HTTPFiles) CreateFolder(ctx context.Context, in files.FolderInput) (string, error) {
	parent := in.Parent
	if parent == "" {
		parent = "root"
	}
	body, _ := json.Marshal(map[string]any{"name": in.Name, "folder": map[string]any{}, "@microsoft.graph.conflictBehavior": "fail"})
	path := "/me/drive/root/children"
	if parent != "root" {
		path = "/me/drive/items/" + url.PathEscape(parent) + "/children"
	}
	res, err := c.doBody(ctx, http.MethodPost, path, body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var g struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&g)
	return g.ID, nil
}

func (c *HTTPFiles) Delete(ctx context.Context, id string) error {
	res, err := c.do(ctx, http.MethodDelete, "/me/drive/items/"+url.PathEscape(id))
	if err != nil {
		return err
	}
	_ = res.Body.Close()
	return nil
}

func (c *HTTPFiles) Move(ctx context.Context, in files.MoveInput) error {
	m := map[string]any{}
	if in.Name != "" {
		m["name"] = in.Name
	}
	if in.Folder != "" {
		m["parentReference"] = map[string]string{"id": in.Folder}
	}
	body, _ := json.Marshal(m)
	res, err := c.doBody(ctx, http.MethodPatch, "/me/drive/items/"+url.PathEscape(in.ID), body)
	if err != nil {
		return err
	}
	_ = res.Body.Close()
	return nil
}

type graphDriveItem struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Size                 int64     `json:"size"`
	LastModifiedDateTime string    `json:"lastModifiedDateTime"`
	WebURL               string    `json:"webUrl"`
	Folder               *struct{} `json:"folder"`
}

func (g graphDriveItem) toItem() domain.DriveItem {
	return domain.DriveItem{ID: g.ID, Name: g.Name, Size: g.Size, IsFolder: g.Folder != nil, LastModified: g.LastModifiedDateTime, WebURL: g.WebURL}
}
