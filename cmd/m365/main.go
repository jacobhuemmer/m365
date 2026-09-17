package main

import (
	"os"
	"path/filepath"

	"github.com/masonhuemmer/m365/internal/adapters/cli"
	"github.com/masonhuemmer/m365/internal/adapters/fs"
	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/adapters/keychain"
	"github.com/masonhuemmer/m365/internal/adapters/watchstate"
	"github.com/masonhuemmer/m365/internal/config"
)

func main() {
	cfg, _ := config.Load()
	d := cli.Deps{Config: cfg, Write: fs.WriteFile, Watch: &watchstate.File{}}
	if os.Getenv("M365_FAKE") == "1" {
		mem := graph.Seed()
		d.Mail = graph.MailAPI{Memory: mem}
		d.Teams = graph.TeamsAPI{Memory: mem}
		d.Calendar = graph.CalendarAPI{Memory: mem}
		d.Files = graph.FilesAPI{Memory: mem}
		d.Login = graph.FakeLoginAll(true, true, true, true)
		home, _ := os.UserHomeDir()
		state := os.Getenv("XDG_STATE_HOME")
		if state == "" {
			state = filepath.Join(home, ".local", "state")
		}
		d.Store = &keychain.FileStore{Path: filepath.Join(state, "m365", "fake-session.json")}
	} else {
		store := &keychain.Fallback{
			Primary:   keychain.Keyring{},
			Secondary: &keychain.FileStore{Path: keychain.LiveSessionPath()},
		}
		ref := &graph.Refresher{
			Store:  store,
			Config: graph.PKCEConfig(cfg.ClientID, cfg.TenantID, ""),
		}
		httpc := &graph.HTTPClient{
			Base:    "https://graph.microsoft.com/v1.0",
			Refresh: ref.Token,
		}
		d.Store = store
		d.Mail = httpc
		d.Teams = &graph.HTTPTeams{HTTPClient: httpc}
		d.Calendar = &graph.HTTPCalendar{HTTPClient: httpc}
		d.Files = &graph.HTTPFiles{HTTPClient: httpc}
		d.Login = graph.RealLogin(cfg.ClientID, cfg.TenantID)
	}
	os.Exit(cli.Run(os.Args, d))
}
