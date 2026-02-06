package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davidahmann/gait/core/gate"
)

func TestDecodeToolCallAdapters(t *testing.T) {
	openaiPayload := []byte(`{
  "type": "function",
  "function": {
    "name": "tool.write",
    "arguments": "{\"path\":\"/tmp/out.txt\"}"
  }
}`)
	call, err := DecodeToolCall("openai", openaiPayload)
	if err != nil {
		t.Fatalf("decode openai call: %v", err)
	}
	if call.Name != "tool.write" || call.Args["path"] != "/tmp/out.txt" {
		t.Fatalf("unexpected openai call: %#v", call)
	}

	anthropicPayload := []byte(`{
  "type": "tool_use",
  "name": "tool.fetch",
  "input": {"url":"https://example.local"}
}`)
	call, err = DecodeToolCall("anthropic", anthropicPayload)
	if err != nil {
		t.Fatalf("decode anthropic call: %v", err)
	}
	if call.Name != "tool.fetch" || call.Args["url"] != "https://example.local" {
		t.Fatalf("unexpected anthropic call: %#v", call)
	}

	langchainPayload := []byte(`{
  "tool": "tool.search",
  "tool_input": {"query":"gait"}
}`)
	call, err = DecodeToolCall("langchain", langchainPayload)
	if err != nil {
		t.Fatalf("decode langchain call: %v", err)
	}
	if call.Name != "tool.search" || call.Args["query"] != "gait" {
		t.Fatalf("unexpected langchain call: %#v", call)
	}

	mcpPayload := []byte(`{
  "name":"tool.read",
  "args":{"path":"/tmp/in.txt"},
  "target":"api.external.com"
}`)
	call, err = DecodeToolCall("mcp", mcpPayload)
	if err != nil {
		t.Fatalf("decode mcp call: %v", err)
	}
	if call.Name != "tool.read" || call.Target != "api.external.com" {
		t.Fatalf("unexpected mcp call: %#v", call)
	}

	if _, err := DecodeToolCall("unsupported", []byte(`{}`)); err == nil {
		t.Fatalf("expected unsupported adapter error")
	}
}

func TestEvaluateToolCallDefaultsAndLegacyTarget(t *testing.T) {
	policy, err := gate.ParsePolicyYAML([]byte(`
default_verdict: allow
rules:
  - name: block-host
    effect: block
    match:
      tool_names: [tool.write]
      target_kinds: [host]
      target_values: [api.external.com]
`))
	if err != nil {
		t.Fatalf("parse policy: %v", err)
	}

	result, err := EvaluateToolCall(policy, ToolCall{
		Name:   "tool.write",
		Args:   map[string]any{"path": "/tmp/out.txt"},
		Target: "api.external.com",
	}, gate.EvalOptions{ProducerVersion: "0.0.0-test"})
	if err != nil {
		t.Fatalf("evaluate tool call: %v", err)
	}
	if result.Intent.Context.Identity != defaultIdentity || result.Intent.Context.Workspace != defaultWorkspace {
		t.Fatalf("unexpected default context: %#v", result.Intent.Context)
	}
	if result.Intent.Targets[0].Kind != "host" {
		t.Fatalf("expected inferred host target: %#v", result.Intent.Targets)
	}
	if result.Outcome.Result.Verdict != "block" {
		t.Fatalf("expected block verdict, got %#v", result.Outcome.Result)
	}
}

func TestExportersWriteJSONL(t *testing.T) {
	workDir := t.TempDir()
	logPath := filepath.Join(workDir, "mcp.log.jsonl")
	otelPath := filepath.Join(workDir, "mcp.otel.jsonl")
	event := ExportEvent{
		RunID:           "run_mcp_test",
		TraceID:         "trace_1",
		ToolName:        "tool.write",
		Verdict:         "allow",
		PolicyDigest:    strings.Repeat("a", 64),
		IntentDigest:    strings.Repeat("b", 64),
		ProducerVersion: "0.0.0-test",
	}
	if err := ExportLogEvent(logPath, event); err != nil {
		t.Fatalf("export log event: %v", err)
	}
	if err := ExportOTelEvent(otelPath, event); err != nil {
		t.Fatalf("export otel event: %v", err)
	}

	logRaw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log export: %v", err)
	}
	var logEntry map[string]any
	if err := json.Unmarshal(bytesTrimNewline(logRaw), &logEntry); err != nil {
		t.Fatalf("parse log export: %v", err)
	}
	if logEntry["run_id"] != "run_mcp_test" {
		t.Fatalf("unexpected log run id: %#v", logEntry)
	}

	otelRaw, err := os.ReadFile(otelPath)
	if err != nil {
		t.Fatalf("read otel export: %v", err)
	}
	var otelEntry map[string]any
	if err := json.Unmarshal(bytesTrimNewline(otelRaw), &otelEntry); err != nil {
		t.Fatalf("parse otel export: %v", err)
	}
	attrs, ok := otelEntry["attributes"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected otel attributes: %#v", otelEntry)
	}
	if attrs["gait.run_id"] != "run_mcp_test" {
		t.Fatalf("unexpected otel run id: %#v", attrs)
	}
}

func bytesTrimNewline(raw []byte) []byte {
	return []byte(strings.TrimSpace(string(raw)))
}
