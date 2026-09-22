package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/masonhuemmer/m365/acceptance/runtime"
	"github.com/masonhuemmer/m365/internal/adapters/cli"
	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/adapters/keychain"
	"github.com/masonhuemmer/m365/internal/adapters/mailwatchstate"
	"github.com/masonhuemmer/m365/internal/config"
	"github.com/masonhuemmer/m365/internal/domain"
)

type mailWatchHarness struct {
	deps cli.Deps
	out  *bytes.Buffer
	err  *bytes.Buffer
}

var mailWatchHarnesses sync.Map

func init() {
	runtime.Register("a signed-in fake mailbox for mail watch", prepareMailWatch)
	runtime.Register("I poll mail changes including existing messages", pollMailWatchIncludingExisting)
	runtime.Register("I poll mail changes again", pollMailWatch)
	runtime.Register("I poll mail changes", pollMailWatch)
	runtime.Register("the mail watch command succeeds", mailWatchSucceeds)
	runtime.Register("no mail change events are emitted", noMailWatchEvents)
	runtime.Register("one body-free mail change event is emitted per changed conversation", bodyFreeMailWatchEvents)
}

func prepareMailWatch(world *runtime.World, _ string) error {
	memory := graph.Seed()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	deps := cli.Deps{
		Config:         config.Config{ClientID: "x", TenantID: "y"},
		Store:          &keychain.Fake{},
		Mail:           graph.MailAPI{Memory: memory},
		MailChanges:    graph.MailDeltaAPI{Memory: memory},
		MailWatchState: &mailwatchstate.Memory{},
		Login:          graph.FakeLoginAll(true, true, true, true),
		Stdout:         out,
		Stderr:         errOut,
	}
	if code := cli.Run([]string{"m365", "auth", "login"}, deps); code != domain.ExitOK {
		return fmt.Errorf("fake login exited %d: %s", code, errOut.String())
	}
	out.Reset()
	errOut.Reset()
	mailWatchHarnesses.Store(world, &mailWatchHarness{deps: deps, out: out, err: errOut})
	return nil
}

func pollMailWatch(world *runtime.World, _ string) error {
	return runMailWatch(world, false)
}

func pollMailWatchIncludingExisting(world *runtime.World, _ string) error {
	return runMailWatch(world, true)
}

func runMailWatch(world *runtime.World, includeExisting bool) error {
	value, ok := mailWatchHarnesses.Load(world)
	if !ok {
		return fmt.Errorf("mail watch harness is not initialized")
	}
	harness := value.(*mailWatchHarness)
	harness.out.Reset()
	harness.err.Reset()
	args := []string{"m365", "mail", "watch"}
	if includeExisting {
		args = append(args, "--include-existing")
	}
	world.Code = cli.Run(args, harness.deps)
	world.Out = harness.out.String()
	world.Err = harness.err.String()
	return nil
}

func mailWatchSucceeds(world *runtime.World, _ string) error {
	if world.Code != domain.ExitOK {
		return fmt.Errorf("mail watch exited %d: %s", world.Code, world.Err)
	}
	return nil
}

func noMailWatchEvents(world *runtime.World, _ string) error {
	if strings.TrimSpace(world.Out) != "" {
		return fmt.Errorf("expected no events, got %s", world.Out)
	}
	return nil
}

func bodyFreeMailWatchEvents(world *runtime.World, _ string) error {
	lines := strings.Split(strings.TrimSpace(world.Out), "\n")
	if len(lines) != 1 {
		return fmt.Errorf("expected one changed conversation, got %d lines: %s", len(lines), world.Out)
	}
	var event map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		return err
	}
	if event["event"] != domain.MailChangedEvent {
		return fmt.Errorf("unexpected event: %+v", event)
	}
	for _, forbidden := range []string{"body", "attachments", "cursor", "token", "revision"} {
		if _, ok := event[forbidden]; ok {
			return fmt.Errorf("event exposes %s: %+v", forbidden, event)
		}
	}
	return nil
}
