package scout

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	schemascout "github.com/davidahmann/gait/core/schema/v1/scout"
)

func TestAppendLoadAndReportAdoptionEvents(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "adoption.jsonl")
	base := time.Date(2026, time.February, 6, 10, 0, 0, 0, time.UTC)

	events := []schemascout.AdoptionEvent{
		NewAdoptionEvent("demo", 0, 120*time.Millisecond, "test", base),
		NewAdoptionEvent("verify", 0, 130*time.Millisecond, "test", base.Add(time.Second)),
		NewAdoptionEvent("regress init", 0, 140*time.Millisecond, "test", base.Add(2*time.Second)),
		NewAdoptionEvent("regress run", 0, 150*time.Millisecond, "test", base.Add(3*time.Second)),
	}
	for _, event := range events {
		if err := AppendAdoptionEvent(logPath, event); err != nil {
			t.Fatalf("append adoption event: %v", err)
		}
	}

	loaded, err := LoadAdoptionEvents(logPath)
	if err != nil {
		t.Fatalf("load adoption events: %v", err)
	}
	if len(loaded) != len(events) {
		t.Fatalf("expected %d events got %d", len(events), len(loaded))
	}

	report := BuildAdoptionReport(loaded, logPath, "test", time.Time{})
	if report.SchemaID != "gait.doctor.adoption_report" {
		t.Fatalf("unexpected report schema id: %s", report.SchemaID)
	}
	if !report.ActivationComplete {
		t.Fatalf("expected activation complete")
	}
	if len(report.Blockers) != 0 {
		t.Fatalf("expected no blockers, got %#v", report.Blockers)
	}
	if !report.CreatedAt.Equal(base.Add(3 * time.Second)) {
		t.Fatalf("expected deterministic created_at from last event, got %s", report.CreatedAt.Format(time.RFC3339Nano))
	}
}

func TestAdoptionValidationAndBlockers(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "adoption.jsonl")
	event := NewAdoptionEvent("demo", 0, 10*time.Millisecond, "test", time.Date(2026, time.February, 6, 11, 0, 0, 0, time.UTC))
	event.ExitCode = 300
	if err := AppendAdoptionEvent(logPath, event); err == nil {
		t.Fatalf("expected append error for invalid exit code")
	}

	invalidPath := filepath.Join(t.TempDir(), "invalid.jsonl")
	if err := os.WriteFile(invalidPath, []byte("{\n"), 0o600); err != nil {
		t.Fatalf("write invalid jsonl: %v", err)
	}
	if _, err := LoadAdoptionEvents(invalidPath); err == nil {
		t.Fatalf("expected load error for malformed jsonl")
	}

	report := BuildAdoptionReport([]schemascout.AdoptionEvent{
		NewAdoptionEvent("demo", 0, 10*time.Millisecond, "test", time.Date(2026, time.February, 6, 11, 0, 0, 0, time.UTC)),
	}, "inline", "test", time.Time{})
	if report.ActivationComplete {
		t.Fatalf("expected incomplete activation")
	}
	if len(report.Blockers) == 0 {
		t.Fatalf("expected blockers for missing milestones")
	}
}

func TestNormalizeAdoptionEventValidation(t *testing.T) {
	base := schemascout.AdoptionEvent{
		SchemaID:        "gait.scout.adoption_event",
		SchemaVersion:   "1.0.0",
		CreatedAt:       time.Date(2026, time.February, 6, 16, 0, 0, 0, time.UTC),
		ProducerVersion: "test",
		Command:         "demo",
		Success:         true,
		ExitCode:        0,
		ElapsedMS:       12,
		Environment: schemascout.AdoptionEnvContext{
			OS:   "darwin",
			Arch: "arm64",
		},
	}

	validated, err := normalizeAdoptionEvent(base)
	if err != nil {
		t.Fatalf("normalize valid event: %v", err)
	}
	if validated.SchemaID != "gait.scout.adoption_event" || validated.SchemaVersion != "1.0.0" {
		t.Fatalf("unexpected normalized schema: %#v", validated)
	}

	cases := []schemascout.AdoptionEvent{
		func() schemascout.AdoptionEvent {
			event := base
			event.SchemaID = "bad"
			return event
		}(),
		func() schemascout.AdoptionEvent {
			event := base
			event.SchemaVersion = "2.0.0"
			return event
		}(),
		func() schemascout.AdoptionEvent {
			event := base
			event.CreatedAt = time.Time{}
			return event
		}(),
		func() schemascout.AdoptionEvent {
			event := base
			event.Command = ""
			return event
		}(),
		func() schemascout.AdoptionEvent {
			event := base
			event.ExitCode = -1
			return event
		}(),
		func() schemascout.AdoptionEvent {
			event := base
			event.ExitCode = 256
			return event
		}(),
		func() schemascout.AdoptionEvent {
			event := base
			event.Environment.OS = ""
			return event
		}(),
		func() schemascout.AdoptionEvent {
			event := base
			event.Environment.Arch = ""
			return event
		}(),
	}
	for _, testCase := range cases {
		if _, err := normalizeAdoptionEvent(testCase); err == nil {
			t.Fatalf("expected normalize error for case: %#v", testCase)
		}
	}
}

func TestAdoptionHelpersAndDefaults(t *testing.T) {
	event := NewAdoptionEvent("  ", 0, -10*time.Millisecond, "", time.Time{})
	if event.Command != "unknown" {
		t.Fatalf("expected unknown command default, got %s", event.Command)
	}
	if event.ProducerVersion != "0.0.0-dev" {
		t.Fatalf("expected producer default, got %s", event.ProducerVersion)
	}
	if event.ElapsedMS != 0 {
		t.Fatalf("expected elapsed clamp to 0, got %d", event.ElapsedMS)
	}
	if event.CreatedAt.IsZero() {
		t.Fatalf("expected created_at to be set")
	}

	if tags := milestoneTags("demo", true); len(tags) != 1 || tags[0] != "A1" {
		t.Fatalf("unexpected milestone tags for demo: %#v", tags)
	}
	if tags := milestoneTags("verify", false); len(tags) != 0 {
		t.Fatalf("expected no tags on failure: %#v", tags)
	}
	if !strings.Contains(activationBlockerHint("A3"), "regress init") {
		t.Fatalf("unexpected blocker hint for A3")
	}
	if !strings.Contains(activationBlockerHint("unknown"), "unknown") {
		t.Fatalf("expected fallback blocker hint")
	}

	emptyReport := BuildAdoptionReport(nil, "source", "test", time.Time{})
	if !emptyReport.CreatedAt.Equal(fixedAdoptionTime) {
		t.Fatalf("expected fixed created_at, got %s", emptyReport.CreatedAt.Format(time.RFC3339Nano))
	}
}
