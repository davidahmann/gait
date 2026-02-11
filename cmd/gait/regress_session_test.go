package main

import (
	"path/filepath"
	"testing"
)

func TestRegressInitFromSessionChain(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)

	journalPath := filepath.Join(workDir, "sessions", "demo.journal.jsonl")
	checkpointPath := filepath.Join(workDir, "gait-out", "runpack_demo_cp_0001.zip")
	chainPath := filepath.Join(workDir, "sessions", "demo.journal_chain.json")

	if code := runCommand([]string{
		"session", "start",
		"--journal", journalPath,
		"--session-id", "sess_demo",
		"--run-id", "run_demo",
		"--json",
	}); code != exitOK {
		t.Fatalf("run session start expected %d got %d", exitOK, code)
	}
	if code := runCommand([]string{
		"session", "append",
		"--journal", journalPath,
		"--tool", "tool.write",
		"--verdict", "allow",
		"--intent-id", "intent_1",
		"--json",
	}); code != exitOK {
		t.Fatalf("run session append expected %d got %d", exitOK, code)
	}
	if code := runCommand([]string{
		"session", "checkpoint",
		"--journal", journalPath,
		"--out", checkpointPath,
		"--chain-out", chainPath,
		"--json",
	}); code != exitOK {
		t.Fatalf("run session checkpoint expected %d got %d", exitOK, code)
	}

	if code := runRegressInit([]string{
		"--from", chainPath,
		"--checkpoint", "latest",
		"--name", "fixture_from_chain",
		"--json",
	}); code != exitOK {
		t.Fatalf("runRegressInit from session chain expected %d got %d", exitOK, code)
	}
}
