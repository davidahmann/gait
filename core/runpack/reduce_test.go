package runpack

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	schemarunpack "github.com/davidahmann/gait/core/schema/v1/runpack"
)

func TestReduceToMinimalMissingResult(t *testing.T) {
	workDir := t.TempDir()
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	runpackPath := filepath.Join(workDir, "runpack_run_reduce.zip")
	_, err := WriteRunpack(runpackPath, RecordOptions{
		Run: schemarunpack.Run{
			RunID:           "run_reduce",
			CreatedAt:       now,
			ProducerVersion: "0.0.0-dev",
		},
		Intents: []schemarunpack.IntentRecord{
			{
				IntentID:   "intent_a",
				RunID:      "run_reduce",
				ToolName:   "tool.read",
				ArgsDigest: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			{
				IntentID:   "intent_b",
				RunID:      "run_reduce",
				ToolName:   "tool.write",
				ArgsDigest: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			},
		},
		Results: []schemarunpack.ResultRecord{
			{
				IntentID:     "intent_a",
				RunID:        "run_reduce",
				Status:       "ok",
				ResultDigest: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			},
		},
		Refs: schemarunpack.Refs{RunID: "run_reduce"},
	})
	if err != nil {
		t.Fatalf("write runpack: %v", err)
	}

	outputPath := filepath.Join(workDir, "reduced.zip")
	result, err := ReduceToMinimal(ReduceOptions{
		InputPath:  runpackPath,
		OutputPath: outputPath,
		Predicate:  PredicateMissingResult,
	})
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	if result.Report.SelectedIntentID != "intent_b" {
		t.Fatalf("expected selected intent intent_b got %s", result.Report.SelectedIntentID)
	}
	if result.Report.ReducedIntentCount != 1 {
		t.Fatalf("expected 1 reduced intent got %d", result.Report.ReducedIntentCount)
	}
	if result.Report.ReducedResultCount != 0 {
		t.Fatalf("expected 0 reduced results got %d", result.Report.ReducedResultCount)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected reduced runpack at %s: %v", outputPath, err)
	}
}

func TestReduceToMinimalNonOKStatus(t *testing.T) {
	workDir := t.TempDir()
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	runpackPath := filepath.Join(workDir, "runpack_run_reduce_status.zip")
	_, err := WriteRunpack(runpackPath, RecordOptions{
		Run: schemarunpack.Run{
			RunID:           "run_reduce_status",
			CreatedAt:       now,
			ProducerVersion: "0.0.0-dev",
		},
		Intents: []schemarunpack.IntentRecord{
			{
				IntentID:   "intent_1",
				RunID:      "run_reduce_status",
				ToolName:   "tool.read",
				ArgsDigest: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		},
		Results: []schemarunpack.ResultRecord{
			{
				IntentID:     "intent_1",
				RunID:        "run_reduce_status",
				Status:       "error",
				ResultDigest: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			},
		},
		Refs: schemarunpack.Refs{RunID: "run_reduce_status"},
	})
	if err != nil {
		t.Fatalf("write runpack: %v", err)
	}

	_, err = ReduceToMinimal(ReduceOptions{
		InputPath: runpackPath,
		Predicate: PredicateNonOKStatus,
	})
	if err != nil {
		t.Fatalf("reduce non-ok status: %v", err)
	}
}
