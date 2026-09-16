package graph

import "testing"

func TestSaveCommentPackage(t *testing.T) {
	if Seed() == nil {
		t.Fatal("seed")
	}
}
