package domain

const (
	MaxAttachBytes = 10 * 1024 * 1024
	MaxAttachCount = 10
)

type Attachment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type,omitempty"`
}

type OutboundFile struct {
	Path string
	Name string
	Size int64
}

func ValidateOutbound(files []OutboundFile) error {
	if len(files) == 0 {
		return nil
	}
	if len(files) > MaxAttachCount {
		return Usagef("too many attachments: %d (max %d)", len(files), MaxAttachCount)
	}
	for _, f := range files {
		if f.Size <= 0 {
			return Usagef("empty file rejected: %s", f.Name)
		}
		if f.Size > MaxAttachBytes {
			return Usagef("file exceeds 10 MiB: %s", f.Name)
		}
	}
	return nil
}
