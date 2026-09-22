package mail

import (
	"context"
	"errors"
	netmail "net/mail"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/masonhuemmer/m365/internal/app/auth"
	"github.com/masonhuemmer/m365/internal/domain"
)

func Watch(
	ctx context.Context,
	changes ChangeStore,
	threads ThreadStore,
	classifier ResponseClassifier,
	states WatchStateStore,
	sink EventSink,
	sess domain.Session,
	cfg WatchConfig,
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
	target := domain.ResponseTarget{}
	if query.Classify {
		if !cfg.ClassificationEnabled {
			return domain.Usage("experimental mail response classification is disabled")
		}
		if cfg.ActionableThreshold <= 0 || cfg.ActionableThreshold > 1 {
			return domain.Usage("mail response classification threshold must be greater than 0 and at most 1")
		}
		var err error
		target, err = normalizeResponseTarget(sess.Account, query.TargetAddresses, query.TargetNames)
		if err != nil {
			return err
		}
	}
	account := strings.ToLower(strings.TrimSpace(sess.Account))
	if account == "" {
		return domain.Usage("signed-in account is required for mail watch")
	}
	if changes == nil || states == nil || sink == nil {
		return domain.Usage("mail watch is unavailable")
	}
	if query.Classify && (threads == nil || classifier == nil) {
		return domain.Usage("mail response classification is unavailable")
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

	messages := make([]domain.MailMessage, 0, len(candidates))
	for _, message := range candidates {
		messages = append(messages, message)
	}
	sort.Slice(messages, func(i, j int) bool {
		return mailMessageLess(messages[i], messages[j])
	})

	events := make([]domain.MailWatchEvent, 0, len(messages))
	for _, message := range messages {
		event := domain.MailWatchEvent{
			Event:          domain.MailChangedEvent,
			MessageID:      message.ID,
			ConversationID: message.Conversation,
			Received:       message.Received,
			Subject:        message.Subject,
			From:           message.From,
		}
		if query.Classify {
			thread, err := threads.LatestThread(ctx, message.ID, domain.MaxClassificationMessages)
			if err != nil {
				return err
			}
			input := buildClassificationInput(target, thread)
			classification, err := classifier.Classify(ctx, input)
			if err != nil {
				return err
			}
			actionable := classification.Status == domain.ResponseWaitingOnTarget && classification.TargetProbability >= cfg.ActionableThreshold
			eventTarget := target
			eventClassification := classification
			event.Event = domain.MailResponseClassifiedEvent
			event.Target = &eventTarget
			event.Classification = &eventClassification
			event.Actionable = &actionable
		}
		events = append(events, event)
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

func normalizeResponseTarget(account string, addresses, names []string) (domain.ResponseTarget, error) {
	target := domain.ResponseTarget{Names: []string{}, Addresses: []string{}}
	seenAddresses := map[string]bool{}
	if normalized, ok := normalizedEmailAddress(account); ok {
		target.Addresses = append(target.Addresses, normalized)
		seenAddresses[normalized] = true
	}
	for _, address := range addresses {
		normalized, ok := normalizedEmailAddress(address)
		if !ok {
			return domain.ResponseTarget{}, domain.Usagef("invalid target address %q", strings.TrimSpace(address))
		}
		if !seenAddresses[normalized] {
			seenAddresses[normalized] = true
			target.Addresses = append(target.Addresses, normalized)
		}
	}
	seenNames := map[string]bool{}
	for _, name := range names {
		normalized := strings.TrimSpace(name)
		key := strings.ToLower(normalized)
		if normalized != "" && !seenNames[key] {
			seenNames[key] = true
			target.Names = append(target.Names, normalized)
		}
	}
	if len(target.Addresses) == 0 {
		return domain.ResponseTarget{}, domain.Usage("at least one target address is required for mail response classification")
	}
	return target, nil
}

func normalizedEmailAddress(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	parsed, err := netmail.ParseAddress(trimmed)
	if err != nil || !strings.EqualFold(parsed.Address, trimmed) {
		return "", false
	}
	return strings.ToLower(parsed.Address), true
}

func buildClassificationInput(target domain.ResponseTarget, thread domain.MailThread) domain.ClassificationInput {
	items := append([]domain.MailMessage(nil), thread.Items...)
	sort.Slice(items, func(i, j int) bool {
		return mailMessageLess(items[i], items[j])
	})
	truncated := thread.Count > len(items)
	if len(items) > domain.MaxClassificationMessages {
		items = items[len(items)-domain.MaxClassificationMessages:]
		truncated = true
	}
	messages := make([]domain.ClassificationMessage, len(items))
	remainingBodyBytes := domain.MaxClassificationBodyBytes
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		body := strings.TrimSpace(strings.ToValidUTF8(item.Body, "�"))
		if len(body) > remainingBodyBytes {
			body = truncateUTF8(body, remainingBodyBytes)
			truncated = true
		}
		remainingBodyBytes -= len(body)
		messages[i] = domain.ClassificationMessage{
			From:     normalizedParticipantAddress(item.From.Address),
			To:       normalizedParticipantAddresses(item.To),
			CC:       normalizedParticipantAddresses(item.CC),
			Received: item.Received,
			Subject:  strings.TrimSpace(strings.ToValidUTF8(item.Subject, "�")),
			BodyText: body,
		}
	}
	return domain.ClassificationInput{
		Target: target,
		Thread: messages,
		Truncation: domain.ClassificationTruncation{
			MessageLimit:  domain.MaxClassificationMessages,
			BodyByteLimit: domain.MaxClassificationBodyBytes,
			Truncated:     truncated,
		},
	}
}

func normalizedParticipantAddress(address string) string {
	return strings.ToLower(strings.TrimSpace(address))
}

func normalizedParticipantAddresses(people []domain.Person) []string {
	addresses := make([]string, 0, len(people))
	seen := map[string]bool{}
	for _, person := range people {
		address := normalizedParticipantAddress(person.Address)
		if address != "" && !seen[address] {
			seen[address] = true
			addresses = append(addresses, address)
		}
	}
	return addresses
}

func truncateUTF8(value string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(value) <= maxBytes {
		return value
	}
	for maxBytes > 0 && !utf8.ValidString(value[:maxBytes]) {
		maxBytes--
	}
	return value[:maxBytes]
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
