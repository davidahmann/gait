package pack

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	coreerrors "github.com/davidahmann/gait/core/errors"
	schemavoice "github.com/davidahmann/gait/core/schema/v1/voice"
)

func TestBuildCallPackVerifyInspectAndExtractRunpack(t *testing.T) {
	workDir := t.TempDir()
	runpackPath := createRunpackFixture(t, workDir, "run_voice_case")
	callRecordPath := writeCallRecordFixture(t, workDir, runpackPath, "call_voice_case")

	result, err := BuildCallPack(BuildCallOptions{
		CallRecordPath:  callRecordPath,
		OutputPath:      filepath.Join(workDir, "callpack.zip"),
		ProducerVersion: "test-vvoice",
	})
	if err != nil {
		t.Fatalf("build call pack: %v", err)
	}
	if result.Manifest.PackType != string(BuildTypeCall) {
		t.Fatalf("unexpected call pack type: %s", result.Manifest.PackType)
	}
	verifyResult, err := Verify(result.Path, VerifyOptions{})
	if err != nil {
		t.Fatalf("verify call pack: %v", err)
	}
	if verifyResult.PackType != string(BuildTypeCall) {
		t.Fatalf("unexpected verify pack type: %s", verifyResult.PackType)
	}
	if len(verifyResult.HashMismatches) > 0 || len(verifyResult.MissingFiles) > 0 {
		t.Fatalf("expected clean verify result, got %#v", verifyResult)
	}
	inspectResult, err := Inspect(result.Path)
	if err != nil {
		t.Fatalf("inspect call pack: %v", err)
	}
	if inspectResult.CallPayload == nil {
		t.Fatalf("expected call payload in inspect output")
	}
	if inspectResult.CallPayload.CallID != "call_voice_case" {
		t.Fatalf("unexpected call payload call_id: %s", inspectResult.CallPayload.CallID)
	}
	diffResult, err := Diff(result.Path, result.Path)
	if err != nil {
		t.Fatalf("diff identical call packs: %v", err)
	}
	if diffResult.Result.Summary.Changed {
		t.Fatalf("identical callpack diff should not report changed")
	}
	extractedRunpack, err := ExtractRunpack(result.Path)
	if err != nil {
		t.Fatalf("extract runpack from callpack: %v", err)
	}
	originalRunpack, err := os.ReadFile(runpackPath)
	if err != nil {
		t.Fatalf("read source runpack: %v", err)
	}
	if !bytes.Equal(extractedRunpack, originalRunpack) {
		t.Fatalf("expected extracted runpack bytes to match source")
	}
}

func TestVerifyRejectsCallPackSpeakWithoutToken(t *testing.T) {
	workDir := t.TempDir()
	runpackPath := createRunpackFixture(t, workDir, "run_voice_missing_token")
	callRecordPath := writeCallRecordFixture(t, workDir, runpackPath, "call_voice_missing_token")

	result, err := BuildCallPack(BuildCallOptions{
		CallRecordPath: callRecordPath,
		OutputPath:     filepath.Join(workDir, "callpack.zip"),
	})
	if err != nil {
		t.Fatalf("build call pack: %v", err)
	}

	mutatedPath := filepath.Join(workDir, "callpack_missing_token.zip")
	invalidSpeak := []byte(`{"call_id":"call_voice_missing_token","call_seq":5,"turn_index":2,"commitment_class":"quote","say_token_id":"","spoken_digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","emitted_at":"2026-02-15T00:00:05Z"}` + "\n")
	if err := rewritePackWithMutatedPayloadAndManifest(result.Path, mutatedPath, map[string][]byte{
		"speak_receipts.jsonl": invalidSpeak,
	}); err != nil {
		t.Fatalf("rewrite callpack with invalid speak receipt: %v", err)
	}
	if _, err := Verify(mutatedPath, VerifyOptions{}); err == nil {
		t.Fatalf("expected verify to fail when speak receipt has missing say_token_id")
	} else if coreerrors.CategoryOf(err) != coreerrors.CategoryVerification {
		t.Fatalf("expected verification-category error, got %q (%v)", coreerrors.CategoryOf(err), err)
	}
}

