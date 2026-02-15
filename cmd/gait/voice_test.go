package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunVoiceTokenMintAndVerify(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)
	privateKeyPath := filepath.Join(workDir, "voice_private.key")
	writePrivateKey(t, privateKeyPath)

	policyPath := filepath.Join(workDir, "policy_voice.yaml")
	mustWriteFile(t, policyPath, "default_verdict: allow\n")

	intentPath := filepath.Join(workDir, "commitment_intent.json")
	createdAt := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	mustWriteFile(t, intentPath, `{
  "schema_id": "gait.voice.commitment_intent",
  "schema_version": "1.0.0",
  "created_at": "`+createdAt+`",
  "producer_version": "test",
  "call_id": "call_voice_cli",
  "turn_index": 2,
  "call_seq": 5,
  "commitment_class": "quote",
  "quote_min_cents": 1000,
  "quote_max_cents": 1200,
  "context": {
    "identity": "agent.voice",
    "workspace": "/srv/voice",
    "risk_class": "high"
  }
}
`)

	var mintCode int
	mintRaw := captureStdout(t, func() {
		mintCode = runVoice([]string{
			"token",
			"mint",
			"--intent", intentPath,
			"--policy", policyPath,
			"--ttl", "2m",
			"--private-key", privateKeyPath,
			"--key-mode", "prod",
			"--json",
		})
	})
	if mintCode != exitOK {
		t.Fatalf("voice token mint expected %d got %d raw=%q", exitOK, mintCode, mintRaw)
	}
	var mintOut voiceTokenOutput
	if err := json.Unmarshal([]byte(mintRaw), &mintOut); err != nil {
		t.Fatalf("decode mint output: %v raw=%q", err, mintRaw)
	}
	if mintOut.TokenPath == "" || mintOut.TokenID == "" {
		t.Fatalf("unexpected mint output: %#v", mintOut)
	}

	verifyCode := runVoice([]string{
		"token",
		"verify",
		"--token", mintOut.TokenPath,
		"--private-key", privateKeyPath,
		"--intent-digest", mintOut.IntentDigest,
		"--policy-digest", mintOut.PolicyDigest,
		"--call-id", "call_voice_cli",
		"--turn-index", "2",
		"--call-seq", "5",
		"--commitment-class", "quote",
		"--json",
	})
	if verifyCode != exitOK {
		t.Fatalf("voice token verify expected %d got %d", exitOK, verifyCode)
	}
}

func TestRunVoicePackLifecycle(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)
	if code := runDemo(nil); code != exitOK {
		t.Fatalf("demo expected %d got %d", exitOK, code)
	}
	runpackPath, err := resolveRunpackPath("run_demo")
	if err != nil {
		t.Fatalf("resolve runpack: %v", err)
	}
	callRecordPath := writeVoiceCallRecordForPackCLI(t, workDir, runpackPath, "call_voice_pack")

	var buildCode int
	buildRaw := captureStdout(t, func() {
		buildCode = runVoice([]string{
			"pack",
			"build",
			"--from", callRecordPath,
			"--json",
		})
	})
	if buildCode != exitOK {
		t.Fatalf("voice pack build expected %d got %d raw=%q", exitOK, buildCode, buildRaw)
	}
	var buildOut packOutput
	if err := json.Unmarshal([]byte(buildRaw), &buildOut); err != nil {
		t.Fatalf("decode build output: %v raw=%q", err, buildRaw)
	}
	if buildOut.PackType != "call" || buildOut.Path == "" {
		t.Fatalf("unexpected voice pack build output: %#v", buildOut)
	}

	verifyCode := runVoice([]string{"pack", "verify", buildOut.Path, "--json"})
	if verifyCode != exitOK {
		t.Fatalf("voice pack verify expected %d got %d", exitOK, verifyCode)
	}
	inspectCode := runVoice([]string{"pack", "inspect", buildOut.Path, "--json"})
	if inspectCode != exitOK {
		t.Fatalf("voice pack inspect expected %d got %d", exitOK, inspectCode)
	}
}

