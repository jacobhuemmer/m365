package domain

type Person struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
}

type MailMessage struct {
	ID             string       `json:"id"`
	Conversation   string       `json:"conversation_id,omitempty"`
	Subject        string       `json:"subject,omitempty"`
	From           Person       `json:"from"`
	To             []Person     `json:"to,omitempty"`
	CC             []Person     `json:"cc,omitempty"`
	Received       string       `json:"received,omitempty"`
	IsRead         bool         `json:"is_read"`
	HasAttachments bool         `json:"has_attachments"`
	Body           string       `json:"body,omitempty"`
	Attachments    []Attachment `json:"attachments,omitempty"`
}

type MailThread struct {
	ConversationID string        `json:"conversation_id"`
	Count          int           `json:"count"`
	Items          []MailMessage `json:"items"`
}

type MailPage struct {
	Limit    int           `json:"limit"`
	Count    int           `json:"count"`
	NextPage *string       `json:"next_page"`
	Items    []MailMessage `json:"items"`
}
