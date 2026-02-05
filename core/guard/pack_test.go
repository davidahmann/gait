package guard

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidahmann/gait/core/runpack"
	schemagate "github.com/davidahmann/gait/core/schema/v1/gate"
	schemaregress "github.com/davidahmann/gait/core/schema/v1/regress"
	schemarunpack "github.com/davidahmann/gait/core/schema/v1/runpack"
	schemascout "github.com/davidahmann/gait/core/schema/v1/scout"
	"github.com/davidahmann/gait/core/zipx"
)

func TestBuildAndVerifyPack(t *testing.T) {
	workDir := t.TempDir()
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	runpackPath := filepath.Join(workDir, "runpack_run_guard.zip")
	_, err := runpack.WriteRunpack(runpackPath, runpack.RecordOptions{
		Run: schemarunpack.Run{
			RunID:           "run_guard",
			CreatedAt:       now,
			ProducerVersion: "0.0.0-dev",
		},
		Intents: []schemarunpack.IntentRecord{{
			IntentID:   "intent_1",
			RunID:      "run_guard",
			ToolName:   "tool.delete",
			ArgsDigest: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		}},
		Results: []schemarunpack.ResultRecord{{
			IntentID:     "intent_1",
			RunID:        "run_guard",
			Status:       "ok",
			ResultDigest: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		}},
		Refs: schemarunpack.Refs{
			RunID: "run_guard",
			Receipts: []schemarunpack.RefReceipt{{
				RefID:         "ref_1",
				SourceType:    "web",
				SourceLocator: "https://example.com",
				QueryDigest:   "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				ContentDigest: "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
				RetrievedAt:   now,
				RedactionMode: "reference",
			}},
		},
		CaptureMode: "reference",
	})
	if err != nil {
		t.Fatalf("write runpack: %v", err)
	}

	inventory := schemascout.InventorySnapshot{
		SchemaID:        "gait.scout.inventory_snapshot",
		SchemaVersion:   "1.0.0",
		CreatedAt:       now,
		ProducerVersion: "0.0.0-dev",
		SnapshotID:      "snap_test",
		Items: []schemascout.InventoryItem{{
			ID:      "tool:framework:langchain:delete_user",
			Kind:    "tool",
			Name:    "delete_user",
			Locator: "agent.py",
		}},
	}
	inventoryPath := filepath.Join(workDir, "inventory.json")
	mustWriteJSON(t, inventoryPath, inventory)

	trace := schemagate.TraceRecord{
		SchemaID:        "gait.gate.trace",
		SchemaVersion:   "1.0.0",
		CreatedAt:       now,
		ProducerVersion: "0.0.0-dev",
		TraceID:         "trace_1",
		ToolName:        "tool.delete",
		ArgsDigest:      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		IntentDigest:    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		PolicyDigest:    "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		Verdict:         "allow",
	}
	tracePath := filepath.Join(workDir, "trace.json")
	mustWriteJSON(t, tracePath, trace)

	regress := schemaregress.RegressResult{
		SchemaID:        "gait.regress.result",
		SchemaVersion:   "1.0.0",
		CreatedAt:       now,
		ProducerVersion: "0.0.0-dev",
		FixtureSet:      "run_guard",
		Status:          "pass",
	}
	regressPath := filepath.Join(workDir, "regress.json")
	mustWriteJSON(t, regressPath, regress)

	packPath := filepath.Join(workDir, "evidence_pack.zip")
	buildResult, err := BuildPack(BuildOptions{
		RunpackPath:    runpackPath,
		OutputPath:     packPath,
		CaseID:         "case_1",
		InventoryPaths: []string{inventoryPath},
		TracePaths:     []string{tracePath},
		RegressPaths:   []string{regressPath},
	})
	if err != nil {
		t.Fatalf("build pack: %v", err)
	}
	if buildResult.Manifest.PackID == "" {
		t.Fatalf("expected pack id")
	}

	verifyResult, err := VerifyPack(packPath)
	if err != nil {
		t.Fatalf("verify pack: %v", err)
	}
	if len(verifyResult.MissingFiles) > 0 || len(verifyResult.HashMismatches) > 0 {
		t.Fatalf("expected clean verify result, got missing=%d mismatches=%d", len(verifyResult.MissingFiles), len(verifyResult.HashMismatches))
	}

	tamperedPath := filepath.Join(workDir, "evidence_pack_tampered.zip")
	tamperPackMissingFile(t, packPath, tamperedPath, "runpack_summary.json")
	tamperedVerify, err := VerifyPack(tamperedPath)
	if err != nil {
		t.Fatalf("verify tampered pack: %v", err)
	}
	if len(tamperedVerify.MissingFiles) == 0 {
		t.Fatalf("expected missing file in tampered pack")
	}
}

func tamperPackMissingFile(t *testing.T, source string, destination string, remove string) {
	t.Helper()
	reader, err := zip.OpenReader(source)
	if err != nil {
		t.Fatalf("open pack: %v", err)
	}
	defer func() {
		_ = reader.Close()
	}()
	files := make([]zipx.File, 0, len(reader.File))
	for _, zipFile := range reader.File {
		if zipFile.Name == remove {
			continue
		}
		fileReader, openErr := zipFile.Open()
		if openErr != nil {
			t.Fatalf("open zip entry %s: %v", zipFile.Name, openErr)
		}
		content := new(bytes.Buffer)
		if _, copyErr := content.ReadFrom(fileReader); copyErr != nil {
			_ = fileReader.Close()
			t.Fatalf("read zip entry %s: %v", zipFile.Name, copyErr)
		}
		_ = fileReader.Close()
		files = append(files, zipx.File{
			Path: zipFile.Name,
			Data: content.Bytes(),
			Mode: 0o644,
		})
	}
	var out bytes.Buffer
	if err := zipx.WriteDeterministicZip(&out, files); err != nil {
		t.Fatalf("write tampered zip: %v", err)
	}
	if err := os.WriteFile(destination, out.Bytes(), 0o600); err != nil {
		t.Fatalf("write tampered zip: %v", err)
	}
}

func mustWriteJSON(t *testing.T, path string, value any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write json file %s: %v", path, err)
	}
}
