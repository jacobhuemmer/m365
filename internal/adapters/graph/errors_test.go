package graph

import (
	"net/http"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestMapStatus(t *testing.T) {
	if domain.ExitOf(MapStatus(http.StatusForbidden)) != domain.ExitAuth {
		t.Fatal("403")
	}
	if domain.ExitOf(MapStatus(http.StatusNotFound)) != domain.ExitNotFound {
		t.Fatal("404")
	}
	if domain.ExitOf(MapStatus(http.StatusTooManyRequests)) != domain.ExitService {
		t.Fatal("429")
	}
}

func TestMapGraphErrorInvalidID(t *testing.T) {
	err := MapGraphError(http.StatusBadRequest, []byte(`{"error":{"code":"ErrorInvalidId"}}`))
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatalf("got %v", err)
	}
	err = MapGraphError(http.StatusNotFound, []byte(`{}`))
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatal(err)
	}
}