func TestBuildCallPackRejectsSpeakWithoutPriorAllowDecision(t *testing.T) {
	workDir := t.TempDir()
	runpackPath := createRunpackFixture(t, workDir, "run_voice_allow_order")
	callRecordPath := writeCallRecordFixture(t, workDir, runpackPath, "call_voice_allow_order")

	recordBytes, err := os.ReadFile(callRecordPath)
	if err != nil {
		t.Fatalf("read call record: %v", err)
	}
	var record map[string]any
	if err := json.Unmarshal(recordBytes, &record); err != nil {
		t.Fatalf("decode call record: %v", err)
	}
	decisions, ok := record["gate_decisions"].([]any)
	if !ok || len(decisions) == 0 {
		t.Fatalf("expected gate_decisions in call record")
	}
	firstDecision, ok := decisions[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected gate_decisions shape")
	}
	firstDecision["call_seq"] = 9
	mutated, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("encode mutated call record: %v", err)
	}
	mutated = append(mutated, '\n')
	if err := os.WriteFile(callRecordPath, mutated, 0o600); err != nil {
		t.Fatalf("write mutated call record: %v", err)
	}

	_, err = BuildCallPack(BuildCallOptions{
		CallRecordPath: callRecordPath,
		OutputPath:     filepath.Join(workDir, "callpack.zip"),
	})
	if err == nil {
		t.Fatalf("expected callpack build to fail without prior allow decision")
	}
	if !strings.Contains(err.Error(), "missing prior allow gate decision") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildCallPackRejectsGatedEmitWithoutSpeakReceipt(t *testing.T) {
	workDir := t.TempDir()
	runpackPath := createRunpackFixture(t, workDir, "run_voice_missing_receipt")
	callRecordPath := writeCallRecordFixture(t, workDir, runpackPath, "call_voice_missing_receipt")

	recordBytes, err := os.ReadFile(callRecordPath)
	if err != nil {
		t.Fatalf("read call record: %v", err)
	}
	var record map[string]any
	if err := json.Unmarshal(recordBytes, &record); err != nil {
		t.Fatalf("decode call record: %v", err)
	}
	record["speak_receipts"] = []any{}
	mutated, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("encode mutated call record: %v", err)
	}
	mutated = append(mutated, '\n')
	if err := os.WriteFile(callRecordPath, mutated, 0o600); err != nil {
		t.Fatalf("write mutated call record: %v", err)
	}

	_, err = BuildCallPack(BuildCallOptions{
		CallRecordPath: callRecordPath,
		OutputPath:     filepath.Join(workDir, "callpack.zip"),
	})
	if err == nil {
		t.Fatalf("expected callpack build to fail without speak receipt for gated tts.emitted")
	}
	if !strings.Contains(err.Error(), "missing speak receipt") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeCallRecordFixture(t *testing.T, dir string, runpackPath string, callID string) string {
	t.Helper()
	createdAt := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	record := map[string]any{
		"schema_id":        callRecordSchemaID,
		"schema_version":   callRecordSchemaVersion,
		"created_at":       createdAt.Format(time.RFC3339Nano),
		"producer_version": "test",
		"call_id":          callID,
		"runpack_path":     runpackPath,
		"privacy_mode":     "hash_only",
		"events": []map[string]any{
			{"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": createdAt.Format(time.RFC3339Nano), "call_id": callID, "call_seq": 1, "turn_index": 1, "event_type": "asr.final", "payload_digest": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			{"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": createdAt.Add(1 * time.Second).Format(time.RFC3339Nano), "call_id": callID, "call_seq": 2, "turn_index": 2, "event_type": "commitment.declared", "commitment_class": "quote"},
			{"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": createdAt.Add(2 * time.Second).Format(time.RFC3339Nano), "call_id": callID, "call_seq": 3, "turn_index": 2, "event_type": "gate.decision", "commitment_class": "quote", "intent_digest": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "policy_digest": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"},
			{"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": createdAt.Add(3 * time.Second).Format(time.RFC3339Nano), "call_id": callID, "call_seq": 4, "turn_index": 2, "event_type": "tts.request", "commitment_class": "quote"},
			{"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": createdAt.Add(4 * time.Second).Format(time.RFC3339Nano), "call_id": callID, "call_seq": 5, "turn_index": 2, "event_type": "tts.emitted", "commitment_class": "quote", "say_token_id": "say_demo"},
			{"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": createdAt.Add(5 * time.Second).Format(time.RFC3339Nano), "call_id": callID, "call_seq": 6, "turn_index": 2, "event_type": "tool.intent"},
			{"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": createdAt.Add(6 * time.Second).Format(time.RFC3339Nano), "call_id": callID, "call_seq": 7, "turn_index": 2, "event_type": "tool.result"},
			{"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": createdAt.Add(7 * time.Second).Format(time.RFC3339Nano), "call_id": callID, "call_seq": 8, "turn_index": 2, "event_type": "approval.granted", "commitment_class": "quote"},
		},
		"commitments": []schemavoice.CommitmentIntent{
			{
				SchemaID:        "gait.voice.commitment_intent",
				SchemaVersion:   "1.0.0",
				CreatedAt:       createdAt.Add(1 * time.Second),
				ProducerVersion: "test",
				CallID:          callID,
				TurnIndex:       2,
				CallSeq:         2,
				CommitmentClass: "quote",
				Context: schemavoice.CommitmentContext{
					Identity:  "agent.voice",
					Workspace: "/srv/voice",
					RiskClass: "high",
				},
				QuoteMinCents: 1000,
				QuoteMaxCents: 1200,
			},
		},
		"gate_decisions": []schemavoice.GateDecision{
			{
				CallID:          callID,
				CallSeq:         3,
				TurnIndex:       2,
				CommitmentClass: "quote",
				Verdict:         "allow",
				IntentDigest:    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				PolicyDigest:    "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				ReasonCodes:     []string{"allow_rule"},
			},
		},
		"speak_receipts": []schemavoice.SpeakReceipt{
			{
				CallID:          callID,
				CallSeq:         5,
				TurnIndex:       2,
				CommitmentClass: "quote",
				SayTokenID:      "say_demo",
				SpokenDigest:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				EmittedAt:       createdAt.Add(4 * time.Second),
			},
		},
		"reference_digests": []schemavoice.ReferenceDigest{
			{RefID: "ref_1", SHA256: "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"},
		},
	}
	recordPath := filepath.Join(dir, "call_record.json")
	content, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("marshal call record: %v", err)
	}
	content = append(content, '\n')
	if err := os.WriteFile(recordPath, content, 0o600); err != nil {
		t.Fatalf("write call record: %v", err)
	}
	return recordPath
}
