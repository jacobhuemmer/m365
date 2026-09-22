package domain

import "errors"

const (
	MailChangedEvent            = "mail.changed"
	MailResponseClassifiedEvent = "mail.response_classified"
	MaxMailWatchRevisions       = 5000
	MaxClassificationMessages   = 10
	MaxClassificationBodyBytes  = 64 * 1024
	ResponseWaitingOnTarget     = ResponseStatus("waiting_on_target")
	ResponseWaitingOnOther      = ResponseStatus("waiting_on_other")
	ResponseNoResponseExpected  = ResponseStatus("no_response_expected")
	ResponseUnclear             = ResponseStatus("unclear")
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
	Event          string                  `json:"event"`
	MessageID      string                  `json:"message_id"`
	ConversationID string                  `json:"conversation_id"`
	Received       string                  `json:"received"`
	Subject        string                  `json:"subject"`
	From           Person                  `json:"from"`
	Target         *ResponseTarget         `json:"target,omitempty"`
	Classification *ResponseClassification `json:"classification,omitempty"`
	Actionable     *bool                   `json:"actionable,omitempty"`
}

type ResponseStatus string

type ResponseTarget struct {
	Names     []string `json:"names"`
	Addresses []string `json:"addresses"`
}

type ClassificationMessage struct {
	From     string   `json:"from"`
	To       []string `json:"to"`
	CC       []string `json:"cc,omitempty"`
	Received string   `json:"received"`
	Subject  string   `json:"subject"`
	BodyText string   `json:"body_text"`
}

type ClassificationTruncation struct {
	MessageLimit  int  `json:"message_limit"`
	BodyByteLimit int  `json:"body_byte_limit"`
	Truncated     bool `json:"truncated"`
}

type ClassificationInput struct {
	Target     ResponseTarget           `json:"target"`
	Thread     []ClassificationMessage  `json:"thread"`
	Truncation ClassificationTruncation `json:"truncation"`
}

type ResponseClassification struct {
	Status            ResponseStatus             `json:"status"`
	TargetProbability float64                    `json:"target_probability"`
	Probabilities     map[ResponseStatus]float64 `json:"probabilities"`
	Confidence        *float64                   `json:"confidence,omitempty"`
	Model             string                     `json:"model"`
}

func NewMailWatchState() MailWatchState {
	return MailWatchState{Revisions: map[string]string{}, RevisionOrder: []string{}}
}
