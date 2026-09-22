package mail

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/masonhuemmer/m365/internal/domain"
)

type deltaResult struct {
	page domain.MailDeltaPage
	err  error
}

type changeStoreStub struct {
	results map[string]deltaResult
	calls   []DeltaQuery
}

func (s *changeStoreStub) Delta(_ context.Context, q DeltaQuery) (domain.MailDeltaPage, error) {
	s.calls = append(s.calls, q)
	r, ok := s.results[q.Token]
	if !ok {
		return domain.MailDeltaPage{}, fmt.Errorf("unexpected delta token %q", q.Token)
	}
	return r.page, r.err
}

type watchStateStub struct {
	state   domain.MailWatchState
	loadErr error
	saveErr error
	loads   [][2]string
	saves   []domain.MailWatchState
}

func (s *watchStateStub) Load(account, folder string) (domain.MailWatchState, error) {
	s.loads = append(s.loads, [2]string{account, folder})
	return cloneWatchState(s.state), s.loadErr
}

func (s *watchStateStub) Save(account, folder string, state domain.MailWatchState) error {
	s.loads = append(s.loads, [2]string{account, folder})
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saves = append(s.saves, cloneWatchState(state))
	return nil
}

type eventSinkStub struct {
	events []domain.MailWatchEvent
	failAt int
}

type threadStoreStub struct {
	threads map[string]domain.MailThread
	err     error
	calls   []struct {
		id    string
		limit int
	}
}

func (s *threadStoreStub) LatestThread(_ context.Context, id string, limit int) (domain.MailThread, error) {
	s.calls = append(s.calls, struct {
		id    string
		limit int
	}{id: id, limit: limit})
	if s.err != nil {
		return domain.MailThread{}, s.err
	}
	return s.threads[id], nil
}

type classifierStub struct {
	result domain.ResponseClassification
	errAt  int
	inputs []domain.ClassificationInput
}

func (s *classifierStub) Classify(_ context.Context, input domain.ClassificationInput) (domain.ResponseClassification, error) {
	s.inputs = append(s.inputs, input)
	if s.errAt > 0 && len(s.inputs) == s.errAt {
		return domain.ResponseClassification{}, errors.New("classifier failed")
	}
	return s.result, nil
}

func (s *eventSinkStub) Emit(_ context.Context, event domain.MailWatchEvent) error {
	if s.failAt > 0 && len(s.events)+1 == s.failAt {
		return errors.New("output failed")
	}
	s.events = append(s.events, event)
	return nil
}

func mailWatchSession() domain.Session {
	return domain.Session{
		SignedIn:      true,
		SessionUsable: true,
		Account:       " User@Example.COM ",
		MailConsented: true,
	}
}

func deltaChange(id, conversation, received, revision string) domain.MailDeltaChange {
	return domain.MailDeltaChange{
		Message: domain.MailMessage{
			ID: id, Conversation: conversation, Received: received,
			Subject: "subject " + id,
			From:    domain.Person{Name: "Sender", Address: "sender@example.com"},
			Body:    "must not be emitted",
		},
		Revision: revision,
	}
}

func TestWatchRequiresConsentAndRejectsMailboxWideFolder(t *testing.T) {
	changes := &changeStoreStub{}
	states := &watchStateStub{}
	sink := &eventSinkStub{}

	err := Watch(context.Background(), changes, nil, nil, states, sink, domain.SignedOut(), WatchConfig{}, WatchQuery{})
	if domain.ExitOf(err) != domain.ExitAuth {
		t.Fatalf("signed-out error = %v", err)
	}
	err = Watch(context.Background(), changes, nil, nil, states, sink, mailWatchSession(), WatchConfig{}, WatchQuery{Folder: "ALL"})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatalf("all-folder error = %v", err)
	}
	if len(changes.calls) != 0 || len(states.loads) != 0 {
		t.Fatalf("dependencies called before validation: changes=%v states=%v", changes.calls, states.loads)
	}
}

