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
	"github.com/masonhuemmer/m365/internal/adapters/mailclassifier"
	"github.com/masonhuemmer/m365/internal/adapters/mailwatchstate"
	"github.com/masonhuemmer/m365/internal/config"
	"github.com/masonhuemmer/m365/internal/domain"
)

type mailWatchHarness struct {
	deps  cli.Deps
	state *mailwatchstate.Memory
	out   *bytes.Buffer
	err   *bytes.Buffer
}

var mailWatchHarnesses sync.Map

func init() {
	runtime.Register("a signed-in fake mailbox for mail watch", prepareMailWatch)
	runtime.Register("a signed-in fake mailbox with response classification enabled above the fake probability", prepareMailWatchClassificationAboveProbability)
	runtime.Register("a signed-in fake mailbox with response classification enabled", prepareMailWatchClassification)
	runtime.Register("I poll mail changes including existing messages", pollMailWatchIncludingExisting)
	runtime.Register("I poll mail changes again", pollMailWatch)
	runtime.Register("I poll mail changes", pollMailWatch)
	runtime.Register("I classify existing mail for", classifyExistingMail)
	runtime.Register("I classify mail for", classifyMail)
	runtime.Register("the mail watch command succeeds", mailWatchSucceeds)
	runtime.Register("the mail watch command is rejected as usage", mailWatchRejectedAsUsage)
	runtime.Register("no mail change events are emitted", noMailWatchEvents)
	runtime.Register("one body-free mail change event is emitted per changed conversation", bodyFreeMailWatchEvents)
	runtime.Register("one privacy-safe response classification event is emitted per changed conversation", privacySafeClassificationEvents)
	runtime.Register("response actionability follows the configured threshold", classificationIsActionable)
	runtime.Register("every response classification event is not actionable", classificationIsNotActionable)
	runtime.Register("the mail watch checkpoint is unchanged", mailWatchCheckpointUnchanged)
}

func prepareMailWatch(world *runtime.World, _ string) error {
	memory := graph.Seed()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	state := &mailwatchstate.Memory{}
	deps := cli.Deps{
		Config:         config.Config{ClientID: "x", TenantID: "y"},
		Store:          &keychain.Fake{},
		Mail:           graph.MailAPI{Memory: memory},
		MailChanges:    graph.MailDeltaAPI{Memory: memory},
		MailThreads:    graph.MailDeltaAPI{Memory: memory},
		MailClassifier: &mailclassifier.Fake{},
		MailWatchState: state,
		Login:          graph.FakeLoginAll(true, true, true, true),
		Stdout:         out,
		Stderr:         errOut,
	}
	if code := cli.Run([]string{"m365", "auth", "login"}, deps); code != domain.ExitOK {
		return fmt.Errorf("fake login exited %d: %s", code, errOut.String())
	}
	out.Reset()
	errOut.Reset()
	mailWatchHarnesses.Store(world, &mailWatchHarness{deps: deps, state: state, out: out, err: errOut})
	return nil
}

func prepareMailWatchClassification(world *runtime.World, text string) error {
	if err := prepareMailWatch(world, text); err != nil {
		return err
	}
	harness, err := loadMailWatchHarness(world)
	if err != nil {
		return err
	}
	harness.deps.Config.Experimental.MailResponseClassification = config.MailResponseClassification{
		Enabled: true, Provider: "jev", Model: "jev-latest", ActionableThreshold: 0.8,
	}
	return nil
}

func prepareMailWatchClassificationAboveProbability(world *runtime.World, text string) error {
	if err := prepareMailWatchClassification(world, text); err != nil {
		return err
	}
	harness, err := loadMailWatchHarness(world)
	if err != nil {
		return err
	}
	harness.deps.Config.Experimental.MailResponseClassification.ActionableThreshold = 0.95
	return nil
}

func pollMailWatch(world *runtime.World, _ string) error {
	return runMailWatch(world, false)
}

func pollMailWatchIncludingExisting(world *runtime.World, _ string) error {
	return runMailWatch(world, true)
}

