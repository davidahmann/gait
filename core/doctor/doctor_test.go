package doctor

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/davidahmann/gait/core/sign"
)

func TestRunDetectsMissingSchemasAsNonFixable(t *testing.T) {
	workDir := t.TempDir()
	result := Run(Options{
		WorkDir:         workDir,
		OutputDir:       filepath.Join(workDir, "gait-out"),
		ProducerVersion: "test",
		KeyMode:         sign.ModeDev,
	})

	if result.Status != statusFail {
		t.Fatalf("expected fail status, got: %s", result.Status)
	}
	if !result.NonFixable {
		t.Fatalf("expected non-fixable result")
	}
	if !checkStatus(result.Checks, "schema_files", statusFail) {
		t.Fatalf("expected schema_files fail check")
	}
}

func TestRunPassesWithValidWorkspaceAndSchemas(t *testing.T) {
	root := repoRoot(t)
	outputDir := filepath.Join(t.TempDir(), "gait-out")
	if err := ensureDir(outputDir); err != nil {
		t.Fatalf("create output dir: %v", err)
	}

	result := Run(Options{
		WorkDir:         root,
		OutputDir:       outputDir,
		ProducerVersion: "test",
		KeyMode:         sign.ModeDev,
	})

	if result.Status != statusPass {
		t.Fatalf("expected pass status, got: %s (%s)", result.Status, result.Summary)
	}
	if result.NonFixable {
		t.Fatalf("expected non-fixable to be false")
	}
	if len(result.Checks) != 4 {
		t.Fatalf("unexpected checks count: %d", len(result.Checks))
	}
}

func TestRunDetectsProdKeyConfigFailure(t *testing.T) {
	root := repoRoot(t)
	outputDir := filepath.Join(t.TempDir(), "gait-out")
	if err := ensureDir(outputDir); err != nil {
		t.Fatalf("create output dir: %v", err)
	}

	result := Run(Options{
		WorkDir:         root,
		OutputDir:       outputDir,
		ProducerVersion: "test",
		KeyMode:         sign.ModeProd,
	})

	if result.Status != statusFail {
		t.Fatalf("expected fail status for prod key failure, got: %s", result.Status)
	}
	if result.NonFixable {
		t.Fatalf("expected fixable failure for key config")
	}
	if !checkStatus(result.Checks, "key_config", statusFail) {
		t.Fatalf("expected key_config fail check")
	}
}

func checkStatus(checks []Check, name string, status string) bool {
	for _, check := range checks {
		if check.Name == name && check.Status == status {
			return true
		}
	}
	return false
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o750)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("unable to locate test file")
	}
	dir := filepath.Dir(filename)
	return filepath.Clean(filepath.Join(dir, "..", ".."))
}