func TestWatchFirstRoundConsumesEveryPageAndSavesQuietBaseline(t *testing.T) {
	changes := &changeStoreStub{results: map[string]deltaResult{
		"": {
			page: domain.MailDeltaPage{
				Changes:   []domain.MailDeltaChange{deltaChange("msg-1", "conv-1", "2026-01-01T00:00:00Z", "rev-1")},
				NextToken: "page-2",
			},
		},
		"page-2": {
			page: domain.MailDeltaPage{
				Changes:    []domain.MailDeltaChange{deltaChange("msg-2", "conv-1", "2026-01-01T01:00:00Z", "rev-2")},
				DeltaToken: "cursor-1",
			},
		},
	}}
	states := &watchStateStub{}
	sink := &eventSinkStub{}

	if err := Watch(context.Background(), changes, nil, nil, states, sink, mailWatchSession(), WatchConfig{}, WatchQuery{}); err != nil {
		t.Fatal(err)
	}
	if got := changes.calls; !reflect.DeepEqual(got, []DeltaQuery{{Folder: "inbox"}, {Folder: "inbox", Token: "page-2"}}) {
		t.Fatalf("delta calls = %+v", got)
	}
	if len(sink.events) != 0 {
		t.Fatalf("baseline emitted events: %+v", sink.events)
	}
	if len(states.saves) != 1 {
		t.Fatalf("save count = %d", len(states.saves))
	}
	saved := states.saves[0]
	if saved.Cursor != "cursor-1" || saved.Revisions["msg-1"] != "rev-1" || saved.Revisions["msg-2"] != "rev-2" {
		t.Fatalf("saved baseline = %+v", saved)
	}
	if got := states.loads[0]; got != [2]string{"user@example.com", "inbox"} {
		t.Fatalf("state key = %v", got)
	}
}

