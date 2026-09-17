package graph

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/masonhuemmer/m365/internal/domain"
)

func isCalendarPath(p string) bool {
	return strings.Contains(p, "/calendars") || strings.Contains(p, "/calendar") || strings.Contains(p, "/events")
}

func isFilesPath(p string) bool {
	return strings.Contains(p, "/drive") || strings.Contains(p, "/sites")
}

func denyWorkload(tok, p string) bool {
	if strings.Contains(p, "/users/") && p != "/me" {
		return true
	}
	if strings.Contains(p, "/sites") {
		return true
	}
	cal, files := isCalendarPath(p), isFilesPath(p)
	switch tok {
	case "fake-both", "fake-mail", "fake-teams":
		return cal || files
	case "fake-calendar":
		return files || strings.Contains(p, "/me/messages") || strings.Contains(p, "/chats")
	case "fake-files":
		return cal || strings.Contains(p, "/me/messages") || strings.Contains(p, "/chats")
	case "fake-all":
		return false
	default:
		return cal || files
	}
}

func handleCalendarFiles(w http.ResponseWriter, r *http.Request, mem *Memory) bool {
	p := r.URL.Path
	q := r.URL.Query()
	if p == "/me/calendars" {
		vals := []map[string]any{}
		for _, c := range mem.Calendars {
			tz := c.Timezone
			if tz == "" {
				tz = "America/Chicago"
			}
			vals = append(vals, map[string]any{"id": c.ID, "name": c.Name, "isDefaultCalendar": c.IsDefault, "timeZone": tz})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"value": vals})
		return true
	}
	if strings.Contains(p, "calendarView") {
		start, end := q.Get("startDateTime"), q.Get("endDateTime")
		cal := "cal-1"
		if strings.Contains(p, "/me/calendars/") {
			rest := strings.TrimPrefix(p, "/me/calendars/")
			cal = strings.Split(rest, "/")[0]
		}
		vals := []map[string]any{}
		for _, e := range mem.CalEvents {
			if e.CalendarID != cal {
				continue
			}
			if start != "" && e.Start < start {
				continue
			}
			if end != "" && e.Start >= end {
				continue
			}
			vals = append(vals, graphEvent(e))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"value": vals})
		return true
	}
	if strings.HasPrefix(p, "/me/events/") {
		id := strings.TrimPrefix(p, "/me/events/")
		if i := strings.IndexByte(id, '/'); i >= 0 {
			id = id[:i]
		}
		if r.Method == http.MethodDelete {
			_ = mem.DeleteEvent(r.Context(), id)
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		if r.Method == http.MethodPatch {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
			return true
		}
		for _, e := range mem.CalEvents {
			if e.ID == id {
				_ = json.NewEncoder(w).Encode(graphEvent(e))
				return true
			}
		}
		http.Error(w, `{"error":{"code":"itemNotFound"}}`, http.StatusNotFound)
		return true
	}
	if r.Method == http.MethodPost && strings.HasSuffix(p, "/events") {
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "ev-new"})
		return true
	}
	if p == "/me/drive" || p == "/me/drive/root" {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": mem.Drive.ID, "name": mem.Drive.Name, "folder": map[string]any{}})
		return true
	}
	if strings.HasSuffix(p, "/children") && strings.Contains(p, "/drive") {
		parent := "root"
		if strings.Contains(p, "/items/") {
			rest := strings.TrimPrefix(p[strings.Index(p, "/items/"):], "/items/")
			parent = strings.TrimSuffix(rest, "/children")
		}
		vals := []map[string]any{}
		for _, it := range mem.Items {
			if it.ParentID == parent {
				vals = append(vals, graphItem(it))
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"value": vals, "@odata.nextLink": "https://graph.microsoft.com/v1.0/me/drive/root/children?$skip=20"})
		return true
	}
	if strings.Contains(p, "/drive/items/") {
		id := p[strings.Index(p, "/items/")+len("/items/"):]
		if i := strings.IndexByte(id, '/'); i >= 0 {
			rest := id[i:]
			id = id[:i]
			if strings.HasSuffix(rest, "/content") {
				w.Header().Set("Content-Type", "application/octet-stream")
				_, _ = w.Write(mem.FileBytes[id])
				return true
			}
		}
		if r.Method == http.MethodDelete {
			_ = mem.DeleteItem(r.Context(), id)
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		for _, it := range mem.Items {
			if it.ID == id {
				_ = json.NewEncoder(w).Encode(graphItem(it))
				return true
			}
		}
		http.Error(w, `{"error":{"code":"itemNotFound"}}`, http.StatusNotFound)
		return true
	}
	if r.Method == http.MethodPut && strings.Contains(p, "/content") {
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "file-new", "name": "up.txt"})
		return true
	}
	if strings.Contains(p, "createUploadSession") {
		_ = json.NewEncoder(w).Encode(map[string]any{"uploadUrl": "http://127.0.0.1/upload"})
		return true
	}
	return false
}

func graphEvent(e domain.CalendarEvent) map[string]any {
	return map[string]any{
		"id": e.ID, "subject": e.Subject, "start": map[string]string{"dateTime": e.Start, "timeZone": "UTC"},
		"end":       map[string]string{"dateTime": e.End, "timeZone": "UTC"},
		"location":  map[string]string{"displayName": e.Location},
		"organizer": map[string]any{"emailAddress": map[string]string{"address": e.Organizer.Address, "name": e.Organizer.Name}},
		"body":      map[string]string{"content": e.Body},
	}
}

func graphItem(it domain.DriveItem) map[string]any {
	m := map[string]any{"id": it.ID, "name": it.Name, "size": it.Size, "lastModifiedDateTime": it.LastModified, "webUrl": it.WebURL}
	if it.IsFolder {
		m["folder"] = map[string]any{}
	} else {
		m["file"] = map[string]any{}
	}
	return m
}
