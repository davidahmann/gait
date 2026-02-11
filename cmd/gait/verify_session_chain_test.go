package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunVerifySessionChainScenarios(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)

	if code := runVerifySessionChain([]string{"--help"}); code != exitOK {
		t.Fatalf("runVerifySessionChain help expected %d got %d", exitOK, code)
	}
	if code := runVerifySessionChain([]string{"--json"}); code != exitInvalidInput {
		t.Fatalf("runVerifySessionChain missing chain expected %d got %d", exitInvalidInput, code)
	}
	if code := runVerifySessionChain([]string{"--profile", "invalid", "--chain", "x.json", "--json"}); code != exitInvalidInput {
		t.Fatalf("runVerifySessionChain invalid profile expected %d got %d", exitInvalidInput, code)
	}
	if code := runVerifySessionChain([]string{"--profile", "strict", "--chain", "x.json", "--json"}); code != exitInvalidInput {
		t.Fatalf("runVerifySessionChain strict profile without key expected %d got %d", exitInvalidInput, code)
	}
	if code := runVerifySessionChain([]string{"--chain", filepath.Join(workDir, "missing.json"), "--json"}); code != exitInvalidInput {
		t.Fatalf("runVerifySessionChain missing chain file expected %d got %d", exitInvalidInput, code)
	}

	journalPath := filepath.Join(workDir, "sessions", "verify_chain.journal.jsonl")
	checkpointPath := filepath.Join(workDir, "gait-out", "verify_chain_cp_0001.zip")
	if code := runSessionStart([]string{
		"--journal", journalPath,
		"--session-id", "sess_verify",
		"--run-id", "run_verify",
		"--json",
	}); code != exitOK {
		t.Fatalf("runSessionStart expected %d got %d", exitOK, code)
	}
	if code := runSessionAppend([]string{
		"--journal", journalPath,
		"--tool", "tool.write",
		"--verdict", "allow",
		"--intent-id", "intent_verify",
		"--json",
	}); code != exitOK {
		t.Fatalf("runSessionAppend expected %d got %d", exitOK, code)
	}
	if code := runSessionCheckpoint([]string{
		"--journal", journalPath,
		"--out", checkpointPath,
		"--json",
	}); code != exitOK {
		t.Fatalf("runSessionCheckpoint expected %d got %d", exitOK, code)
	}

	chainPath := filepath.Join(workDir, "sessions", "verify_chain.journal_chain.json")
	if code := runVerifySessionChain([]string{"--json", chainPath}); code != exitOK {
		t.Fatalf("runVerifySessionChain positional chain expected %d got %d", exitOK, code)
	}
}

func TestWriteVerifySessionAndChainOutputTextModes(t *testing.T) {
	text := captureStdout(t, func() {
		if code := writeVerifyChainOutput(false, verifyChainOutput{
			OK: true,
			Run: verifyOutput{
				Path: "/tmp/runpack.zip",
			},
			Trace: &traceVerifyOutput{
				Path:            "/tmp/trace.json",
				SignatureStatus: "verified",
			},
			Pack: &guardVerifyOutput{
				Path:            "/tmp/pack.zip",
				SignatureStatus: "verified",
			},
		}, exitOK); code != exitOK {
			t.Fatalf("writeVerifyChainOutput ok expected exit %d got %d", exitOK, code)
		}
		if code := writeVerifyChainOutput(false, verifyChainOutput{
			OK:    false,
			Error: "chain_error",
		}, exitInternalFailure); code != exitInternalFailure {
			t.Fatalf("writeVerifyChainOutput error expected exit %d got %d", exitInternalFailure, code)
		}
		if code := writeVerifyChainOutput(false, verifyChainOutput{
			OK: false,
			Run: verifyOutput{
				Path: "/tmp/runpack.zip",
			},
			Trace: &traceVerifyOutput{
				Path:            "/tmp/trace.json",
				SignatureStatus: "failed",
			},
		}, exitVerifyFailed); code != exitVerifyFailed {
			t.Fatalf("writeVerifyChainOutput failed expected exit %d got %d", exitVerifyFailed, code)
		}

		if code := writeVerifySessionChainOutput(false, verifySessionChainOutput{
			OK:                 true,
			ChainPath:          "/tmp/session_chain.json",
			SessionID:          "sess",
			RunID:              "run",
			CheckpointsChecked: 2,
		}, exitOK); code != exitOK {
			t.Fatalf("writeVerifySessionChainOutput ok expected exit %d got %d", exitOK, code)
		}
		if code := writeVerifySessionChainOutput(false, verifySessionChainOutput{
			OK:    false,
			Error: "session_chain_error",
		}, exitInternalFailure); code != exitInternalFailure {
			t.Fatalf("writeVerifySessionChainOutput error expected exit %d got %d", exitInternalFailure, code)
		}
		if code := writeVerifySessionChainOutput(false, verifySessionChainOutput{
			OK:               false,
			ChainPath:        "/tmp/session_chain.json",
			LinkageErrors:    []string{"digest_mismatch"},
			CheckpointErrors: []string{"missing_checkpoint"},
		}, exitVerifyFailed); code != exitVerifyFailed {
			t.Fatalf("writeVerifySessionChainOutput failed expected exit %d got %d", exitVerifyFailed, code)
		}
	})

	expectedSnippets := []string{
		"verify chain: ok",
		"trace: /tmp/trace.json (verified)",
		"pack: /tmp/pack.zip (verified)",
		"verify chain error: chain_error",
		"verify chain failed",
		"verify session-chain: ok",
		"verify session-chain error: session_chain_error",
		"verify session-chain failed: /tmp/session_chain.json",
		"linkage errors: digest_mismatch",
		"checkpoint errors: missing_checkpoint",
	}
	for _, snippet := range expectedSnippets {
		if !strings.Contains(text, snippet) {
			t.Fatalf("expected output to contain %q, got:\n%s", snippet, text)
		}
	}
}

func TestParseArtifactVerifyProfile(t *testing.T) {
	standard, err := parseArtifactVerifyProfile("standard")
	if err != nil {
		t.Fatalf("parseArtifactVerifyProfile standard: %v", err)
	}
	if standard != verifyProfileStandard {
		t.Fatalf("expected standard profile, got %s", standard)
	}
	strict, err := parseArtifactVerifyProfile("strict")
	if err != nil {
		t.Fatalf("parseArtifactVerifyProfile strict: %v", err)
	}
	if strict != verifyProfileStrict {
		t.Fatalf("expected strict profile, got %s", strict)
	}
	if _, err := parseArtifactVerifyProfile("bad_profile"); err == nil {
		t.Fatalf("expected invalid profile parse error")
	}
}
