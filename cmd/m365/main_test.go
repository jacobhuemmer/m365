package main

import (
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/jev"
	"github.com/masonhuemmer/m365/internal/adapters/mailclassifier"
	"github.com/masonhuemmer/m365/internal/config"
	"github.com/masonhuemmer/m365/internal/domain"
)

func TestBuildDepsConstructsClassifierOnlyWhenEnabled(t *testing.T) {
	enabled := config.Config{Experimental: config.ExperimentalConfig{
		MailResponseClassification: config.MailResponseClassification{
			Enabled: true, Provider: "jev", Model: "jev-latest", ActionableThreshold: 0.8,
		},
	}}
	disabled := enabled
	disabled.Experimental.MailResponseClassification.Enabled = false

	t.Run("disabled ignores present key", func(t *testing.T) {
		d := buildDeps(disabled, nil, testEnvironment(map[string]string{"TYPESAFE_API_KEY": "present"}))
		if d.MailClassifier != nil || d.MailClassifierError != nil {
			t.Fatalf("classifier=%T error=%v", d.MailClassifier, d.MailClassifierError)
		}
	})

	t.Run("enabled requires key", func(t *testing.T) {
		d := buildDeps(enabled, nil, testEnvironment(nil))
		if d.MailClassifier != nil || domain.ExitOf(d.MailClassifierError) != domain.ExitUsage {
			t.Fatalf("classifier=%T error=%v", d.MailClassifier, d.MailClassifierError)
		}
	})

	t.Run("enabled constructs live Jev client", func(t *testing.T) {
		d := buildDeps(enabled, nil, testEnvironment(map[string]string{"TYPESAFE_API_KEY": "present"}))
		if _, ok := d.MailClassifier.(*jev.Client); !ok || d.MailClassifierError != nil {
			t.Fatalf("classifier=%T error=%v", d.MailClassifier, d.MailClassifierError)
		}
	})

	t.Run("fake mode keeps deterministic classifier without key", func(t *testing.T) {
		d := buildDeps(enabled, nil, testEnvironment(map[string]string{"M365_FAKE": "1"}))
		if _, ok := d.MailClassifier.(*mailclassifier.Fake); !ok || d.MailClassifierError != nil {
			t.Fatalf("classifier=%T error=%v", d.MailClassifier, d.MailClassifierError)
		}
	})
}

func testEnvironment(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}
