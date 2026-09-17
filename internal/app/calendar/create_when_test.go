package calendar

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestCreateWhenPhrase(t *testing.T) {
	now, loc := freeze()
	st := &stub{}
	in := WriteInput{Subject: "Sync", When: "tomorrow at 1:30 pm", Now: now, TZ: loc.String(), DryRun: true}
	id, err := Create(context.Background(), st, sessCal(), &in)
	if err != nil || id != "" || st.mut != 0 {
		t.Fatalf("%q %v mut=%d", id, err, st.mut)
	}
	if in.Start == "" || in.End == "" {
		t.Fatal("resolved times")
	}
	in2 := WriteInput{Subject: "Sync", When: "1:30", Now: now, TZ: loc.String()}
	_, err = Create(context.Background(), st, sessCal(), &in2)
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	in3 := WriteInput{Subject: "Sync", When: "today at 9am", Now: now, TZ: loc.String()}
	_, err = Create(context.Background(), st, sessCal(), &in3)
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	in4 := WriteInput{Subject: "Sync", When: "tomorrow at 1:30 pm", Start: "2026-09-17T13:30:00-05:00", Now: now, TZ: loc.String()}
	_, err = Create(context.Background(), st, sessCal(), &in4)
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal("when and start")
	}
}
