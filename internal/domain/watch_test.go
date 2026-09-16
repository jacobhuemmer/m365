package domain

import "testing"

func TestNewCheckpoint(t *testing.T) {
	cp := NewCheckpoint()
	if cp.Marks == nil {
		t.Fatal("marks")
	}
}
