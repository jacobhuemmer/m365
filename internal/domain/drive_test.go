package domain

import "testing"

func TestValidateUpload(t *testing.T) {
	if err := ValidateUpload(OutboundFile{Name: "ok.txt", Size: 12}); err != nil {
		t.Fatal(err)
	}
	if ExitOf(ValidateUpload(OutboundFile{Name: "z", Size: 0})) != ExitUsage {
		t.Fatal("zero")
	}
	if ExitOf(ValidateUpload(OutboundFile{Name: "big", Size: MaxUploadBytes + 1})) != ExitUsage {
		t.Fatal("oversize")
	}
}
