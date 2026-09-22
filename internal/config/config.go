package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/masonhuemmer/m365/internal/domain"
)

type Config struct {
	ClientID     string
	TenantID     string
	Experimental ExperimentalConfig
}

type ExperimentalConfig struct {
	MailResponseClassification MailResponseClassification `json:"mail_response_classification"`
}

type MailResponseClassification struct {
	Enabled             bool    `json:"enabled"`
	Provider            string  `json:"provider"`
	Model               string  `json:"model"`
	ActionableThreshold float64 `json:"actionable_threshold"`
}

type fileShape struct {
	ClientID     string                 `json:"client_id"`
	TenantID     string                 `json:"tenant_id"`
	Experimental fileExperimentalConfig `json:"experimental"`
}

type fileExperimentalConfig struct {
	MailResponseClassification fileMailResponseClassification `json:"mail_response_classification"`
}

type fileMailResponseClassification struct {
	Enabled             bool     `json:"enabled"`
	Provider            string   `json:"provider"`
	Model               string   `json:"model"`
	ActionableThreshold *float64 `json:"actionable_threshold"`
}

func Load() (Config, error) {
	c := Config{
		ClientID: os.Getenv("M365_CLIENT_ID"),
		TenantID: os.Getenv("M365_TENANT_ID"),
		Experimental: ExperimentalConfig{MailResponseClassification: MailResponseClassification{
			Provider: "jev", Model: "jev-latest", ActionableThreshold: 0.8,
		}},
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
		fileClassification := f.Experimental.MailResponseClassification
		classification := &c.Experimental.MailResponseClassification
		classification.Enabled = fileClassification.Enabled
		if provider := strings.TrimSpace(fileClassification.Provider); provider != "" {
			classification.Provider = provider
		}
		if model := strings.TrimSpace(fileClassification.Model); model != "" {
			classification.Model = model
		}
		if fileClassification.ActionableThreshold != nil {
			classification.ActionableThreshold = *fileClassification.ActionableThreshold
		}
	}
	if err := c.validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}

func (c Config) validate() error {
	classification := c.Experimental.MailResponseClassification
	if classification.Provider != "jev" {
		return domain.Usage("unsupported mail response classification provider")
	}
	if classification.ActionableThreshold <= 0 || classification.ActionableThreshold > 1 {
		return domain.Usage("mail response classification threshold must be greater than 0 and at most 1")
	}
	return nil
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
