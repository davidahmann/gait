package main

import (
	"encoding/json"
	"path/filepath"
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
