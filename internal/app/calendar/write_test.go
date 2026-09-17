package calendar

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestDryRunDoesNotMutate(t *testing.T) {
	st := &stub{}
	in := WriteInput{Subject: "t", Start: "2026-09-16T10:00:00Z", End: "2026-09-16T11:00:00Z", DryRun: true}
	id, err := Create(context.Background(), st, sessCal(), &in)
	if err != nil || id != "" || st.mut != 0 {
		t.Fatalf("dry create %q %v mut=%d", id, err, st.mut)
	}
	if err := Update(context.Background(), st, sessCal(), WriteInput{ID: "ev-1", Subject: "x", DryRun: true}); err != nil || st.mut != 0 {
		t.Fatal(err, st.mut)
	}
	if err := Delete(context.Background(), st, sessCal(), "ev-1", true); err != nil || st.mut != 0 {
		t.Fatal(err, st.mut)
	}
}

func TestCreateValidation(t *testing.T) {
	st := &stub{}
	_, err := Create(context.Background(), st, sessCal(), &WriteInput{Start: "2026-09-16T10:00:00Z", End: "2026-09-16T11:00:00Z"})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	_, err = Create(context.Background(), st, sessCal(), &WriteInput{Subject: "t", Start: "2026-09-16T11:00:00Z", End: "2026-09-16T10:00:00Z"})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}
