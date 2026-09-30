package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/masonhuemmer/m365/internal/app/teams"
	"github.com/masonhuemmer/m365/internal/domain"
)

// teamsFilesFolder is where Teams itself keeps files shared in chats.
const teamsFilesFolder = "Microsoft Teams Chat Files"

var etagGUID = regexp.MustCompile(`\{([0-9A-Fa-f-]{36})\}`)

type sharedFile struct {
	name, webURL, attachID string
}

// shareFiles uploads each file to OneDrive and grants in.ShareWith read access
// without emailing anyone. Any failure returns an error naming what was already
// uploaded, so the caller never posts a message for files people cannot open.
func (c *HTTPTeams) shareFiles(ctx context.Context, in teams.SendInput) ([]sharedFile, error) {
	var done []sharedFile
	for _, f := range in.Files {
		item, err := c.uploadToOneDrive(ctx, f)
		if err != nil {
			return nil, partialUpload(err, done)
		}
		done = append(done, item.file)
		if err := c.grantRead(ctx, item.id, in.ShareWith); err != nil {
			return nil, partialUpload(err, done)
		}
	}
	return done, nil
}

type uploadedItem struct {
	id   string
	file sharedFile
}

func (c *HTTPTeams) uploadToOneDrive(ctx context.Context, f domain.OutboundFile) (uploadedItem, error) {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return uploadedItem{}, domain.Usagef("unreadable file: %s", f.Path)
	}
	path := "/me/drive/root:/" + url.PathEscape(teamsFilesFolder) + "/" + url.PathEscape(f.Name) +
		":/content?@microsoft.graph.conflictBehavior=rename"
	res, err := c.request(ctx, http.MethodPut, strings.TrimRight(c.Base, "/")+path, data, "application/octet-stream", "")
	if err != nil {
		return uploadedItem{}, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return uploadedItem{}, MapStatus(res.StatusCode)
	}
	var g struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		WebURL string `json:"webUrl"`
		ETag   string `json:"eTag"`
	}
	if err := json.NewDecoder(res.Body).Decode(&g); err != nil {
		return uploadedItem{}, domain.Service("invalid upload response")
	}
	m := etagGUID.FindStringSubmatch(g.ETag)
	if g.ID == "" || g.WebURL == "" || m == nil {
		return uploadedItem{}, domain.Service("upload response missing id, webUrl or eTag")
	}
	return uploadedItem{id: g.ID, file: sharedFile{
		name: g.Name, webURL: g.WebURL, attachID: strings.ToLower(strings.ReplaceAll(m[1], "-", "")),
	}}, nil
}

// grantRead gives the listed people read access. sendInvitation is false, so
// nobody is emailed; the chat message is the notification.
func (c *HTTPTeams) grantRead(ctx context.Context, itemID string, emails []string) error {
	if len(emails) == 0 {
		return nil
	}
	recipients := make([]map[string]string, 0, len(emails))
	for _, e := range emails {
		recipients = append(recipients, map[string]string{"email": e})
	}
	return c.doJSON(ctx, http.MethodPost, "/me/drive/items/"+url.PathEscape(itemID)+"/invite", map[string]any{
		"requireSignIn": true, "sendInvitation": false, "roles": []string{"read"}, "recipients": recipients,
	})
}

// attachRefs appends an <attachment> tag per file to the HTML body and returns
// the matching reference attachments for the message.
func attachRefs(content string, files []sharedFile) (string, []map[string]any) {
	atts := make([]map[string]any, 0, len(files))
	for _, f := range files {
		content += `<attachment id="` + f.attachID + `"></attachment>`
		atts = append(atts, map[string]any{"id": f.attachID, "contentType": "reference", "contentUrl": f.webURL, "name": f.name})
	}
	return content, atts
}

// partialUpload keeps the failure's class and says which files already sit in
// OneDrive, since the message was not posted.
func partialUpload(err error, uploaded []sharedFile) error {
	if len(uploaded) == 0 {
		return err
	}
	class := domain.ClassService
	msg := err.Error()
	if de, ok := err.(*domain.Error); ok {
		class, msg = de.Class, de.Message
	}
	return &domain.Error{Class: class, Message: fmt.Sprintf(
		"%s (already uploaded to OneDrive folder %q, message not posted: %s)", msg, teamsFilesFolder, uploadedNames(uploaded))}
}

func uploadedNames(files []sharedFile) string {
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, f.name)
	}
	return strings.Join(names, ", ")
}
