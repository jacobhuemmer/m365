package graph

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/domain"
)

type HTTPMailDelta struct {
	HTTPClient *HTTPClient
}

func (d *HTTPMailDelta) LatestThread(ctx context.Context, messageID string, limit int) (domain.MailThread, error) {
	if d == nil || d.HTTPClient == nil {
		return domain.MailThread{}, domain.Service("mail thread client is unavailable")
	}
	thread, err := d.HTTPClient.Thread(ctx, messageID, true)
	if err != nil {
		return domain.MailThread{}, err
	}
	return latestThreadWindow(thread, limit)
}

func (d *HTTPMailDelta) Delta(ctx context.Context, query mail.DeltaQuery) (domain.MailDeltaPage, error) {
	if d == nil || d.HTTPClient == nil {
		return domain.MailDeltaPage{}, domain.Service("mail delta client is unavailable")
	}
	raw, err := d.deltaURL(query)
	if err != nil {
		return domain.MailDeltaPage{}, err
	}
	response, err := d.HTTPClient.request(ctx, http.MethodGet, raw, nil, "", MailBodyPrefer)
	if err != nil {
		return domain.MailDeltaPage{}, err
	}
	if response.StatusCode == http.StatusGone {
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		return domain.MailDeltaPage{}, domain.ErrMailDeltaReset
	}
	if response.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(response.Body)
		_ = response.Body.Close()
		return domain.MailDeltaPage{}, MapGraphError(response.StatusCode, body)
	}
	defer response.Body.Close()

	var rawPage struct {
		Value []graphMsg `json:"value"`
		Next  string     `json:"@odata.nextLink"`
		Delta string     `json:"@odata.deltaLink"`
	}
	if err := json.NewDecoder(response.Body).Decode(&rawPage); err != nil {
		return domain.MailDeltaPage{}, domain.Service("invalid mail delta")
	}
	page := domain.MailDeltaPage{
		Changes:    make([]domain.MailDeltaChange, 0, len(rawPage.Value)),
		NextToken:  rawPage.Next,
		DeltaToken: rawPage.Delta,
	}
	for _, item := range rawPage.Value {
		removed := len(item.Removed) > 0 && string(item.Removed) != "null"
		if !removed && item.ChangeKey == "" {
			return domain.MailDeltaPage{}, domain.Service("mail delta change is missing a revision")
		}
		message := item.toMail()
		message.Body = ""
		message.Attachments = nil
		page.Changes = append(page.Changes, domain.MailDeltaChange{
			Message:  message,
			Revision: item.ChangeKey,
			Removed:  removed,
		})
	}
	return page, nil
}

func (d *HTTPMailDelta) deltaURL(query mail.DeltaQuery) (string, error) {
	client := d.HTTPClient
	if query.Token != "" {
		if !allowedNext(client.Base, query.Token) {
			return "", domain.Usage("invalid mail delta cursor")
		}
		return query.Token, nil
	}
	values := url.Values{}
	values.Set("$select", "id,changeKey,subject,from,toRecipients,ccRecipients,receivedDateTime,isRead,hasAttachments,conversationId")
	path := "/me/mailFolders/" + url.PathEscape(strings.TrimSpace(query.Folder)) + "/messages/delta?" + values.Encode()
	return strings.TrimRight(client.Base, "/") + path, nil
}

func latestThreadWindow(thread domain.MailThread, limit int) (domain.MailThread, error) {
	if limit <= 0 {
		return domain.MailThread{}, domain.Usage("mail thread limit must be greater than zero")
	}
	if len(thread.Items) > limit {
		thread.Items = append([]domain.MailMessage(nil), thread.Items[len(thread.Items)-limit:]...)
	}
	return thread, nil
}
