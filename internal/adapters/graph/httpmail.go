package graph

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/domain"
)

type HTTPClient struct {
	Base    string
	Token   string
	TokenFn func() string
	Client  *http.Client
}

func (c *HTTPClient) bearer() string {
	if c.TokenFn != nil {
		if t := c.TokenFn(); t != "" {
			return t
		}
	}
	return c.Token
}

func (c *HTTPClient) httpc() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return http.DefaultClient
}

func (c *HTTPClient) do(ctx context.Context, method, path string) (*http.Response, error) {
	raw := path
	if !strings.HasPrefix(path, "http") {
		raw = strings.TrimRight(c.Base, "/") + path
	} else if !allowedNext(c.Base, path) {
		return nil, domain.Usage("invalid page token")
	}
	req, err := http.NewRequestWithContext(ctx, method, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.bearer())
	req.Header.Set("Prefer", MailBodyPrefer)
	res, err := c.httpc().Do(req)
	if err != nil {
		return nil, domain.Service(err.Error())
	}
	if res.StatusCode >= 400 {
		b, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		return nil, MapGraphError(res.StatusCode, b)
	}
	return res, nil
}

func allowedNext(base, next string) bool {
	u, err := url.Parse(next)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return false
	}
	host := strings.ToLower(u.Host)
	if strings.Contains(host, "graph.microsoft.com") {
		return true
	}
	b, err := url.Parse(base)
	return err == nil && strings.EqualFold(b.Host, u.Host)
}

func encodeNext(next string) string {
	if next == "" {
		return ""
	}
	return "p." + base64.RawURLEncoding.EncodeToString([]byte(next))
}

func decodeNext(tok string) (string, bool) {
	if !strings.HasPrefix(tok, "p.") {
		return "", false
	}
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(tok, "p."))
	if err != nil {
		return "", false
	}
	return string(b), true
}

func (c *HTTPClient) List(ctx context.Context, q mail.ListQuery) (domain.MailPage, error) {
	sel := "$select=id,subject,from,toRecipients,ccRecipients,receivedDateTime,isRead,hasAttachments,conversationId"
	u := "/me/messages?$top=" + strconv.Itoa(q.Top) + "&" + sel + "&$expand=attachments"
	if q.PageToken != "" {
		if next, ok := decodeNext(q.PageToken); ok {
			u = next
		} else {
			u += "&$skiptoken=" + url.QueryEscape(q.PageToken)
		}
	}
	res, err := c.do(ctx, http.MethodGet, u)
	if err != nil {
		return domain.MailPage{}, err
	}
	defer res.Body.Close()
	var raw graphList
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return domain.MailPage{}, domain.Service("invalid mail list")
	}
	items := make([]domain.MailMessage, 0, len(raw.Value))
	for _, g := range raw.Value {
		m := g.toMail()
		m.Body = ""
		items = append(items, m)
	}
	p := domain.MailPage{Limit: q.Top, Count: len(items), Items: items}
	if tok := encodeNext(raw.Next); tok != "" {
		p.NextPage = &tok
	}
	return p, nil
}

func (c *HTTPClient) Get(ctx context.Context, id string) (domain.MailMessage, error) {
	res, err := c.do(ctx, http.MethodGet, "/me/messages/"+url.PathEscape(id)+"?$expand=attachments")
	if err != nil {
		return domain.MailMessage{}, err
	}
	defer res.Body.Close()
	var g graphMsg
	if err := json.NewDecoder(res.Body).Decode(&g); err != nil {
		return domain.MailMessage{}, domain.Service("invalid message")
	}
	return g.toMail(), nil
}

func (c *HTTPClient) Thread(ctx context.Context, id string, bodies bool) (domain.MailThread, error) {
	msg, err := c.Get(ctx, id)
	if err != nil {
		return domain.MailThread{}, err
	}
	res, err := c.do(ctx, http.MethodGet, "/me/messages?$filter="+url.QueryEscape("conversationId eq '"+msg.Conversation+"'")+"&$expand=attachments")
	if err != nil {
		return domain.MailThread{}, err
	}
	defer res.Body.Close()
	var raw graphList
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return domain.MailThread{}, domain.Service("invalid thread")
	}
	items := make([]domain.MailMessage, 0, len(raw.Value))
	for _, g := range raw.Value {
		m := g.toMail()
		if !bodies {
			m.Body = ""
		}
		items = append(items, m)
	}
	return domain.MailThread{ConversationID: msg.Conversation, Count: len(items), Items: items}, nil
}

func (c *HTTPClient) Attachments(ctx context.Context, messageID string) ([]domain.Attachment, error) {
	m, err := c.Get(ctx, messageID)
	if err != nil {
		return nil, err
	}
	return m.Attachments, nil
}

type graphList struct {
	Value []graphMsg `json:"value"`
	Next  string     `json:"@odata.nextLink"`
}

type graphMsg struct {
	ID             string `json:"id"`
	Subject        string `json:"subject"`
	ConversationID string `json:"conversationId"`
	Body           struct {
		Content string `json:"content"`
	} `json:"body"`
	From struct {
		EmailAddress struct {
			Name, Address string
		} `json:"emailAddress"`
	} `json:"from"`
	HasAttachments bool `json:"hasAttachments"`
	Attachments    []struct {
		ID, Name, ContentType string
		Size                  int64
	} `json:"attachments"`
}

func (g graphMsg) toMail() domain.MailMessage {
	atts := make([]domain.Attachment, 0, len(g.Attachments))
	for _, a := range g.Attachments {
		atts = append(atts, domain.Attachment{ID: a.ID, Name: a.Name, Size: a.Size, ContentType: a.ContentType})
	}
	return domain.MailMessage{
		ID: g.ID, Conversation: g.ConversationID, Subject: g.Subject,
		From: domain.Person{Name: g.From.EmailAddress.Name, Address: g.From.EmailAddress.Address},
		Body: g.Body.Content, HasAttachments: g.HasAttachments, Attachments: atts,
	}
}
