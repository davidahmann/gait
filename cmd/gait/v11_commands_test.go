package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidahmann/gait/core/runpack"
	schemarunpack "github.com/davidahmann/gait/core/schema/v1/runpack"
	schemascout "github.com/davidahmann/gait/core/schema/v1/scout"
)

func TestScoutCommandsAndWriters(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)

	left := schemascout.InventorySnapshot{
		SchemaID:        "gait.scout.inventory_snapshot",
		SchemaVersion:   "1.0.0",
		CreatedAt:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		ProducerVersion: "0.0.0-dev",
		SnapshotID:      "snap_left",
		Items: []schemascout.InventoryItem{{
			ID:      "tool:a",
			Kind:    "tool",
			Name:    "a",
			Locator: "a.py",
		}},
	}
	right := left
	right.SnapshotID = "snap_right"
	right.Items = append(right.Items[:0:0], right.Items...)
	right.Items[0].Locator = "b.py"

	leftPath := filepath.Join(workDir, "left.json")
	rightPath := filepath.Join(workDir, "right.json")
	diffOut := filepath.Join(workDir, "diff.json")
	mustWriteJSONFile(t, leftPath, left)
	mustWriteJSONFile(t, rightPath, right)

	if code := runScoutDiff([]string{"--left", leftPath, "--right", rightPath, "--out", diffOut, "--json"}); code != exitVerifyFailed {
		t.Fatalf("runScoutDiff changed expected %d got %d", exitVerifyFailed, code)
	}
	if _, err := os.Stat(diffOut); err != nil {
		t.Fatalf("expected scout diff output: %v", err)
	}
	if code := runScoutDiff([]string{"--json", leftPath, leftPath}); code != exitOK {
		t.Fatalf("runScoutDiff identical expected %d got %d", exitOK, code)
	}
	if _, err := readInventorySnapshot(leftPath); err != nil {
		t.Fatalf("readInventorySnapshot: %v", err)
	}
	if _, err := readInventorySnapshot(filepath.Join(workDir, "missing.json")); err == nil {
		t.Fatalf("expected readInventorySnapshot missing file error")
	}

	if code := writeScoutSnapshotOutput(true, scoutSnapshotOutput{OK: true, SnapshotID: "snap"}, exitOK); code != exitOK {
		t.Fatalf("writeScoutSnapshotOutput json expected %d got %d", exitOK, code)
	}
	if code := writeScoutSnapshotOutput(false, scoutSnapshotOutput{OK: false, Error: "x"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeScoutSnapshotOutput text expected %d got %d", exitInvalidInput, code)
	}
	if code := writeScoutDiffOutput(true, scoutDiffOutput{OK: true}, exitOK); code != exitOK {
		t.Fatalf("writeScoutDiffOutput json expected %d got %d", exitOK, code)
	}
	if code := writeScoutDiffOutput(false, scoutDiffOutput{OK: false, Error: "x"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeScoutDiffOutput text expected %d got %d", exitInvalidInput, code)
	}
	printScoutUsage()
	printScoutSnapshotUsage()
	printScoutDiffUsage()
}

func TestGuardRegistryAndReduceWriters(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)

	if code := writeGuardPackOutput(true, guardPackOutput{OK: true, PackID: "pack_1"}, exitOK); code != exitOK {
		t.Fatalf("writeGuardPackOutput json expected %d got %d", exitOK, code)
	}
	if code := writeGuardPackOutput(false, guardPackOutput{OK: false, Error: "x"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeGuardPackOutput text expected %d got %d", exitInvalidInput, code)
	}
	if code := writeGuardVerifyOutput(true, guardVerifyOutput{OK: true, PackID: "pack_1"}, exitOK); code != exitOK {
		t.Fatalf("writeGuardVerifyOutput json expected %d got %d", exitOK, code)
	}
	if code := writeGuardVerifyOutput(false, guardVerifyOutput{OK: false, Error: "x"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeGuardVerifyOutput text expected %d got %d", exitInvalidInput, code)
	}
	if code := writeRegistryInstallOutput(true, registryInstallOutput{OK: true, PackName: "p"}, exitOK); code != exitOK {
		t.Fatalf("writeRegistryInstallOutput json expected %d got %d", exitOK, code)
	}
	if code := writeRegistryInstallOutput(false, registryInstallOutput{OK: false, Error: "x"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeRegistryInstallOutput text expected %d got %d", exitInvalidInput, code)
	}
	if code := writeRegistryListOutput(true, registryListOutput{OK: true, Packs: nil}, exitOK); code != exitOK {
		t.Fatalf("writeRegistryListOutput json expected %d got %d", exitOK, code)
	}
	if code := writeRegistryListOutput(false, registryListOutput{OK: false, Error: "x"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeRegistryListOutput text expected %d got %d", exitInvalidInput, code)
	}
	if code := writeRegistryVerifyOutput(true, registryVerifyOutput{OK: true, PackName: "pack"}, exitOK); code != exitOK {
		t.Fatalf("writeRegistryVerifyOutput json expected %d got %d", exitOK, code)
	}
	if code := writeRegistryVerifyOutput(false, registryVerifyOutput{OK: false, Error: "x"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeRegistryVerifyOutput text error expected %d got %d", exitInvalidInput, code)
	}
	if code := writeRegistryVerifyOutput(false, registryVerifyOutput{OK: false, SignatureError: "bad sig"}, exitVerifyFailed); code != exitVerifyFailed {
		t.Fatalf("writeRegistryVerifyOutput signature branch expected %d got %d", exitVerifyFailed, code)
	}
	if code := writeRegistryVerifyOutput(false, registryVerifyOutput{OK: false, PackName: "pack-no-pin"}, exitVerifyFailed); code != exitVerifyFailed {
		t.Fatalf("writeRegistryVerifyOutput pin-mismatch branch expected %d got %d", exitVerifyFailed, code)
	}
	printGuardUsage()
	printGuardPackUsage()
	printGuardVerifyUsage()
	printRegistryUsage()
	printRegistryInstallUsage()
	printRegistryListUsage()
	printRegistryVerifyUsage()
	if code := runRegistryList([]string{"unexpected"}); code != exitInvalidInput {
		t.Fatalf("runRegistryList positional arg expected %d got %d", exitInvalidInput, code)
	}
	if code := runRegistryVerify([]string{"--json"}); code != exitInvalidInput {
		t.Fatalf("runRegistryVerify missing path expected %d got %d", exitInvalidInput, code)
	}

	runpackPath := filepath.Join(workDir, "runpack_run_reduce_writer.zip")
	_, err := runpack.WriteRunpack(runpackPath, runpack.RecordOptions{
		Run: schemarunpack.Run{
			RunID:           "run_reduce_writer",
			CreatedAt:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			ProducerVersion: "0.0.0-dev",
		},
		Intents: []schemarunpack.IntentRecord{{
			IntentID:   "intent_1",
			RunID:      "run_reduce_writer",
			ToolName:   "tool.write",
			ArgsDigest: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		}},
		Results: []schemarunpack.ResultRecord{{
			IntentID:     "intent_1",
			RunID:        "run_reduce_writer",
			Status:       "error",
			ResultDigest: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		}},
		Refs: schemarunpack.Refs{RunID: "run_reduce_writer"},
	})
	if err != nil {
		t.Fatalf("write runpack: %v", err)
	}
	if code := runReduce([]string{"--from", runpackPath, "--predicate", "non_ok_status", "--json"}); code != exitOK {
		t.Fatalf("runReduce non_ok_status expected %d got %d", exitOK, code)
	}
	if code := runReduce([]string{"--from", runpackPath, "--predicate", "bad"}); code != exitInvalidInput {
		t.Fatalf("runReduce invalid predicate expected %d got %d", exitInvalidInput, code)
	}
	if code := writeReduceOutput(true, reduceOutput{OK: true, RunID: "r"}, exitOK); code != exitOK {
		t.Fatalf("writeReduceOutput json expected %d got %d", exitOK, code)
	}
	if code := writeReduceOutput(false, reduceOutput{OK: false, Error: "x"}, exitInvalidInput); code != exitInvalidInput {
		t.Fatalf("writeReduceOutput text expected %d got %d", exitInvalidInput, code)
	}
	printReduceUsage()
}

func mustWriteJSONFile(t *testing.T, path string, value any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write json file %s: %v", path, err)
	}
}
