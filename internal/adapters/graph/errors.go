package graph

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/masonhuemmer/m365/internal/domain"
)

func MapStatus(code int) error {
	switch code {
	case http.StatusUnauthorized, http.StatusForbidden:
		return domain.Auth("missing consent or session")
	case http.StatusNotFound:
		return domain.NotFound("not found")
	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return domain.Service("microsoft 365 service error")
	default:
		return domain.Service("microsoft 365 service error")
	}
}

func MapGraphError(code int, body []byte) error {
	if code == http.StatusNotFound {
		return domain.NotFound("not found")
	}
	var ge struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &ge)
	c := strings.ToLower(ge.Error.Code)
	if strings.Contains(c, "notfound") || strings.Contains(c, "invalidid") || strings.Contains(c, "itemnotfound") {
		return domain.NotFound("not found")
	}
	return MapStatus(code)
}