func TestRunVoiceInvalidArgs(t *testing.T) {
	if code := runVoice([]string{}); code != exitInvalidInput {
		t.Fatalf("runVoice missing args expected %d got %d", exitInvalidInput, code)
	}
	if code := runVoice([]string{"token"}); code != exitInvalidInput {
		t.Fatalf("runVoice token missing args expected %d got %d", exitInvalidInput, code)
	}
	if code := runVoice([]string{"pack"}); code != exitInvalidInput {
		t.Fatalf("runVoice pack missing args expected %d got %d", exitInvalidInput, code)
	}
}

func TestRunVoiceExplainAndHelpPaths(t *testing.T) {
	if code := runVoice([]string{"--explain"}); code != exitOK {
		t.Fatalf("runVoice explain expected %d got %d", exitOK, code)
	}
	if code := runVoicePack([]string{"--explain"}); code != exitOK {
		t.Fatalf("runVoicePack explain expected %d got %d", exitOK, code)
	}
	if code := runVoiceToken([]string{"--explain"}); code != exitOK {
		t.Fatalf("runVoiceToken explain expected %d got %d", exitOK, code)
	}
	if code := runVoiceTokenMint([]string{"--help"}); code != exitOK {
		t.Fatalf("runVoiceTokenMint help expected %d got %d", exitOK, code)
	}
	if code := runVoiceTokenVerify([]string{"--help"}); code != exitOK {
		t.Fatalf("runVoiceTokenVerify help expected %d got %d", exitOK, code)
	}
}

func TestRunVoiceTokenMintNonAllowVerdicts(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)
	privateKeyPath := filepath.Join(workDir, "voice_private.key")
	writePrivateKey(t, privateKeyPath)

	intentPath := filepath.Join(workDir, "commitment_intent.json")
	createdAt := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	mustWriteFile(t, intentPath, `{
  "schema_id": "gait.voice.commitment_intent",
  "schema_version": "1.0.0",
  "created_at": "`+createdAt+`",
  "producer_version": "test",
  "call_id": "call_voice_cli",
  "turn_index": 2,
  "call_seq": 5,
  "commitment_class": "quote",
  "quote_min_cents": 1000,
  "quote_max_cents": 1200,
  "context": {"identity": "agent.voice", "workspace": "/srv/voice", "risk_class": "high"}
}
`)

	blockPolicyPath := filepath.Join(workDir, "policy_block.yaml")
	mustWriteFile(t, blockPolicyPath, "default_verdict: block\n")
	blockRaw := captureStdout(t, func() {
		code := runVoice([]string{
			"token", "mint",
			"--intent", intentPath,
			"--policy", blockPolicyPath,
			"--private-key", privateKeyPath,
			"--key-mode", "prod",
			"--json",
		})
		if code != exitPolicyBlocked {
			t.Fatalf("block verdict expected %d got %d", exitPolicyBlocked, code)
		}
	})
	var blockOut voiceTokenOutput
	if err := json.Unmarshal([]byte(blockRaw), &blockOut); err != nil {
		t.Fatalf("decode block output: %v raw=%q", err, blockRaw)
	}
	if blockOut.Verdict != "block" || blockOut.TracePath == "" {
		t.Fatalf("unexpected block output: %#v", blockOut)
	}

	approvalPolicyPath := filepath.Join(workDir, "policy_require_approval.yaml")
	mustWriteFile(t, approvalPolicyPath, "default_verdict: require_approval\n")
	approvalRaw := captureStdout(t, func() {
		code := runVoice([]string{
			"token", "mint",
			"--intent", intentPath,
			"--policy", approvalPolicyPath,
			"--private-key", privateKeyPath,
			"--key-mode", "prod",
			"--json",
		})
		if code != exitApprovalRequired {
			t.Fatalf("require_approval verdict expected %d got %d", exitApprovalRequired, code)
		}
	})
	var approvalOut voiceTokenOutput
	if err := json.Unmarshal([]byte(approvalRaw), &approvalOut); err != nil {
		t.Fatalf("decode approval output: %v raw=%q", err, approvalRaw)
	}
	if approvalOut.Verdict != "require_approval" || approvalOut.TracePath == "" {
		t.Fatalf("unexpected require_approval output: %#v", approvalOut)
	}
}

