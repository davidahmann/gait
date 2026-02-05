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

func TestValidateJSON(t *testing.T) {
	root := repoRoot(t)
	schema := filepath.Join(root, "schemas", "v1", "runpack", "manifest.schema.json")
	valid := []byte(`{
		"schema_id":"gait.runpack.manifest",
		"schema_version":"1.0.0",
		"created_at":"2026-02-05T00:00:00Z",
		"producer_version":"0.0.0-dev",
		"run_id":"run_demo",
		"capture_mode":"reference",
		"files":[{"path":"run.json","sha256":"0000000000000000000000000000000000000000000000000000000000000000"}],
		"manifest_digest":"1111111111111111111111111111111111111111111111111111111111111111"
	}`)
	invalid := []byte(`{`)

	if err := ValidateJSON(schema, valid); err != nil {
		t.Fatalf("expected valid json, got error: %v", err)
	}
	if err := ValidateJSON(schema, invalid); err == nil {
		t.Fatalf("expected invalid json to fail")
	}
}

func TestValidateJSONL(t *testing.T) {
	root := repoRoot(t)
	schema := filepath.Join(root, "schemas", "v1", "runpack", "intent.schema.json")
	data := []byte("\n" +
		`{"schema_id":"gait.runpack.intent","schema_version":"1.0.0","created_at":"2026-02-05T00:00:00Z","producer_version":"0.0.0-dev","run_id":"run_demo","intent_id":"intent_1","tool_name":"tool.demo","args_digest":"2222222222222222222222222222222222222222222222222222222222222222","args":{"foo":"bar"}}` +
		"\n")
	if err := ValidateJSONL(schema, data); err != nil {
		t.Fatalf("expected valid jsonl, got error: %v", err)
	}
}

func TestValidateSchemaMissing(t *testing.T) {
	err := ValidateJSONFile("does-not-exist.json", "also-missing.json")
	if err == nil {
		t.Fatalf("expected error for missing schema file")
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
