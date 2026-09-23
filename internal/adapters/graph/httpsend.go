package graph

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/app/teams"
	"github.com/masonhuemmer/m365/internal/domain"
)

func (c *HTTPClient) doJSON(ctx context.Context, method, path string, payload any) error {
	var body []byte
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return domain.Usage(err.Error())
		}
		body = b
	}
	res, err := c.request(ctx, method, strings.TrimRight(c.Base, "/")+path, body, "application/json", "")
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return MapStatus(res.StatusCode)
	}
	return nil
}

func (c *HTTPClient) Send(ctx context.Context, in mail.SendInput) (string, error) {
	ctype := "Text"
	if in.HTML {
		ctype = "HTML"
	}
	to := make([]map[string]any, 0, len(in.To))
	for _, a := range in.To {
		to = append(to, map[string]any{"emailAddress": map[string]string{"address": a}})
	}
	cc := make([]map[string]any, 0, len(in.CC))
	for _, a := range in.CC {
		cc = append(cc, map[string]any{"emailAddress": map[string]string{"address": a}})
	}
	atts, err := fileAttachments(in.Files)
	if err != nil {
		return "", err
	}
	msg := map[string]any{
		"subject":      in.Subject,
		"body":         map[string]string{"contentType": ctype, "content": in.Body},
		"toRecipients": to,
	}
	if len(cc) > 0 {
		msg["ccRecipients"] = cc
	}
	if len(atts) > 0 {
		msg["attachments"] = atts
	}
	if err := c.doJSON(ctx, http.MethodPost, "/me/sendMail", map[string]any{"message": msg}); err != nil {
		return "", err
	}
	return "sent", nil
}

func (c *HTTPClient) Reply(ctx context.Context, in mail.ReplyInput) (string, error) {
	path := "/me/messages/" + url.PathEscape(in.ID) + "/reply"
	if in.All {
		path = "/me/messages/" + url.PathEscape(in.ID) + "/replyAll"
	}
	// Graph renders comment as HTML above the quoted thread. message.body
	// would replace the whole reply and drop the thread.
	comment := in.Body
	if !in.HTML {
		comment = plainTextToHTML(in.Body)
	}
	if err := c.doJSON(ctx, http.MethodPost, path, map[string]any{"comment": comment}); err != nil {
		return "", err
	}
	return "sent", nil
}

func (c *HTTPClient) Download(ctx context.Context, messageID, attach string) ([]byte, domain.Attachment, error) {
	res, err := c.do(ctx, http.MethodGet, "/me/messages/"+url.PathEscape(messageID)+"/attachments/"+url.PathEscape(attach)+"/$value")
	if err != nil {
		return nil, domain.Attachment{}, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, domain.Attachment{}, domain.Service(err.Error())
	}
	return b, domain.Attachment{ID: attach}, nil
}

func (c *HTTPTeams) Send(ctx context.Context, in teams.SendInput) (string, error) {
	// Teams collapses newlines in contentType text, so plain text goes out as HTML too.
	content := plainTextToHTML(in.Text)
	switch {
	case in.HTML:
		content = in.Text
	case in.MD:
		content = mdSubsetToHTML(in.Text)
	}
	body := map[string]any{"body": map[string]string{"contentType": "html", "content": content}}
	if err := c.doJSON(ctx, http.MethodPost, "/me/chats/"+url.PathEscape(in.ChatID)+"/messages", body); err != nil {
		return "", err
	}
	return "sent", nil
}

func fileAttachments(files []domain.OutboundFile) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(files))
	for _, f := range files {
		b, err := os.ReadFile(f.Path)
		if err != nil {
			return nil, domain.Usagef("unreadable file: %s", f.Path)
		}
		out = append(out, map[string]any{
			"@odata.type":  "#microsoft.graph.fileAttachment",
			"name":         f.Name,
			"contentBytes": base64.StdEncoding.EncodeToString(b),
		})
	}
	return out, nil
}
