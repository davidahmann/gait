package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMCPProxyBlockWithArtifacts(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)

	privateKeyPath := filepath.Join(workDir, "trace_private.key")
	writePrivateKey(t, privateKeyPath)

	policyPath := filepath.Join(workDir, "policy.yaml")
	mustWriteFile(t, policyPath, `default_verdict: allow
rules:
  - name: block-write-host
    effect: block
    match:
      tool_names: [tool.write]
      target_kinds: [host]
      target_values: [api.external.com]
`)
	callPath := filepath.Join(workDir, "call.json")
	mustWriteFile(t, callPath, `{
  "name":"tool.write",
  "args":{"path":"/tmp/out.txt"},
  "target":"api.external.com",
  "context":{"identity":"alice","workspace":"/repo/gait","risk_class":"high","run_id":"run_mcp_case"}
}`)

	tracePath := filepath.Join(workDir, "trace_mcp.json")
	runpackPath := filepath.Join(workDir, "runpack_mcp.zip")
	logPath := filepath.Join(workDir, "mcp_events.jsonl")
	otelPath := filepath.Join(workDir, "mcp_otel.jsonl")

	if code := runMCPProxy([]string{
		"--policy", policyPath,
		"--call", callPath,
		"--trace-out", tracePath,
		"--runpack-out", runpackPath,
		"--export-log-out", logPath,
		"--export-otel-out", otelPath,
		"--key-mode", "prod",
		"--private-key", privateKeyPath,
		"--json",
	}); code != exitPolicyBlocked {
		t.Fatalf("runMCPProxy blocked expected %d got %d", exitPolicyBlocked, code)
	}

	if _, err := os.Stat(tracePath); err != nil {
		t.Fatalf("expected trace artifact: %v", err)
	}
	if _, err := os.Stat(runpackPath); err != nil {
		t.Fatalf("expected runpack artifact: %v", err)
	}
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected log export artifact: %v", err)
	}
	if _, err := os.Stat(otelPath); err != nil {
		t.Fatalf("expected otel export artifact: %v", err)
	}
	if code := runVerify([]string{"--json", runpackPath}); code != exitOK {
		t.Fatalf("runVerify expected %d got %d", exitOK, code)
	}
}

func TestRunMCPProxyOpenAIAdapter(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)

	policyPath := filepath.Join(workDir, "policy.yaml")
	mustWriteFile(t, policyPath, `default_verdict: allow`)
	callPath := filepath.Join(workDir, "openai_call.json")
	mustWriteFile(t, callPath, `{
  "type":"function",
  "function":{
    "name":"tool.search",
    "arguments":"{\"query\":\"gait\"}"
  }
}`)

	if code := runMCPProxy([]string{
		"--policy", policyPath,
		"--call", callPath,
		"--adapter", "openai",
		"--json",
	}); code != exitOK {
		t.Fatalf("runMCPProxy openai expected %d got %d", exitOK, code)
	}
}

func TestRunMCPProxyAdaptersSupportRunpackAndRegressInit(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)

	policyPath := filepath.Join(workDir, "policy_allow.yaml")
	mustWriteFile(t, policyPath, `default_verdict: allow`)

	cases := []struct {
		name    string
		adapter string
		payload string
	}{
		{
			name:    "mcp",
			adapter: "mcp",
			payload: `{"name":"tool.search","args":{"query":"gait"}}`,
		},
		{
			name:    "openai",
			adapter: "openai",
			payload: `{"type":"function","function":{"name":"tool.search","arguments":"{\"query\":\"gait\"}"}}`,
		},
		{
			name:    "anthropic",
			adapter: "anthropic",
			payload: `{"type":"tool_use","name":"tool.search","input":{"query":"gait"}}`,
		},
		{
			name:    "langchain",
			adapter: "langchain",
			payload: `{"tool":"tool.search","tool_input":{"query":"gait"}}`,
		},
	}

	for _, testCase := range cases {
		callPath := filepath.Join(workDir, testCase.name+"_call.json")
		runpackPath := filepath.Join(workDir, testCase.name+"_runpack.zip")
		mustWriteFile(t, callPath, testCase.payload)

		if code := runMCPProxy([]string{
			"--policy", policyPath,
			"--call", callPath,
			"--adapter", testCase.adapter,
			"--runpack-out", runpackPath,
			"--json",
		}); code != exitOK {
			t.Fatalf("runMCPProxy %s expected %d got %d", testCase.adapter, exitOK, code)
		}

		if code := runVerify([]string{"--json", runpackPath}); code != exitOK {
			t.Fatalf("runVerify %s expected %d got %d", testCase.adapter, exitOK, code)
		}

		fixtureName := "fixture_" + testCase.name
		if code := runRegressInit([]string{
			"--from", runpackPath,
			"--name", fixtureName,
			"--json",
		}); code != exitOK {
			t.Fatalf("runRegressInit %s expected %d got %d", testCase.adapter, exitOK, code)
		}
	}
}

