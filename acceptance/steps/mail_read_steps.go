package steps

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/masonhuemmer/m365/acceptance/runtime"
	"github.com/masonhuemmer/m365/internal/domain"
)

func init() {
	runtime.Register("the message includes participants, received time, and read state", func(w *runtime.World, _ string) error {
		var message map[string]any
		if err := json.Unmarshal([]byte(w.Out), &message); err != nil {
			w.T.Fatalf("invalid message JSON: %v", err)
		}
		if message["from"] == nil || message["to"] == nil || message["cc"] == nil || message["received"] == nil || message["is_read"] == nil {
			w.T.Fatalf("incomplete message metadata: %s", w.Out)
		}
		return nil
	})
	runtime.Register("the message includes attachment metadata without attachment bytes", func(w *runtime.World, _ string) error {
		var message domain.MailMessage
		if err := json.Unmarshal([]byte(w.Out), &message); err != nil {
			w.T.Fatalf("invalid message JSON: %v", err)
		}
		if len(message.Attachments) == 0 || message.Attachments[0].Name == "" || message.Attachments[0].ContentType == "" {
			w.T.Fatalf("missing attachment metadata: %s", w.Out)
		}
		if strings.Contains(w.Out, "synthetic-ok") {
			w.T.Fatalf("attachment bytes leaked: %s", w.Out)
		}
		return nil
	})
	runtime.Register("the complete conversation is ordered oldest-first with message ID tie-breakers", func(w *runtime.World, _ string) error {
		var thread domain.MailThread
		if err := json.Unmarshal([]byte(w.Out), &thread); err != nil {
			w.T.Fatalf("invalid thread JSON: %v", err)
		}
		if thread.Count != len(thread.Items) || thread.Count < 2 {
			w.T.Fatalf("incomplete thread: %s", w.Out)
		}
		for i := 1; i < len(thread.Items); i++ {
			previous, current := thread.Items[i-1], thread.Items[i]
			previousTime, previousErr := time.Parse(time.RFC3339Nano, previous.Received)
			currentTime, currentErr := time.Parse(time.RFC3339Nano, current.Received)
			if previousErr != nil || currentErr != nil || previousTime.After(currentTime) || (previousTime.Equal(currentTime) && previous.ID > current.ID) {
				w.T.Fatalf("thread is not oldest-first: %s", w.Out)
			}
		}
		return nil
	})
	runtime.Register("every message retains its body", func(w *runtime.World, _ string) error {
		var thread domain.MailThread
		if err := json.Unmarshal([]byte(w.Out), &thread); err != nil {
			w.T.Fatalf("invalid thread JSON: %v", err)
		}
		for _, message := range thread.Items {
			if message.Body == "" {
				w.T.Fatalf("thread body missing: %s", w.Out)
			}
		}
		return nil
	})
}
