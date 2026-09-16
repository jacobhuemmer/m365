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
