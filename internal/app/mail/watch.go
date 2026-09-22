package mail

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/masonhuemmer/m365/internal/app/auth"
	"github.com/masonhuemmer/m365/internal/domain"
)

func Watch(
	ctx context.Context,
	changes ChangeStore,
	states WatchStateStore,
	sink EventSink,
	sess domain.Session,
	query WatchQuery,
) error {
	if err := auth.Require(sess, true, false); err != nil {
		return err
	}
	folder := strings.TrimSpace(query.Folder)
	if folder == "" {
		folder = "inbox"
	}
	if strings.EqualFold(folder, "all") {
		return domain.Usage("mail watch requires a folder; all is not supported")
	}
	account := strings.ToLower(strings.TrimSpace(sess.Account))
	if account == "" {
		return domain.Usage("signed-in account is required for mail watch")
	}
	if changes == nil || states == nil || sink == nil {
		return domain.Usage("mail watch is unavailable")
	}

	state, err := states.Load(account, folder)
	if err != nil {
		return err
	}
	firstRound := state.Cursor == ""
	normalizeRevisionState(&state)

	deltaChanges, cursor, err := consumeDeltaRound(ctx, changes, folder, state.Cursor)
	if errors.Is(err, domain.ErrMailDeltaReset) {
		deltaChanges, cursor, err = consumeDeltaRound(ctx, changes, folder, "")
		if errors.Is(err, domain.ErrMailDeltaReset) {
			return domain.Service("mail delta reset failed")
		}
	}
	if err != nil {
		return err
	}

	candidates := make(map[string]domain.MailMessage)
	for _, change := range deltaChanges {
		if change.Removed {
			continue
		}
		if change.Message.ID == "" || change.Revision == "" {
			return domain.Service("invalid mail delta change")
		}
		if state.Revisions[change.Message.ID] == change.Revision {
			continue
		}
		touchRevision(&state, change.Message.ID, change.Revision)
		conversation := change.Message.Conversation
		if conversation == "" {
			conversation = "message:" + change.Message.ID
		}
		if previous, ok := candidates[conversation]; !ok || previous.ID == change.Message.ID || mailMessageLess(previous, change.Message) {
			candidates[conversation] = change.Message
		}
	}
	state.Cursor = cursor
	pruneRevisions(&state, domain.MaxMailWatchRevisions)

	if firstRound && !query.IncludeExisting {
		return states.Save(account, folder, state)
	}

	events := make([]domain.MailWatchEvent, 0, len(candidates))
	for _, message := range candidates {
		events = append(events, domain.MailWatchEvent{
			Event:          domain.MailChangedEvent,
			MessageID:      message.ID,
			ConversationID: message.Conversation,
			Received:       message.Received,
			Subject:        message.Subject,
			From:           message.From,
		})
	}
	sort.Slice(events, func(i, j int) bool {
		return eventLess(events[i], events[j])
	})
	for _, event := range events {
		if err := sink.Emit(ctx, event); err != nil {
			return err
		}
	}
	return states.Save(account, folder, state)
}

func consumeDeltaRound(ctx context.Context, changes ChangeStore, folder, initialToken string) ([]domain.MailDeltaChange, string, error) {
	token := initialToken
	seen := map[string]bool{}
	var collected []domain.MailDeltaChange
	for {
		if seen[token] {
			return nil, "", domain.Service("mail delta pagination loop")
		}
		seen[token] = true
		page, err := changes.Delta(ctx, DeltaQuery{Folder: folder, Token: token})
		if err != nil {
			return nil, "", err
		}
		collected = append(collected, page.Changes...)
		switch {
		case page.NextToken != "" && page.DeltaToken != "":
			return nil, "", domain.Service("invalid mail delta page")
		case page.NextToken != "":
			token = page.NextToken
		case page.DeltaToken != "":
			return collected, page.DeltaToken, nil
		default:
			return nil, "", domain.Service("mail delta round did not return a cursor")
		}
	}
}

func normalizeRevisionState(state *domain.MailWatchState) {
	if state.Revisions == nil {
		state.Revisions = map[string]string{}
	}
	seen := make(map[string]bool, len(state.Revisions))
	order := make([]string, 0, len(state.Revisions))
	for _, id := range state.RevisionOrder {
		if _, ok := state.Revisions[id]; ok && !seen[id] {
			seen[id] = true
			order = append(order, id)
		}
	}
	missing := make([]string, 0, len(state.Revisions)-len(order))
	for id := range state.Revisions {
		if !seen[id] {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	state.RevisionOrder = append(order, missing...)
}

func touchRevision(state *domain.MailWatchState, id, revision string) {
	state.Revisions[id] = revision
	for i, knownID := range state.RevisionOrder {
		if knownID == id {
			state.RevisionOrder = append(state.RevisionOrder[:i], state.RevisionOrder[i+1:]...)
			break
		}
	}
	state.RevisionOrder = append(state.RevisionOrder, id)
}

func pruneRevisions(state *domain.MailWatchState, limit int) {
	if limit < 0 {
		limit = 0
	}
	for len(state.RevisionOrder) > limit {
		oldest := state.RevisionOrder[0]
		state.RevisionOrder = state.RevisionOrder[1:]
		delete(state.Revisions, oldest)
	}
}

func eventLess(a, b domain.MailWatchEvent) bool {
	if receivedLess(a.Received, b.Received) {
		return true
	}
	if receivedLess(b.Received, a.Received) {
		return false
	}
	return a.MessageID < b.MessageID
}

func mailMessageLess(a, b domain.MailMessage) bool {
	if receivedLess(a.Received, b.Received) {
		return true
	}
	if receivedLess(b.Received, a.Received) {
		return false
	}
	return a.ID < b.ID
}

func receivedLess(a, b string) bool {
	at, aerr := time.Parse(time.RFC3339Nano, a)
	bt, berr := time.Parse(time.RFC3339Nano, b)
	if aerr == nil && berr == nil {
		return at.Before(bt)
	}
	return a < b
}