func TestWatchIncludesExistingOncePerConversationInStableOrder(t *testing.T) {
	olderRevision := deltaChange("msg-2", "conv-1", "2026-01-01T01:00:00Z", "rev-2a")
	olderRevision.Message.Subject = "older revision"
	newerRevision := deltaChange("msg-2", "conv-1", "2026-01-01T01:00:00Z", "rev-2b")
	newerRevision.Message.Subject = "newer revision"
	changes := &changeStoreStub{results: map[string]deltaResult{
		"": {page: domain.MailDeltaPage{
			Changes: []domain.MailDeltaChange{
				deltaChange("msg-3", "conv-2", "2026-01-01T02:00:00Z", "rev-3"),
				deltaChange("msg-1", "conv-1", "2026-01-01T00:00:00Z", "rev-1"),
				olderRevision,
				newerRevision,
				{Message: domain.MailMessage{ID: "deleted"}, Removed: true},
			},
			DeltaToken: "cursor-1",
		}},
	}}
	states := &watchStateStub{}
	sink := &eventSinkStub{}

	err := Watch(context.Background(), changes, nil, nil, states, sink, mailWatchSession(), WatchConfig{}, WatchQuery{IncludeExisting: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(sink.events) != 2 {
		t.Fatalf("events = %+v", sink.events)
	}
	if sink.events[0].MessageID != "msg-2" || sink.events[1].MessageID != "msg-3" {
		t.Fatalf("event order/grouping = %+v", sink.events)
	}
	if sink.events[0].Subject != "newer revision" || states.saves[0].Revisions["msg-2"] != "rev-2b" {
		t.Fatalf("latest repeated revision was not retained: event=%+v state=%+v", sink.events[0], states.saves[0])
	}
	for _, event := range sink.events {
		if event.Event != domain.MailChangedEvent {
			t.Fatalf("event name = %q", event.Event)
		}
	}
}

func TestWatchResetsOnceAndRetainsKnownRevisions(t *testing.T) {
	changes := &changeStoreStub{results: map[string]deltaResult{
		"expired": {err: domain.ErrMailDeltaReset},
		"": {page: domain.MailDeltaPage{
			Changes: []domain.MailDeltaChange{
				deltaChange("msg-1", "conv-1", "2026-01-01T00:00:00Z", "rev-1"),
				deltaChange("msg-2", "conv-2", "2026-01-01T01:00:00Z", "rev-2"),
			},
			DeltaToken: "fresh",
		}},
	}}
	states := &watchStateStub{state: domain.MailWatchState{
		Cursor:        "expired",
		Revisions:     map[string]string{"msg-1": "rev-1"},
		RevisionOrder: []string{"msg-1"},
	}}
	sink := &eventSinkStub{}

	if err := Watch(context.Background(), changes, nil, nil, states, sink, mailWatchSession(), WatchConfig{}, WatchQuery{}); err != nil {
		t.Fatal(err)
	}
	if got := changes.calls; !reflect.DeepEqual(got, []DeltaQuery{{Folder: "inbox", Token: "expired"}, {Folder: "inbox"}}) {
		t.Fatalf("delta calls = %+v", got)
	}
	if len(sink.events) != 1 || sink.events[0].MessageID != "msg-2" {
		t.Fatalf("reset events = %+v", sink.events)
	}
	if len(states.saves) != 1 || states.saves[0].Cursor != "fresh" {
		t.Fatalf("reset save = %+v", states.saves)
	}
}

func TestWatchDoesNotEmitOrSaveAfterIncompletePaging(t *testing.T) {
	changes := &changeStoreStub{results: map[string]deltaResult{
		"old": {page: domain.MailDeltaPage{
			Changes:   []domain.MailDeltaChange{deltaChange("msg-1", "conv-1", "2026-01-01T00:00:00Z", "rev-1")},
			NextToken: "page-2",
		}},
		"page-2": {err: errors.New("graph failed")},
	}}
	states := &watchStateStub{state: domain.MailWatchState{Cursor: "old", Revisions: map[string]string{}}}
	sink := &eventSinkStub{}

	if err := Watch(context.Background(), changes, nil, nil, states, sink, mailWatchSession(), WatchConfig{}, WatchQuery{}); err == nil {
		t.Fatal("expected paging error")
	}
	if len(sink.events) != 0 || len(states.saves) != 0 {
		t.Fatalf("partial round advanced: events=%+v saves=%+v", sink.events, states.saves)
	}
}

func TestWatchDoesNotSaveAfterOutputFailure(t *testing.T) {
	changes := &changeStoreStub{results: map[string]deltaResult{
		"old": {page: domain.MailDeltaPage{
			Changes: []domain.MailDeltaChange{
				deltaChange("msg-1", "conv-1", "2026-01-01T00:00:00Z", "rev-1"),
				deltaChange("msg-2", "conv-2", "2026-01-01T01:00:00Z", "rev-2"),
			},
			DeltaToken: "new",
		}},
	}}
	states := &watchStateStub{state: domain.MailWatchState{Cursor: "old", Revisions: map[string]string{}}}
	sink := &eventSinkStub{failAt: 2}

	if err := Watch(context.Background(), changes, nil, nil, states, sink, mailWatchSession(), WatchConfig{}, WatchQuery{}); err == nil {
		t.Fatal("expected output error")
	}
	if len(sink.events) != 1 || len(states.saves) != 0 {
		t.Fatalf("output failure advanced state: events=%+v saves=%+v", sink.events, states.saves)
	}
}

func TestWatchPrunesOldestRevisions(t *testing.T) {
	revisions := make(map[string]string, domain.MaxMailWatchRevisions)
	order := make([]string, 0, domain.MaxMailWatchRevisions)
	for i := 0; i < domain.MaxMailWatchRevisions; i++ {
		id := fmt.Sprintf("old-%04d", i)
		revisions[id] = "rev"
		order = append(order, id)
	}
	changes := &changeStoreStub{results: map[string]deltaResult{
		"old": {page: domain.MailDeltaPage{
			Changes:    []domain.MailDeltaChange{deltaChange("new", "conv-new", "2026-01-01T00:00:00Z", "rev-new")},
			DeltaToken: "new",
		}},
	}}
	states := &watchStateStub{state: domain.MailWatchState{Cursor: "old", Revisions: revisions, RevisionOrder: order}}

	if err := Watch(context.Background(), changes, nil, nil, states, &eventSinkStub{}, mailWatchSession(), WatchConfig{}, WatchQuery{}); err != nil {
		t.Fatal(err)
	}
	saved := states.saves[0]
	if len(saved.Revisions) != domain.MaxMailWatchRevisions || saved.Revisions["old-0000"] != "" || saved.Revisions["new"] != "rev-new" {
		t.Fatalf("pruned revisions = %d oldest=%q new=%q", len(saved.Revisions), saved.Revisions["old-0000"], saved.Revisions["new"])
	}
}

func TestWatchClassificationGateAndAuthoritativeTargetFailBeforePolling(t *testing.T) {
	changes := &changeStoreStub{}
	states := &watchStateStub{}
	threads := &threadStoreStub{}
	classifier := &classifierStub{}
	query := WatchQuery{Classify: true, TargetAddresses: []string{"user@example.com"}}

	err := Watch(context.Background(), changes, threads, classifier, states, &eventSinkStub{}, mailWatchSession(), WatchConfig{}, query)
	if domain.ExitOf(err) != domain.ExitUsage || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("disabled error = %v", err)
	}
	if len(changes.calls) != 0 || len(states.loads) != 0 || len(threads.calls) != 0 || len(classifier.inputs) != 0 {
		t.Fatalf("disabled classification used dependencies: changes=%v states=%v threads=%v classifier=%v", changes.calls, states.loads, threads.calls, classifier.inputs)
	}

	session := mailWatchSession()
	session.Account = "signed-in"
	err = Watch(context.Background(), changes, threads, classifier, states, &eventSinkStub{}, session, WatchConfig{ClassificationEnabled: true, ActionableThreshold: 0.8}, WatchQuery{Classify: true})
	if domain.ExitOf(err) != domain.ExitUsage || !strings.Contains(err.Error(), "target address") {
		t.Fatalf("missing target error = %v", err)
	}
	if len(changes.calls) != 0 || len(states.loads) != 0 || len(threads.calls) != 0 || len(classifier.inputs) != 0 {
		t.Fatalf("missing target used dependencies: changes=%v states=%v threads=%v classifier=%v", changes.calls, states.loads, threads.calls, classifier.inputs)
	}
}

func TestWatchClassifiesNormalizedTargetAndDerivesActionable(t *testing.T) {
	changes := &changeStoreStub{results: map[string]deltaResult{
		"": {page: domain.MailDeltaPage{
			Changes:    []domain.MailDeltaChange{deltaChange("msg-1", "conv-1", "2026-01-01T00:00:00Z", "rev-1")},
			DeltaToken: "cursor-1",
		}},
	}}
	thread := domain.MailThread{ConversationID: "conv-1", Count: 1, Items: []domain.MailMessage{{
		ID: "msg-1", Conversation: "conv-1", Subject: "Approval", Received: "2026-01-01T00:00:00Z",
		From: domain.Person{Name: "Alice", Address: " Alice@Example.com "},
		To:   []domain.Person{{Name: "User", Address: "USER@example.com"}},
		CC:   []domain.Person{{Name: "Ops", Address: "OPS@example.com"}},
		Body: " Please approve. ", Attachments: []domain.Attachment{{ID: "secret-attachment"}},
	}}}
	threads := &threadStoreStub{threads: map[string]domain.MailThread{"msg-1": thread}}
	confidence := 0.88
	classifier := &classifierStub{result: classificationResult(domain.ResponseWaitingOnTarget, 0.8, confidence)}
	states := &watchStateStub{}
	sink := &eventSinkStub{}
	query := WatchQuery{
		Classify:        true,
		IncludeExisting: true,
		TargetAddresses: []string{" USER@example.com ", "alias@example.com", "ALIAS@example.com"},
		TargetNames:     []string{" Mason Huemmer ", "mason huemmer"},
	}

	err := Watch(context.Background(), changes, threads, classifier, states, sink, mailWatchSession(), WatchConfig{ClassificationEnabled: true, ActionableThreshold: 0.8}, query)
	if err != nil {
		t.Fatal(err)
	}
	if len(threads.calls) != 1 || threads.calls[0].id != "msg-1" || threads.calls[0].limit != domain.MaxClassificationMessages {
		t.Fatalf("thread calls = %+v", threads.calls)
	}
	if len(classifier.inputs) != 1 {
		t.Fatalf("classifier inputs = %+v", classifier.inputs)
	}
	input := classifier.inputs[0]
	if !reflect.DeepEqual(input.Target.Addresses, []string{"user@example.com", "alias@example.com"}) || !reflect.DeepEqual(input.Target.Names, []string{"Mason Huemmer"}) {
		t.Fatalf("normalized target = %+v", input.Target)
	}
	if len(input.Thread) != 1 || input.Thread[0].From != "alice@example.com" || !reflect.DeepEqual(input.Thread[0].To, []string{"user@example.com"}) || !reflect.DeepEqual(input.Thread[0].CC, []string{"ops@example.com"}) || input.Thread[0].BodyText != "Please approve." {
		t.Fatalf("classification thread = %+v", input.Thread)
	}
	if len(sink.events) != 1 || sink.events[0].Event != domain.MailResponseClassifiedEvent || sink.events[0].Target == nil || sink.events[0].Classification == nil || sink.events[0].Actionable == nil || !*sink.events[0].Actionable {
		t.Fatalf("classified event = %+v", sink.events)
	}
	if len(states.saves) != 1 || states.saves[0].Cursor != "cursor-1" {
		t.Fatalf("classification save = %+v", states.saves)
	}
}

func TestWatchBoundsLatestThreadMessagesAndUTF8BodyBytes(t *testing.T) {
	changes := &changeStoreStub{results: map[string]deltaResult{
		"": {page: domain.MailDeltaPage{
			Changes:    []domain.MailDeltaChange{deltaChange("changed", "conv-1", "2026-01-01T12:00:00Z", "rev-1")},
			DeltaToken: "cursor-1",
		}},
	}}
	items := make([]domain.MailMessage, 12)
	for i := range items {
		items[i] = domain.MailMessage{
			ID: fmt.Sprintf("msg-%02d", i), Conversation: "conv-1",
			Received: fmt.Sprintf("2026-01-01T%02d:00:00Z", i), Body: strings.Repeat("é", 4000),
		}
	}
	threads := &threadStoreStub{threads: map[string]domain.MailThread{"changed": {ConversationID: "conv-1", Count: len(items), Items: items}}}
	classifier := &classifierStub{result: classificationResult(domain.ResponseWaitingOnOther, 0.1, 0.75)}

	err := Watch(context.Background(), changes, threads, classifier, &watchStateStub{}, &eventSinkStub{}, mailWatchSession(), WatchConfig{ClassificationEnabled: true, ActionableThreshold: 0.8}, WatchQuery{Classify: true, IncludeExisting: true})
	if err != nil {
		t.Fatal(err)
	}
	input := classifier.inputs[0]
	if len(input.Thread) != domain.MaxClassificationMessages {
		t.Fatalf("thread message count = %d", len(input.Thread))
	}
	if input.Thread[0].Received != items[2].Received || input.Thread[len(input.Thread)-1].Received != items[11].Received {
		t.Fatalf("thread window = first %q last %q", input.Thread[0].Received, input.Thread[len(input.Thread)-1].Received)
	}
	totalBytes := 0
	for _, message := range input.Thread {
		if !utf8.ValidString(message.BodyText) {
			t.Fatalf("invalid UTF-8 body: %q", message.BodyText)
		}
		totalBytes += len(message.BodyText)
	}
	if totalBytes > domain.MaxClassificationBodyBytes || !input.Truncation.Truncated {
		t.Fatalf("body budget = %d truncation=%+v", totalBytes, input.Truncation)
	}
}

func TestWatchClassifierFailureEmitsNothingAndPreservesCheckpoint(t *testing.T) {
	changes := &changeStoreStub{results: map[string]deltaResult{
		"old": {page: domain.MailDeltaPage{
			Changes: []domain.MailDeltaChange{
				deltaChange("msg-1", "conv-1", "2026-01-01T00:00:00Z", "rev-1"),
				deltaChange("msg-2", "conv-2", "2026-01-01T01:00:00Z", "rev-2"),
			},
			DeltaToken: "new",
		}},
	}}
	threads := &threadStoreStub{threads: map[string]domain.MailThread{
		"msg-1": {ConversationID: "conv-1", Items: []domain.MailMessage{{ID: "msg-1", Body: "one"}}},
		"msg-2": {ConversationID: "conv-2", Items: []domain.MailMessage{{ID: "msg-2", Body: "two"}}},
	}}
	classifier := &classifierStub{result: classificationResult(domain.ResponseWaitingOnTarget, 0.9, 0.8), errAt: 2}
	states := &watchStateStub{state: domain.MailWatchState{Cursor: "old", Revisions: map[string]string{}}}
	sink := &eventSinkStub{}

	err := Watch(context.Background(), changes, threads, classifier, states, sink, mailWatchSession(), WatchConfig{ClassificationEnabled: true, ActionableThreshold: 0.8}, WatchQuery{Classify: true})
	if err == nil {
		t.Fatal("expected classifier error")
	}
	if len(classifier.inputs) != 2 || len(sink.events) != 0 || len(states.saves) != 0 {
		t.Fatalf("classifier failure advanced delivery: inputs=%d events=%+v saves=%+v", len(classifier.inputs), sink.events, states.saves)
	}
}

func classificationResult(status domain.ResponseStatus, targetProbability, confidence float64) domain.ResponseClassification {
	return domain.ResponseClassification{
		Status:            status,
		TargetProbability: targetProbability,
		Probabilities: map[domain.ResponseStatus]float64{
			domain.ResponseWaitingOnTarget:    targetProbability,
			domain.ResponseWaitingOnOther:     0.05,
			domain.ResponseNoResponseExpected: 0.03,
			domain.ResponseUnclear:            0.02,
		},
		Confidence: &confidence,
		Model:      "fake-1",
	}
}

func cloneWatchState(state domain.MailWatchState) domain.MailWatchState {
	copyState := domain.MailWatchState{
		Cursor:        state.Cursor,
		Revisions:     make(map[string]string, len(state.Revisions)),
		RevisionOrder: append([]string(nil), state.RevisionOrder...),
	}
	for id, revision := range state.Revisions {
		copyState.Revisions[id] = revision
	}
	return copyState
}
