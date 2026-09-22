package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestLoadEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("M365_CLIENT_ID", "")
	t.Setenv("M365_TENANT_ID", "")
	cfgDir := filepath.Join(dir, "m365")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"client_id":"fromfile","tenant_id":"tfile"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ClientID != "fromfile" || c.TenantID != "tfile" {
		t.Fatalf("file: %+v", c)
	}
	t.Setenv("M365_CLIENT_ID", "fromenv")
	c, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ClientID != "fromenv" || c.TenantID != "tfile" {
		t.Fatalf("override: %+v", c)
	}
}

func TestRequireApp(t *testing.T) {
	err := Config{}.RequireApp()
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatalf("got %v", err)
	}
	if err := (Config{ClientID: "a", TenantID: "b"}).RequireApp(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadMailResponseClassificationDefaultsDisabled(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("M365_CLIENT_ID", "")
	t.Setenv("M365_TENANT_ID", "")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	got := c.Experimental.MailResponseClassification
	if got.Enabled || got.Provider != "jev" || got.Model != "jev-latest" || got.ActionableThreshold != 0.8 {
		t.Fatalf("classification defaults = %+v", got)
	}
}

func TestLoadMailResponseClassificationOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("M365_CLIENT_ID", "")
	t.Setenv("M365_TENANT_ID", "")
	cfgDir := filepath.Join(dir, "m365")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	payload := `{
		"experimental": {
			"mail_response_classification": {
				"enabled": true,
				"provider": "jev",
				"model": "jev-preview",
				"actionable_threshold": 0.65
			}
		}
	}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	got := c.Experimental.MailResponseClassification
	if !got.Enabled || got.Provider != "jev" || got.Model != "jev-preview" || got.ActionableThreshold != 0.65 {
		t.Fatalf("classification config = %+v", got)
	}
}

func TestLoadRejectsInvalidMailResponseClassification(t *testing.T) {
	tests := []struct {
		name    string
		setting string
	}{
		{name: "provider", setting: `"provider":"other"`},
		{name: "zero threshold", setting: `"actionable_threshold":0`},
		{name: "high threshold", setting: `"actionable_threshold":1.01`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", dir)
			t.Setenv("M365_CLIENT_ID", "")
			t.Setenv("M365_TENANT_ID", "")
			cfgDir := filepath.Join(dir, "m365")
			if err := os.MkdirAll(cfgDir, 0o700); err != nil {
				t.Fatal(err)
			}
			payload := `{"experimental":{"mail_response_classification":{` + tt.setting + `}}}`
			if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(payload), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(); domain.ExitOf(err) != domain.ExitUsage {
				t.Fatalf("invalid config error = %v", err)
			}
		})
	}
}
