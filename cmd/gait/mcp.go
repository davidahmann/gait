package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/davidahmann/gait/core/gate"
	"github.com/davidahmann/gait/core/mcp"
	"github.com/davidahmann/gait/core/runpack"
	schemarunpack "github.com/davidahmann/gait/core/schema/v1/runpack"
	"github.com/davidahmann/gait/core/sign"
)

type mcpProxyOutput struct {
	OK           bool     `json:"ok"`
	Adapter      string   `json:"adapter,omitempty"`
	RunID        string   `json:"run_id,omitempty"`
	Verdict      string   `json:"verdict,omitempty"`
	ReasonCodes  []string `json:"reason_codes,omitempty"`
	Violations   []string `json:"violations,omitempty"`
	PolicyDigest string   `json:"policy_digest,omitempty"`
	IntentDigest string   `json:"intent_digest,omitempty"`
	TraceID      string   `json:"trace_id,omitempty"`
	TracePath    string   `json:"trace_path,omitempty"`
	RunpackPath  string   `json:"runpack_path,omitempty"`
	LogExport    string   `json:"log_export,omitempty"`
	OTelExport   string   `json:"otel_export,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	Error        string   `json:"error,omitempty"`
}

func runMCP(arguments []string) int {
	if hasExplainFlag(arguments) {
		return writeExplain("Proxy tool-call protocol payloads through Gate policy evaluation and emit signed traces with optional exports.")
	}
	if len(arguments) == 0 {
		printMCPUsage()
		return exitInvalidInput
	}
	switch arguments[0] {
	case "proxy":
		return runMCPProxy(arguments[1:])
	case "bridge":
		return runMCPProxy(arguments[1:])
	default:
		printMCPUsage()
		return exitInvalidInput
	}
}

func runMCPProxy(arguments []string) int {
	if hasExplainFlag(arguments) {
		return writeExplain("Decode an MCP or adapter-formatted tool call, evaluate policy deterministically, and emit a signed gate-compatible trace.")
	}
	arguments = reorderInterspersedFlags(arguments, map[string]bool{
		"policy":          true,
		"call":            true,
		"adapter":         true,
		"trace-out":       true,
		"run-id":          true,
		"runpack-out":     true,
		"export-log-out":  true,
		"export-otel-out": true,
		"key-mode":        true,
		"private-key":     true,
		"private-key-env": true,
	})
	flagSet := flag.NewFlagSet("mcp-proxy", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var policyPath string
	var callPath string
	var adapter string
	var tracePath string
	var runID string
	var runpackOut string
	var logExportPath string
	var otelExportPath string
	var keyMode string
	var privateKeyPath string
	var privateKeyEnv string
	var jsonOutput bool
	var helpFlag bool

	flagSet.StringVar(&policyPath, "policy", "", "path to policy YAML")
	flagSet.StringVar(&callPath, "call", "", "path to tool call JSON (use '-' for stdin)")
	flagSet.StringVar(&adapter, "adapter", "mcp", "adapter payload format: mcp|openai|anthropic|langchain")
	flagSet.StringVar(&tracePath, "trace-out", "", "path to emitted trace JSON (default trace_<trace_id>.json)")
	flagSet.StringVar(&runID, "run-id", "", "optional run_id override for proxy artifacts")
	flagSet.StringVar(&runpackOut, "runpack-out", "", "optional path to emit a runpack zip for this proxy decision")
	flagSet.StringVar(&logExportPath, "export-log-out", "", "optional JSONL log export path")
	flagSet.StringVar(&otelExportPath, "export-otel-out", "", "optional OTEL-style JSONL export path")
	flagSet.StringVar(&keyMode, "key-mode", string(sign.ModeDev), "signing key mode: dev or prod")
	flagSet.StringVar(&privateKeyPath, "private-key", "", "path to base64 private signing key")
	flagSet.StringVar(&privateKeyEnv, "private-key-env", "", "env var containing base64 private signing key")
	flagSet.BoolVar(&jsonOutput, "json", false, "emit JSON output")
	flagSet.BoolVar(&helpFlag, "help", false, "show help")

	if err := flagSet.Parse(arguments); err != nil {
		return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	if helpFlag {
		printMCPProxyUsage()
		return exitOK
	}
	remaining := flagSet.Args()
	if strings.TrimSpace(policyPath) == "" && len(remaining) > 0 {
		policyPath = remaining[0]
		remaining = remaining[1:]
	}
	if strings.TrimSpace(callPath) == "" && len(remaining) > 0 {
		callPath = remaining[0]
		remaining = remaining[1:]
	}
	if strings.TrimSpace(policyPath) == "" || strings.TrimSpace(callPath) == "" || len(remaining) > 0 {
		return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{OK: false, Error: "expected --policy <policy.yaml> and --call <tool_call.json|->"}, exitInvalidInput)
	}

	payload, err := readMCPPayload(callPath)
	if err != nil {
		return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	call, err := mcp.DecodeToolCall(adapter, payload)
	if err != nil {
		return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}

	policy, err := gate.LoadPolicyFile(policyPath)
	if err != nil {
		return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}

	evalResult, err := mcp.EvaluateToolCall(policy, call, gate.EvalOptions{ProducerVersion: version})
	if err != nil {
		return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}

	keyPair, warnings, err := sign.LoadSigningKey(sign.KeyConfig{
		Mode:           sign.KeyMode(strings.ToLower(strings.TrimSpace(keyMode))),
		PrivateKeyPath: privateKeyPath,
		PrivateKeyEnv:  privateKeyEnv,
	})
	if err != nil {
		return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}
	traceResult, err := gate.EmitSignedTrace(policy, evalResult.Intent, evalResult.Outcome.Result, gate.EmitTraceOptions{
		ProducerVersion:   version,
		SigningPrivateKey: keyPair.Private,
		TracePath:         tracePath,
	})
	if err != nil {
		return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
	}

	resolvedRunID := normalizeRunID(runID)
	if resolvedRunID == "" {
		resolvedRunID = normalizeRunID(call.Context.RunID)
	}
	if resolvedRunID == "" {
		resolvedRunID = "run_mcp_" + evalResult.Intent.IntentDigest[:12]
	}

	resolvedRunpackPath := ""
	if strings.TrimSpace(runpackOut) != "" {
		resolvedRunpackPath = runpackOut
		if err := writeMCPRunpack(resolvedRunpackPath, resolvedRunID, evalResult, traceResult.Trace.TraceID); err != nil {
			return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
		}
	}

	exportEvent := mcp.ExportEvent{
		CreatedAt:       evalResult.Outcome.Result.CreatedAt,
		ProducerVersion: version,
		RunID:           resolvedRunID,
		TraceID:         traceResult.Trace.TraceID,
		TracePath:       traceResult.TracePath,
		ToolName:        evalResult.Intent.ToolName,
		Verdict:         evalResult.Outcome.Result.Verdict,
		ReasonCodes:     evalResult.Outcome.Result.ReasonCodes,
		PolicyDigest:    traceResult.PolicyDigest,
		IntentDigest:    traceResult.IntentDigest,
	}
	resolvedLogExport := ""
	if strings.TrimSpace(logExportPath) != "" {
		resolvedLogExport = strings.TrimSpace(logExportPath)
		if err := mcp.ExportLogEvent(resolvedLogExport, exportEvent); err != nil {
			return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
		}
	}
	resolvedOTelExport := ""
	if strings.TrimSpace(otelExportPath) != "" {
		resolvedOTelExport = strings.TrimSpace(otelExportPath)
		if err := mcp.ExportOTelEvent(resolvedOTelExport, exportEvent); err != nil {
			return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{OK: false, Error: err.Error()}, exitCodeForError(err, exitInvalidInput))
		}
	}

	exitCode := exitOK
	switch evalResult.Outcome.Result.Verdict {
	case "block":
		exitCode = exitPolicyBlocked
	case "require_approval":
		exitCode = exitApprovalRequired
	}
	return writeMCPProxyOutput(jsonOutput, mcpProxyOutput{
		OK:           true,
		Adapter:      strings.ToLower(strings.TrimSpace(adapter)),
		RunID:        resolvedRunID,
		Verdict:      evalResult.Outcome.Result.Verdict,
		ReasonCodes:  evalResult.Outcome.Result.ReasonCodes,
		Violations:   evalResult.Outcome.Result.Violations,
		PolicyDigest: traceResult.PolicyDigest,
		IntentDigest: traceResult.IntentDigest,
		TraceID:      traceResult.Trace.TraceID,
		TracePath:    traceResult.TracePath,
		RunpackPath:  resolvedRunpackPath,
		LogExport:    resolvedLogExport,
		OTelExport:   resolvedOTelExport,
		Warnings:     warnings,
	}, exitCode)
}

func readMCPPayload(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "-" {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin payload: %w", err)
		}
		return raw, nil
	}
	// #nosec G304 -- call path is explicit local user input.
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read call payload: %w", err)
	}
	return raw, nil
}

func writeMCPRunpack(path string, runID string, evalResult mcp.EvalResult, traceID string) error {
	now := evalResult.Outcome.Result.CreatedAt.UTC()
	if now.IsZero() {
		now = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
	}
	resultPayload := map[string]any{
		"verdict":      evalResult.Outcome.Result.Verdict,
		"reason_codes": evalResult.Outcome.Result.ReasonCodes,
		"violations":   evalResult.Outcome.Result.Violations,
		"trace_id":     traceID,
	}
	resultDigest, err := digestObject(resultPayload)
	if err != nil {
		return fmt.Errorf("digest proxy result: %w", err)
	}
	resultStatus := "ok"
	if evalResult.Outcome.Result.Verdict == "block" || evalResult.Outcome.Result.Verdict == "require_approval" {
		resultStatus = "error"
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create runpack directory: %w", err)
	}
	_, err = runpack.WriteRunpack(path, runpack.RecordOptions{
		Run: schemarunpack.Run{
			SchemaID:        "gait.runpack.run",
			SchemaVersion:   "1.0.0",
			CreatedAt:       now,
			ProducerVersion: version,
			RunID:           runID,
			Timeline: []schemarunpack.TimelineEvt{
				{Event: "proxy_eval_start", TS: now},
				{Event: "proxy_eval_finish", TS: now},
			},
		},
		Intents: []schemarunpack.IntentRecord{{
			SchemaID:        "gait.runpack.intent",
			SchemaVersion:   "1.0.0",
			CreatedAt:       now,
			ProducerVersion: version,
			RunID:           runID,
			IntentID:        "intent_1",
			ToolName:        evalResult.Intent.ToolName,
			ArgsDigest:      evalResult.Intent.ArgsDigest,
			Args:            evalResult.Intent.Args,
		}},
		Results: []schemarunpack.ResultRecord{{
			SchemaID:        "gait.runpack.result",
			SchemaVersion:   "1.0.0",
			CreatedAt:       now,
			ProducerVersion: version,
			RunID:           runID,
			IntentID:        "intent_1",
			Status:          resultStatus,
			ResultDigest:    resultDigest,
			Result:          resultPayload,
		}},
		Refs: schemarunpack.Refs{
			SchemaID:        "gait.runpack.refs",
			SchemaVersion:   "1.0.0",
			CreatedAt:       now,
			ProducerVersion: version,
			RunID:           runID,
			Receipts:        []schemarunpack.RefReceipt{},
		},
		CaptureMode: "reference",
	})
	if err != nil {
		return fmt.Errorf("write proxy runpack: %w", err)
	}
	return nil
}

var runIDSanitizer = regexp.MustCompile(`[^A-Za-z0-9_-]+`)

func normalizeRunID(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	trimmed = runIDSanitizer.ReplaceAllString(trimmed, "_")
	trimmed = strings.Trim(trimmed, "_")
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "run_") {
		return trimmed
	}
	return "run_" + trimmed
}

func writeMCPProxyOutput(jsonOutput bool, output mcpProxyOutput, exitCode int) int {
	if jsonOutput {
		return writeJSONOutput(output, exitCode)
	}
	if output.OK {
		fmt.Printf("mcp proxy: verdict=%s\n", output.Verdict)
		fmt.Printf("trace: %s\n", output.TracePath)
		if output.RunpackPath != "" {
			fmt.Printf("runpack: %s\n", output.RunpackPath)
		}
		return exitCode
	}
	fmt.Printf("mcp proxy error: %s\n", output.Error)
	return exitCode
}

func printMCPUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait mcp proxy --policy <policy.yaml> --call <tool_call.json|-> [--adapter mcp|openai|anthropic|langchain] [--trace-out trace.json] [--run-id run_...] [--runpack-out runpack.zip] [--export-log-out events.jsonl] [--export-otel-out otel.jsonl] [--json] [--explain]")
	fmt.Println("  gait mcp bridge --policy <policy.yaml> --call <tool_call.json|-> [--adapter mcp|openai|anthropic|langchain] [--trace-out trace.json] [--run-id run_...] [--runpack-out runpack.zip] [--export-log-out events.jsonl] [--export-otel-out otel.jsonl] [--json] [--explain]")
}

func printMCPProxyUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gait mcp proxy --policy <policy.yaml> --call <tool_call.json|-> [--adapter mcp|openai|anthropic|langchain] [--trace-out trace.json] [--run-id run_...] [--runpack-out runpack.zip] [--export-log-out events.jsonl] [--export-otel-out otel.jsonl] [--key-mode dev|prod] [--private-key <path>|--private-key-env <VAR>] [--json] [--explain]")
}
