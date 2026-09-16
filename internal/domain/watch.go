package domain

type WatchEvent struct {
	ChatID      string       `json:"chat_id"`
	MessageID   string       `json:"message_id"`
	From        string       `json:"from,omitempty"`
	Text        string       `json:"text,omitempty"`
	Created     string       `json:"created"`
	Reason      string       `json:"reason"`
	Topic       string       `json:"topic,omitempty"`
	ChatType    string       `json:"chat_type,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type WatchMark struct {
	MessageID string `json:"message_id"`
	Created   string `json:"created"`
}

type WatchCheckpoint struct {
	Marks map[string]WatchMark `json:"marks"`
}

func NewCheckpoint() WatchCheckpoint {
	return WatchCheckpoint{Marks: map[string]WatchMark{}}
}
