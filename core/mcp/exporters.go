package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ExportEvent struct {
	CreatedAt       time.Time `json:"created_at"`
	ProducerVersion string    `json:"producer_version"`
	RunID           string    `json:"run_id"`
	TraceID         string    `json:"trace_id"`
	TracePath       string    `json:"trace_path,omitempty"`
	ToolName        string    `json:"tool_name"`
	Verdict         string    `json:"verdict"`
	ReasonCodes     []string  `json:"reason_codes,omitempty"`
	PolicyDigest    string    `json:"policy_digest"`
	IntentDigest    string    `json:"intent_digest"`
}

func ExportLogEvent(path string, event ExportEvent) error {
	entry := event
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
	}
	entry.CreatedAt = entry.CreatedAt.UTC()
	encoded, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("encode log export event: %w", err)
	}
	return appendJSONL(path, encoded)
}

func ExportOTelEvent(path string, event ExportEvent) error {
	createdAt := event.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
	}
	payload := map[string]any{
		"time_unix_nano": createdAt.UnixNano(),
		"severity_text":  "INFO",
		"body":           "gait.mcp.proxy.decision",
		"trace_id":       strings.TrimSpace(event.TraceID),
		"attributes": map[string]any{
			"gait.run_id":        strings.TrimSpace(event.RunID),
			"gait.trace_path":    strings.TrimSpace(event.TracePath),
			"gait.tool_name":     strings.TrimSpace(event.ToolName),
			"gait.verdict":       strings.TrimSpace(event.Verdict),
			"gait.policy_digest": strings.TrimSpace(event.PolicyDigest),
			"gait.intent_digest": strings.TrimSpace(event.IntentDigest),
			"gait.reason_codes":  event.ReasonCodes,
		},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode otel export event: %w", err)
	}
	return appendJSONL(path, encoded)
}

func appendJSONL(path string, line []byte) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create export directory: %w", err)
		}
	}
	// #nosec G304 -- export path is explicit local user input.
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open export file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()
	if _, err := file.Write(line); err != nil {
		return fmt.Errorf("write export file: %w", err)
	}
	if _, err := file.Write([]byte{'\n'}); err != nil {
		return fmt.Errorf("write export newline: %w", err)
	}
	return nil
}