func runMailWatch(world *runtime.World, includeExisting bool) error {
	harness, err := loadMailWatchHarness(world)
	if err != nil {
		return err
	}
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

func classifyExistingMail(world *runtime.World, text string) error {
	return runClassifiedMailWatch(world, targetAddress(text), true)
}

func classifyMail(world *runtime.World, text string) error {
	return runClassifiedMailWatch(world, targetAddress(text), false)
}

func runClassifiedMailWatch(world *runtime.World, address string, includeExisting bool) error {
	harness, err := loadMailWatchHarness(world)
	if err != nil {
		return err
	}
	harness.out.Reset()
	harness.err.Reset()
	args := []string{"m365", "mail", "watch", "--classify", "--target-address", address, "--target-name", "Test User"}
	if includeExisting {
		args = append(args, "--include-existing")
	}
	world.Code = cli.Run(args, harness.deps)
	world.Out = harness.out.String()
	world.Err = harness.err.String()
	return nil
}

func targetAddress(text string) string {
	start := strings.Index(text, `"`)
	end := strings.LastIndex(text, `"`)
	if start >= 0 && end > start {
		return text[start+1 : end]
	}
	return ""
}

func loadMailWatchHarness(world *runtime.World) (*mailWatchHarness, error) {
	value, ok := mailWatchHarnesses.Load(world)
	if !ok {
		return nil, fmt.Errorf("mail watch harness is not initialized")
	}
	return value.(*mailWatchHarness), nil
}

func mailWatchSucceeds(world *runtime.World, _ string) error {
	if world.Code != domain.ExitOK {
		return fmt.Errorf("mail watch exited %d: %s", world.Code, world.Err)
	}
	return nil
}

func mailWatchRejectedAsUsage(world *runtime.World, _ string) error {
	if world.Code != domain.ExitUsage {
		return fmt.Errorf("mail watch exited %d instead of usage: %s", world.Code, world.Err)
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

func privacySafeClassificationEvents(world *runtime.World, _ string) error {
	events, err := decodedMailWatchEvents(world.Out)
	if err != nil {
		return err
	}
	if len(events) != 1 {
		return fmt.Errorf("expected one classified conversation, got %d lines: %s", len(events), world.Out)
	}
	event := events[0]
	if event["event"] != domain.MailResponseClassifiedEvent {
		return fmt.Errorf("unexpected classified event: %+v", event)
	}
	classification, _ := event["classification"].(map[string]any)
	probabilities, _ := classification["probabilities"].(map[string]any)
	if len(probabilities) != 4 {
		return fmt.Errorf("expected all response probabilities: %+v", classification)
	}
	for _, forbidden := range []string{"body", "attachments", "cursor", "token", "revision"} {
		if containsMapKey(event, forbidden) {
			return fmt.Errorf("classification event exposes %s: %+v", forbidden, event)
		}
	}
	return nil
}

func classificationIsActionable(world *runtime.World, _ string) error {
	events, err := decodedMailWatchEvents(world.Out)
	if err != nil {
		return err
	}
	for _, event := range events {
		if event["actionable"] != true {
			return fmt.Errorf("expected actionable classification: %+v", event)
		}
	}
	return nil
}

func classificationIsNotActionable(world *runtime.World, _ string) error {
	events, err := decodedMailWatchEvents(world.Out)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return fmt.Errorf("expected classified events")
	}
	for _, event := range events {
		if event["actionable"] != false {
			return fmt.Errorf("expected non-actionable classification: %+v", event)
		}
	}
	return nil
}

func mailWatchCheckpointUnchanged(world *runtime.World, _ string) error {
	harness, err := loadMailWatchHarness(world)
	if err != nil {
		return err
	}
	state, err := harness.state.Load("user@example.com", "inbox")
	if err != nil {
		return err
	}
	if state.Cursor != "" || len(state.Revisions) != 0 {
		return fmt.Errorf("checkpoint changed: %+v", state)
	}
	return nil
}

func decodedMailWatchEvents(output string) ([]map[string]any, error) {
	if strings.TrimSpace(output) == "" {
		return nil, nil
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	events := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func containsMapKey(value any, key string) bool {
	switch typed := value.(type) {
	case map[string]any:
		for current, child := range typed {
			if current == key || containsMapKey(child, key) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsMapKey(child, key) {
				return true
			}
		}
	}
	return false
}
