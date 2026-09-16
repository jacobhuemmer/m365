package domain

import "testing"

func TestValidateOutbound(t *testing.T) {
	if err := ValidateOutbound(nil); err != nil {
		t.Fatal(err)
	}
	ok := []OutboundFile{{Path: "a", Name: "a", Size: 1}}
	if err := ValidateOutbound(ok); err != nil {
		t.Fatal(err)
	}
	if err := ValidateOutbound([]OutboundFile{{Name: "z", Size: 0}}); ExitOf(err) != ExitUsage {
		t.Fatalf("empty: %v", err)
	}
	if err := ValidateOutbound([]OutboundFile{{Name: "z", Size: MaxAttachBytes + 1}}); ExitOf(err) != ExitUsage {
		t.Fatalf("oversize: %v", err)
	}
	too := make([]OutboundFile, MaxAttachCount+1)
	for i := range too {
		too[i] = OutboundFile{Name: "x", Size: 1}
	}
	if err := ValidateOutbound(too); ExitOf(err) != ExitUsage {
		t.Fatalf("count: %v", err)
	}
}
