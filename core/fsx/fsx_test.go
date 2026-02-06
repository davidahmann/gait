package fsx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicCreatesAndOverwrites(t *testing.T) {
	target := filepath.Join(t.TempDir(), "state.json")

	if err := WriteFileAtomic(target, []byte("first\n"), 0o600); err != nil {
		t.Fatalf("first write: %v", err)
	}
	first, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read first write: %v", err)
	}
	if string(first) != "first\n" {
		t.Fatalf("unexpected first content: %q", string(first))
	}

	if err := WriteFileAtomic(target, []byte("second\n"), 0o600); err != nil {
		t.Fatalf("second write: %v", err)
	}
	second, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read second write: %v", err)
	}
	if string(second) != "second\n" {
		t.Fatalf("unexpected second content: %q", string(second))
	}
}

func TestWriteFileAtomicMode(t *testing.T) {
	target := filepath.Join(t.TempDir(), "secure.json")

	if err := WriteFileAtomic(target, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected mode 0600 got %#o", info.Mode().Perm())
	}
}
