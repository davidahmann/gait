package main

import (
	"testing"
)

func TestCorrelationIDHelpers(t *testing.T) {
	idA := newCorrelationID([]string{"gait", "policy", "test", "a", "--json"})
	idB := newCorrelationID([]string{"gait", "policy", "test", "a", "--json"})
	idC := newCorrelationID([]string{"gait", "policy", "test", "b", "--json"})
	if len(idA) != 24 {
		t.Fatalf("unexpected correlation id length: %s", idA)
	}
	if idA != idB {
		t.Fatalf("expected deterministic correlation ids for same input")
	}
	if idA == idC {
		t.Fatalf("expected different correlation ids for different inputs")
	}
	setCurrentCorrelationID(" cid ")
	if got := currentCorrelationID(); got != "cid" {
		t.Fatalf("unexpected current correlation id: %q", got)
	}
	setCurrentCorrelationID("")
}
