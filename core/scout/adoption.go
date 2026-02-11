package scout

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/davidahmann/gait/core/fsx"
	schemascout "github.com/davidahmann/gait/core/schema/v1/scout"
)

const (
	adoptionEventSchemaID = "gait.scout.adoption_event"
	adoptionEventSchemaV1 = "1.0.0"
	maxAdoptionLineBytes  = 1024 * 1024
)

var (
	adoptionMilestoneOrder = []string{"A1", "A2", "A3", "A4", "E1", "E2", "E3"}
	activationMilestones   = []string{"A1", "A2", "A3", "A4"}
	fixedAdoptionTime      = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
)

type AdoptionCommandStats struct {
	Command string `json:"command"`
	Total   int    `json:"total"`
	Success int    `json:"success"`
	Failure int    `json:"failure"`
}

type AdoptionMilestoneStatus struct {
	Name     string `json:"name"`
	Achieved bool   `json:"achieved"`
}

type AdoptionReport struct {
	SchemaID           string                    `json:"schema_id"`
	SchemaVersion      string                    `json:"schema_version"`
	CreatedAt          time.Time                 `json:"created_at"`
	ProducerVersion    string                    `json:"producer_version"`
	Source             string                    `json:"source"`
	TotalEvents        int                       `json:"total_events"`
	SuccessEvents      int                       `json:"success_events"`
	FailedEvents       int                       `json:"failed_events"`
	FirstEventAt       time.Time                 `json:"first_event_at,omitempty"`
	LastEventAt        time.Time                 `json:"last_event_at,omitempty"`
	Commands           []AdoptionCommandStats    `json:"commands"`
	Milestones         []AdoptionMilestoneStatus `json:"milestones"`
	ActivationComplete bool                      `json:"activation_complete"`
	Blockers           []string                  `json:"blockers,omitempty"`
}

func NewAdoptionEvent(
	command string,
	exitCode int,
	elapsed time.Duration,
	producerVersion string,
	now time.Time,
) schemascout.AdoptionEvent {
	createdAt := now.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	if elapsed < 0 {
		elapsed = 0
	}
	trimmedCommand := strings.TrimSpace(command)
	if trimmedCommand == "" {
		trimmedCommand = "unknown"
	}
	trimmedProducerVersion := strings.TrimSpace(producerVersion)
	if trimmedProducerVersion == "" {
		trimmedProducerVersion = "0.0.0-dev"
	}
	success := exitCode == 0
	return schemascout.AdoptionEvent{
		SchemaID:        adoptionEventSchemaID,
		SchemaVersion:   adoptionEventSchemaV1,
		CreatedAt:       createdAt,
		ProducerVersion: trimmedProducerVersion,
		Command:         trimmedCommand,
		Success:         success,
		ExitCode:        exitCode,
		ElapsedMS:       elapsed.Milliseconds(),
		Milestones:      milestoneTags(trimmedCommand, success),
		Environment: schemascout.AdoptionEnvContext{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	}
}

func AppendAdoptionEvent(path string, event schemascout.AdoptionEvent) error {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return fmt.Errorf("adoption log path is required")
	}
	normalized, err := normalizeAdoptionEvent(event)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("marshal adoption event: %w", err)
	}
	if err := fsx.AppendLineLocked(trimmedPath, encoded, 0o600); err != nil {
		return fmt.Errorf("append adoption log: %w", err)
	}
	return nil
}

func LoadAdoptionEvents(path string) ([]schemascout.AdoptionEvent, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return nil, fmt.Errorf("adoption log path is required")
	}
	// #nosec G304 -- adoption log path is explicit local user input.
	file, err := os.Open(trimmedPath)
	if err != nil {
		return nil, fmt.Errorf("open adoption log: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	events := make([]schemascout.AdoptionEvent, 0)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), maxAdoptionLineBytes)
	line := 0
	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		var event schemascout.AdoptionEvent
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			return nil, fmt.Errorf("parse adoption log line %d: %w", line, err)
		}
		normalized, err := normalizeAdoptionEvent(event)
		if err != nil {
			return nil, fmt.Errorf("validate adoption log line %d: %w", line, err)
		}
		events = append(events, normalized)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan adoption log: %w", err)
	}
	return events, nil
}

