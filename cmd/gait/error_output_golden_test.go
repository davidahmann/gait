package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarshalOutputErrorEnvelopeGolden(t *testing.T) {
	cases := []struct {
		name          string
		payload       map[string]any
		exitCode      int
		correlationID string
		fixture       string
		expectCode    string
		expectCat     string
	}{
		{
			name:          "invalid_input",
			payload:       map[string]any{"ok": false, "error": "missing required --policy"},
			exitCode:      exitInvalidInput,
			correlationID: "cid-golden-invalid",
			fixture:       "error_envelope_invalid_input.golden.json",
			expectCode:    "invalid_input",
			expectCat:     "invalid_input",
		},
		{
			name:          "internal_failure",
			payload:       map[string]any{"ok": false, "error": "unexpected failure"},
			exitCode:      exitInternalFailure,
			correlationID: "cid-golden-internal",
			fixture:       "error_envelope_internal_failure.golden.json",
			expectCode:    "internal_failure",
			expectCat:     "internal_failure",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			setCurrentCorrelationID(testCase.correlationID)
			t.Cleanup(func() {
				setCurrentCorrelationID("")
			})
			encoded, err := marshalOutputWithErrorEnvelope(testCase.payload, testCase.exitCode)
			if err != nil {
				t.Fatalf("marshalOutputWithErrorEnvelope error: %v", err)
			}
			fixturePath := filepath.Join("testdata", testCase.fixture)
			expected, err := os.ReadFile(fixturePath) // #nosec G304 -- static local test fixture path.
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			if strings.TrimSpace(string(encoded)) != strings.TrimSpace(string(expected)) {
				t.Fatalf("golden mismatch for %s\nexpected=%s\nactual=%s", fixturePath, string(expected), string(encoded))
			}

			var decoded map[string]any
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("unmarshal encoded payload: %v", err)
			}
			if decoded["error_code"] != testCase.expectCode {
				t.Fatalf("unexpected error_code: %#v", decoded["error_code"])
			}
			if decoded["error_category"] != testCase.expectCat {
				t.Fatalf("unexpected error_category: %#v", decoded["error_category"])
			}
		})
	}
}
