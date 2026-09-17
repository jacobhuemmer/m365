package domain

const MaxUploadBytes int64 = 100 * 1024 * 1024

type DriveRoot struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DriveItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	IsFolder     bool   `json:"is_folder"`
	LastModified string `json:"last_modified,omitempty"`
	ParentID     string `json:"parent_id,omitempty"`
	WebURL       string `json:"web_url,omitempty"`
}

type DrivePage struct {
	Limit    int         `json:"limit"`
	Count    int         `json:"count"`
	NextPage *string     `json:"next_page"`
	Items    []DriveItem `json:"items"`
}

func ValidateUpload(f OutboundFile) error {
	if f.Size <= 0 {
		return Usagef("empty file rejected: %s", f.Name)
	}
	if f.Size > MaxUploadBytes {
		return Usagef("file exceeds 100 MiB: %s", f.Name)
	}
	return nil
}