func BuildAdoptionReport(
	events []schemascout.AdoptionEvent,
	source string,
	producerVersion string,
	now time.Time,
) AdoptionReport {
	createdAt := now.UTC()
	if createdAt.IsZero() {
		createdAt = fixedAdoptionTime
	}
	trimmedProducerVersion := strings.TrimSpace(producerVersion)
	if trimmedProducerVersion == "" {
		trimmedProducerVersion = "0.0.0-dev"
	}
	statsByCommand := map[string]AdoptionCommandStats{}
	milestonesAchieved := map[string]struct{}{}
	totalSuccess := 0
	totalFailed := 0
	firstEventAt := time.Time{}
	lastEventAt := time.Time{}

	for _, event := range events {
		command := strings.TrimSpace(event.Command)
		if command == "" {
			command = "unknown"
		}
		stats := statsByCommand[command]
		stats.Command = command
		stats.Total++
		if event.Success {
			stats.Success++
			totalSuccess++
		} else {
			stats.Failure++
			totalFailed++
		}
		statsByCommand[command] = stats

		for _, tag := range event.Milestones {
			milestonesAchieved[strings.TrimSpace(tag)] = struct{}{}
		}
		if firstEventAt.IsZero() || event.CreatedAt.Before(firstEventAt) {
			firstEventAt = event.CreatedAt
		}
		if lastEventAt.IsZero() || event.CreatedAt.After(lastEventAt) {
			lastEventAt = event.CreatedAt
		}
	}

	commandStats := make([]AdoptionCommandStats, 0, len(statsByCommand))
	for _, stats := range statsByCommand {
		commandStats = append(commandStats, stats)
	}
	sort.Slice(commandStats, func(i, j int) bool {
		return commandStats[i].Command < commandStats[j].Command
	})

	milestones := make([]AdoptionMilestoneStatus, 0, len(adoptionMilestoneOrder))
	for _, name := range adoptionMilestoneOrder {
		_, achieved := milestonesAchieved[name]
		milestones = append(milestones, AdoptionMilestoneStatus{
			Name:     name,
			Achieved: achieved,
		})
	}

	activationComplete := true
	blockers := make([]string, 0)
	for _, milestone := range activationMilestones {
		if _, ok := milestonesAchieved[milestone]; ok {
			continue
		}
		activationComplete = false
		blockers = append(blockers, activationBlockerHint(milestone))
	}
	blockers = uniqueSorted(blockers)

	if !lastEventAt.IsZero() {
		createdAt = lastEventAt.UTC()
	}

	return AdoptionReport{
		SchemaID:           "gait.doctor.adoption_report",
		SchemaVersion:      "1.0.0",
		CreatedAt:          createdAt,
		ProducerVersion:    trimmedProducerVersion,
		Source:             strings.TrimSpace(source),
		TotalEvents:        len(events),
		SuccessEvents:      totalSuccess,
		FailedEvents:       totalFailed,
		FirstEventAt:       firstEventAt,
		LastEventAt:        lastEventAt,
		Commands:           commandStats,
		Milestones:         milestones,
		ActivationComplete: activationComplete,
		Blockers:           blockers,
	}
}

func normalizeAdoptionEvent(event schemascout.AdoptionEvent) (schemascout.AdoptionEvent, error) {
	output := event
	if strings.TrimSpace(output.SchemaID) == "" {
		output.SchemaID = adoptionEventSchemaID
	}
	if output.SchemaID != adoptionEventSchemaID {
		return schemascout.AdoptionEvent{}, fmt.Errorf("unsupported schema_id %q", output.SchemaID)
	}
	if strings.TrimSpace(output.SchemaVersion) == "" {
		output.SchemaVersion = adoptionEventSchemaV1
	}
	if output.SchemaVersion != adoptionEventSchemaV1 {
		return schemascout.AdoptionEvent{}, fmt.Errorf("unsupported schema_version %q", output.SchemaVersion)
	}
	if output.CreatedAt.IsZero() {
		return schemascout.AdoptionEvent{}, fmt.Errorf("created_at is required")
	}
	output.CreatedAt = output.CreatedAt.UTC()
	output.ProducerVersion = strings.TrimSpace(output.ProducerVersion)
	if output.ProducerVersion == "" {
		output.ProducerVersion = "0.0.0-dev"
	}
	output.Command = strings.TrimSpace(output.Command)
	if output.Command == "" {
		return schemascout.AdoptionEvent{}, fmt.Errorf("command is required")
	}
	if output.ExitCode < 0 {
		return schemascout.AdoptionEvent{}, fmt.Errorf("exit_code must be >= 0")
	}
	if output.ExitCode > 255 {
		return schemascout.AdoptionEvent{}, fmt.Errorf("exit_code must be <= 255")
	}
	if output.ElapsedMS < 0 {
		output.ElapsedMS = 0
	}
	output.Milestones = uniqueSorted(output.Milestones)
	output.Environment.OS = strings.TrimSpace(output.Environment.OS)
	output.Environment.Arch = strings.TrimSpace(output.Environment.Arch)
	if output.Environment.OS == "" {
		return schemascout.AdoptionEvent{}, fmt.Errorf("environment.os is required")
	}
	if output.Environment.Arch == "" {
		return schemascout.AdoptionEvent{}, fmt.Errorf("environment.arch is required")
	}
	return output, nil
}

func milestoneTags(command string, success bool) []string {
	if !success {
		return nil
	}
	switch command {
	case "demo":
		return []string{"A1"}
	case "verify":
		return []string{"A2"}
	case "regress init":
		return []string{"A3"}
	case "regress run":
		return []string{"A4"}
	default:
		return nil
	}
}

func activationBlockerHint(milestone string) string {
	switch milestone {
	case "A1":
		return "missing A1: run `gait demo`"
	case "A2":
		return "missing A2: run `gait verify <run_id>`"
	case "A3":
		return "missing A3: run `gait regress init --from <run_id>`"
	case "A4":
		return "missing A4: run `gait regress run`"
	default:
		return "missing milestone: " + milestone
	}
}
