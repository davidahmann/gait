package main

import (
	"path/filepath"
	"testing"
)

func TestRunDelegateMintAndVerify(t *testing.T) {
	workDir := t.TempDir()
	withWorkingDir(t, workDir)

	privateKeyPath := filepath.Join(workDir, "delegate_private.key")
	writePrivateKey(t, privateKeyPath)
	tokenPath := filepath.Join(workDir, "delegation_token.json")

	if code := runDelegate([]string{
		"mint",
		"--delegator", "agent.lead",
		"--delegate", "agent.specialist",
		"--scope", "tool:tool.write",
		"--ttl", "1h",
		"--key-mode", "prod",
		"--private-key", privateKeyPath,
		"--out", tokenPath,
		"--json",
	}); code != exitOK {
		t.Fatalf("runDelegate mint expected %d got %d", exitOK, code)
	}

	if code := runDelegate([]string{
		"verify",
		"--token", tokenPath,
		"--delegator", "agent.lead",
		"--delegate", "agent.specialist",
		"--scope", "tool:tool.write",
		"--private-key", privateKeyPath,
		"--json",
	}); code != exitOK {
		t.Fatalf("runDelegate verify expected %d got %d", exitOK, code)
	}
}

func TestRunDelegateInvalidArgs(t *testing.T) {
	if code := runDelegate([]string{}); code != exitInvalidInput {
		t.Fatalf("runDelegate missing args expected %d got %d", exitInvalidInput, code)
	}
	if code := runDelegate([]string{"mint"}); code != exitInvalidInput {
		t.Fatalf("runDelegate mint missing args expected %d got %d", exitInvalidInput, code)
	}
	if code := runDelegate([]string{"verify"}); code != exitInvalidInput {
		t.Fatalf("runDelegate verify missing args expected %d got %d", exitInvalidInput, code)
	}
}
