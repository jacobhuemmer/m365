package graph

import (
	"bytes"
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
	var rdr io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return domain.Usage(err.Error())
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.Base, "/")+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.bearer())
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpc().Do(req)
	if err != nil {
		return domain.Service(err.Error())
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
	if err := c.doJSON(ctx, http.MethodPost, path, map[string]any{"comment": in.Body}); err != nil {
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
	body := map[string]any{"body": map[string]string{"contentType": "text", "content": in.Text}}
	if in.HTML || in.MD {
		body = map[string]any{"body": map[string]string{"contentType": "html", "content": in.Text}}
	}
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
