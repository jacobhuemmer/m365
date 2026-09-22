package mail

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

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

	err := Watch(context.Background(), changes, states, sink, domain.SignedOut(), WatchQuery{})
	if domain.ExitOf(err) != domain.ExitAuth {
		t.Fatalf("signed-out error = %v", err)
	}
	err = Watch(context.Background(), changes, states, sink, mailWatchSession(), WatchQuery{Folder: "ALL"})
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

	if err := Watch(context.Background(), changes, states, sink, mailWatchSession(), WatchQuery{}); err != nil {
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

	err := Watch(context.Background(), changes, states, sink, mailWatchSession(), WatchQuery{IncludeExisting: true})
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

	if err := Watch(context.Background(), changes, states, sink, mailWatchSession(), WatchQuery{}); err != nil {
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

	if err := Watch(context.Background(), changes, states, sink, mailWatchSession(), WatchQuery{}); err == nil {
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

	if err := Watch(context.Background(), changes, states, sink, mailWatchSession(), WatchQuery{}); err == nil {
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

	if err := Watch(context.Background(), changes, states, &eventSinkStub{}, mailWatchSession(), WatchQuery{}); err != nil {
		t.Fatal(err)
	}
	saved := states.saves[0]
	if len(saved.Revisions) != domain.MaxMailWatchRevisions || saved.Revisions["old-0000"] != "" || saved.Revisions["new"] != "rev-new" {
		t.Fatalf("pruned revisions = %d oldest=%q new=%q", len(saved.Revisions), saved.Revisions["old-0000"], saved.Revisions["new"])
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
