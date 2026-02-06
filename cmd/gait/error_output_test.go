package main

import (
	stderrors "errors"
	"strings"
	"testing"

	coreerrors "github.com/davidahmann/gait/core/errors"
)

func TestMarshalOutputWithErrorEnvelope(t *testing.T) {
	setCurrentCorrelationID("cid-test")
	t.Cleanup(func() {
		setCurrentCorrelationID("")
	})
	payload := map[string]any{
		"ok":    false,
		"error": "boom",
	}
	encoded, err := marshalOutputWithErrorEnvelope(payload, exitInvalidInput)
	if err != nil {
		t.Fatalf("marshalOutputWithErrorEnvelope error: %v", err)
	}
	result := string(encoded)
	if !strings.Contains(result, `"error_code":"invalid_input"`) {
		t.Fatalf("missing error_code in output: %s", result)
	}
	if !strings.Contains(result, `"error_category":"invalid_input"`) {
		t.Fatalf("missing error_category in output: %s", result)
	}
	if !strings.Contains(result, `"retryable":false`) {
		t.Fatalf("missing retryable in output: %s", result)
	}
	if !strings.Contains(result, `"hint":"check command usage and input schema"`) {
		t.Fatalf("missing hint in output: %s", result)
	}
	if !strings.Contains(result, `"correlation_id":"cid-test"`) {
		t.Fatalf("missing correlation id in output: %s", result)
	}
}

func TestMarshalOutputWithCorrelationForSuccess(t *testing.T) {
	setCurrentCorrelationID("cid-success")
	t.Cleanup(func() {
		setCurrentCorrelationID("")
	})
	payload := map[string]any{"ok": true}
	encoded, err := marshalOutputWithErrorEnvelope(payload, exitOK)
	if err != nil {
		t.Fatalf("marshalOutputWithErrorEnvelope error: %v", err)
	}
	result := string(encoded)
	if !strings.Contains(result, `"correlation_id":"cid-success"`) {
		t.Fatalf("missing correlation_id for success output: %s", result)
	}
}

func TestExitCodeForError(t *testing.T) {
	if got := exitCodeForError(stderrors.New("plain"), exitInvalidInput); got != exitInvalidInput {
		t.Fatalf("expected fallback invalid-input exit, got %d", got)
	}
	wrapped := coreerrors.Wrap(stderrors.New("approval"), coreerrors.CategoryApprovalRequired, "approval_required", "", false)
	if got := exitCodeForError(wrapped, exitInvalidInput); got != exitApprovalRequired {
		t.Fatalf("expected approval-required exit, got %d", got)
	}
	if got := exitCodeForError(stderrors.New("socket timeout"), exitInvalidInput); got != exitInternalFailure {
		t.Fatalf("expected internal failure exit, got %d", got)
	}
}
