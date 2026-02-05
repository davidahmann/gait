package validate

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidateJSONFile(t *testing.T) {
	root := repoRoot(t)
	schema := filepath.Join(root, "schemas", "v1", "runpack", "manifest.schema.json")
	valid := filepath.Join(root, "core", "schema", "testdata", "manifest_valid.json")
	invalid := filepath.Join(root, "core", "schema", "testdata", "manifest_invalid.json")

	if err := ValidateJSONFile(schema, valid); err != nil {
		t.Fatalf("expected valid manifest, got error: %v", err)
	}
	if err := ValidateJSONFile(schema, invalid); err == nil {
		t.Fatalf("expected invalid manifest to fail")
	}
}

func TestValidateJSONLFile(t *testing.T) {
	root := repoRoot(t)
	schema := filepath.Join(root, "schemas", "v1", "runpack", "intent.schema.json")
	valid := filepath.Join(root, "core", "schema", "testdata", "intent_valid.jsonl")
	invalid := filepath.Join(root, "core", "schema", "testdata", "intent_invalid.jsonl")

	if err := ValidateJSONLFile(schema, valid); err != nil {
		t.Fatalf("expected valid jsonl, got error: %v", err)
	}
	if err := ValidateJSONLFile(schema, invalid); err == nil {
		t.Fatalf("expected invalid jsonl to fail")
	}
}

func repoRoot(t *testing.T) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("unable to locate test file")
	}
	dir := filepath.Dir(filename)
	return filepath.Clean(filepath.Join(dir, "..", "..", ".."))
}
