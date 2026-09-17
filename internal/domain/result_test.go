package domain

import "testing"

func TestNormalizeTop(t *testing.T) {
	n, err := NormalizeTop(0, DefaultMailTop)
	if err != nil || n != 10 {
		t.Fatalf("mail default: %d %v", n, err)
	}
	n, err = NormalizeTop(0, DefaultTeamsTop)
	if err != nil || n != 20 {
		t.Fatalf("teams default: %d %v", n, err)
	}
	n, err = NormalizeTop(0, DefaultCalendarTop)
	if err != nil || n != 10 {
		t.Fatalf("calendar events default: %d %v", n, err)
	}
	n, err = NormalizeTop(0, DefaultCalendarsTop)
	if err != nil || n != 20 {
		t.Fatalf("calendars default: %d %v", n, err)
	}
	n, err = NormalizeTop(0, DefaultFilesTop)
	if err != nil || n != 20 {
		t.Fatalf("files default: %d %v", n, err)
	}
	n, err = NormalizeTop(50, 10)
	if err != nil || n != 50 {
		t.Fatalf("max: %d %v", n, err)
	}
	_, err = NormalizeTop(51, 10)
	if ExitOf(err) != ExitUsage {
		t.Fatalf("51 should be usage, got %v", err)
	}
	_, err = NormalizeTop(-1, 10)
	if ExitOf(err) != ExitUsage {
		t.Fatalf("-1 should be usage")
	}
}