func TestRunMCPProxyValidation(t *testing.T) {
	if code := runMCPProxy([]string{}); code != exitInvalidInput {
		t.Fatalf("runMCPProxy missing args expected %d got %d", exitInvalidInput, code)
	}
	if code := runMCP([]string{}); code != exitInvalidInput {
		t.Fatalf("runMCP missing args expected %d got %d", exitInvalidInput, code)
	}
	if code := runMCP([]string{"unknown"}); code != exitInvalidInput {
		t.Fatalf("runMCP unknown expected %d got %d", exitInvalidInput, code)
	}
	if code := runMCP([]string{"bridge", "--help"}); code != exitOK {
		t.Fatalf("runMCP bridge help expected %d got %d", exitOK, code)
	}
	if code := writeMCPProxyOutput(false, mcpProxyOutput{OK: true, Verdict: "allow"}, exitOK); code != exitOK {
		t.Fatalf("writeMCPProxyOutput text success expected %d got %d", exitOK, code)
	}
	if code := writeMCPProxyOutput(true, mcpProxyOutput{OK: true, Verdict: "allow"}, exitOK); code != exitOK {
		t.Fatalf("writeMCPProxyOutput json expected %d got %d", exitOK, code)
	}
	if code := writeMCPProxyOutput(false, mcpProxyOutput{OK: false, Error: "bad"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeMCPProxyOutput text expected %d got %d", exitInvalidInput, code)
	}
	printMCPUsage()
	printMCPProxyUsage()
}

func TestReadMCPPayloadAndRunIDHelpers(t *testing.T) {
	workDir := t.TempDir()
	callPath := filepath.Join(workDir, "call.json")
	mustWriteFile(t, callPath, `{"name":"tool.read"}`)
	payload, err := readMCPPayload(callPath)
	if err != nil {
		t.Fatalf("readMCPPayload file: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(payload, &parsed); err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	if parsed["name"] != "tool.read" {
		t.Fatalf("unexpected payload: %#v", parsed)
	}
	if _, err := readMCPPayload(filepath.Join(workDir, "missing.json")); err == nil {
		t.Fatalf("expected readMCPPayload missing file error")
	}

	stdinPath := filepath.Join(workDir, "stdin_call.json")
	mustWriteFile(t, stdinPath, `{"name":"tool.stdin"}`)
	stdinFile, err := os.Open(stdinPath)
	if err != nil {
		t.Fatalf("open stdin fixture: %v", err)
	}
	defer func() {
		_ = stdinFile.Close()
	}()
	originalStdin := os.Stdin
	defer func() {
		os.Stdin = originalStdin
	}()
	os.Stdin = stdinFile
	stdinPayload, err := readMCPPayload("-")
	if err != nil {
		t.Fatalf("readMCPPayload stdin: %v", err)
	}
	if !strings.Contains(string(stdinPayload), "tool.stdin") {
		t.Fatalf("unexpected stdin payload: %s", string(stdinPayload))
	}

	if normalized := normalizeRunID(""); normalized != "" {
		t.Fatalf("expected empty normalized run id")
	}
	if normalized := normalizeRunID("my run id"); normalized != "run_my_run_id" {
		t.Fatalf("unexpected normalized run id: %s", normalized)
	}
	if normalized := normalizeRunID("run_existing"); normalized != "run_existing" {
		t.Fatalf("unexpected pre-normalized run id: %s", normalized)
	}
}