func TestRunVoiceTokenVerifyFailures(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)
	privateKeyPath := filepath.Join(workDir, "voice_private.key")
	writePrivateKey(t, privateKeyPath)

	policyPath := filepath.Join(workDir, "policy_voice.yaml")
	mustWriteFile(t, policyPath, "default_verdict: allow\n")

	intentPath := filepath.Join(workDir, "commitment_intent.json")
	createdAt := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	mustWriteFile(t, intentPath, `{
  "schema_id": "gait.voice.commitment_intent",
  "schema_version": "1.0.0",
  "created_at": "`+createdAt+`",
  "producer_version": "test",
  "call_id": "call_voice_cli",
  "turn_index": 2,
  "call_seq": 5,
  "commitment_class": "quote",
  "quote_min_cents": 1000,
  "quote_max_cents": 1200,
  "context": {"identity": "agent.voice", "workspace": "/srv/voice", "risk_class": "high"}
}
`)

	mintRaw := captureStdout(t, func() {
		code := runVoice([]string{
			"token", "mint",
			"--intent", intentPath,
			"--policy", policyPath,
			"--private-key", privateKeyPath,
			"--key-mode", "prod",
			"--json",
		})
		if code != exitOK {
			t.Fatalf("mint expected %d got %d", exitOK, code)
		}
	})
	var mintOut voiceTokenOutput
	if err := json.Unmarshal([]byte(mintRaw), &mintOut); err != nil {
		t.Fatalf("decode mint output: %v raw=%q", err, mintRaw)
	}

	failRaw := captureStdout(t, func() {
		code := runVoice([]string{
			"token", "verify",
			"--token", mintOut.TokenPath,
			"--private-key", privateKeyPath,
			"--call-id", "other_call",
			"--json",
		})
		if code != exitVerifyFailed {
			t.Fatalf("verify mismatch expected %d got %d", exitVerifyFailed, code)
		}
	})
	var failOut voiceTokenOutput
	if err := json.Unmarshal([]byte(failRaw), &failOut); err != nil {
		t.Fatalf("decode verify failure output: %v raw=%q", err, failRaw)
	}
	if failOut.ErrorCode != "say_token_call_binding_mismatch" {
		t.Fatalf("unexpected verify failure error code: %#v", failOut)
	}

	missingTokenRaw := captureStdout(t, func() {
		code := runVoice([]string{"token", "verify", "--json"})
		if code != exitInvalidInput {
			t.Fatalf("verify missing token expected %d got %d", exitInvalidInput, code)
		}
	})
	if !strings.Contains(missingTokenRaw, "--token is required") {
		t.Fatalf("expected required token message, got %q", missingTokenRaw)
	}
}

func TestVoiceTokenOutputAndIntentReadErrors(t *testing.T) {
	if code := writeVoiceTokenOutput(false, voiceTokenOutput{OK: true, Operation: "verify", TokenPath: "token.json"}, exitOK); code != exitOK {
		t.Fatalf("writeVoiceTokenOutput success expected %d got %d", exitOK, code)
	}
	if code := writeVoiceTokenOutput(false, voiceTokenOutput{OK: false, Operation: "verify", Error: "bad"}, exitVerifyFailed); code != exitVerifyFailed {
		t.Fatalf("writeVoiceTokenOutput failure expected %d got %d", exitVerifyFailed, code)
	}

	if _, err := readVoiceCommitmentIntent(filepath.Join(t.TempDir(), "missing.json")); err == nil || !strings.Contains(err.Error(), "read voice intent") {
		t.Fatalf("expected read error from missing intent file, got %v", err)
	}

	workDir := t.TempDir()
	badIntentPath := filepath.Join(workDir, "bad_intent.json")
	mustWriteFile(t, badIntentPath, "{")
	if _, err := readVoiceCommitmentIntent(badIntentPath); err == nil || !strings.Contains(err.Error(), "parse voice intent json") {
		t.Fatalf("expected parse error from malformed intent file, got %v", err)
	}
}
