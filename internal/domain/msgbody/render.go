package msgbody

// Mode is how the caller wrote the body.
type Mode int

const (
	Plain Mode = iota
	Markdown
	HTML
)

// Target is where the body is delivered. It decides paragraph layout.
type Target int

const (
	Mail Target = iota
	Teams
)

// Rendered is the exact body a send delivers.
type Rendered struct {
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
}

// Render turns the caller's body into the HTML a send delivers.
func Render(mode Mode, target Target, src string) Rendered {
	return Rendered{}
}
