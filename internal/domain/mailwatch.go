package domain

import "errors"

const (
	MailChangedEvent      = "mail.changed"
	MaxMailWatchRevisions = 5000
)

var ErrMailDeltaReset = errors.New("mail delta cursor reset required")

type MailDeltaChange struct {
	Message  MailMessage
	Revision string
	Removed  bool
}

type MailDeltaPage struct {
	Changes    []MailDeltaChange
	NextToken  string
	DeltaToken string
}

type MailWatchState struct {
	Cursor        string            `json:"cursor"`
	Revisions     map[string]string `json:"revisions"`
	RevisionOrder []string          `json:"revision_order"`
}

type MailWatchEvent struct {
	Event          string `json:"event"`
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Received       string `json:"received"`
	Subject        string `json:"subject"`
	From           Person `json:"from"`
}

func NewMailWatchState() MailWatchState {
	return MailWatchState{Revisions: map[string]string{}, RevisionOrder: []string{}}
}
