package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/masonhuemmer/m365/internal/domain"
)

type Config struct {
	ClientID string
	TenantID string
}

type fileShape struct {
	ClientID string `json:"client_id"`
	TenantID string `json:"tenant_id"`
}

func Load() (Config, error) {
	c := Config{
		ClientID: os.Getenv("M365_CLIENT_ID"),
		TenantID: os.Getenv("M365_TENANT_ID"),
	}
	path := filepath.Join(configDir(), "m365", "config.json")
	b, err := os.ReadFile(path) // #nosec G304 -- fixed XDG config.json, not request input
	if err == nil {
		var f fileShape
		if jerr := json.Unmarshal(b, &f); jerr != nil {
			return Config{}, domain.Usage("invalid config.json")
		}
		if c.ClientID == "" {
			c.ClientID = f.ClientID
		}
		if c.TenantID == "" {
			c.TenantID = f.TenantID
		}
	}
	return c, nil
}

func (c Config) RequireApp() error {
	if c.ClientID == "" || c.TenantID == "" {
		return domain.Usage("missing client id or tenant id")
	}
	return nil
}

func configDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}
