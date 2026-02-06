package integration

import (
	"sync"
	"testing"
	"time"

	coreerrors "github.com/davidahmann/gait/core/errors"
	"github.com/davidahmann/gait/core/gate"
	schemagate "github.com/davidahmann/gait/core/schema/v1/gate"
)

func TestConcurrentGateRateLimitStateIsDeterministic(t *testing.T) {
	workDir := t.TempDir()
	statePath := workDir + "/rate_state.json"
	policy, err := gate.ParsePolicyYAML([]byte(`
default_verdict: allow
rules:
  - name: high-risk-write
    effect: allow
    match:
      tool_names: [tool.write]
    rate_limit:
      requests: 2
      scope: tool_identity
      window: minute
`))
	if err != nil {
		t.Fatalf("parse policy: %v", err)
	}
	intent := schemagate.IntentRequest{
		SchemaID:        "gait.gate.intent_request",
		SchemaVersion:   "1.0.0",
		CreatedAt:       time.Date(2026, time.February, 6, 0, 0, 0, 0, time.UTC),
		ProducerVersion: "0.0.0-test",
		ToolName:        "tool.write",
		Args:            map[string]any{"path": "/tmp/a.txt"},
		Context: schemagate.IntentContext{
			Identity:  "alice",
			Workspace: "/repo/gait",
			RiskClass: "high",
		},
	}

	const workers = 10
	var allowCount int
	var blockCount int
	var contentionCount int
	var mutex sync.Mutex
	var group sync.WaitGroup
	group.Add(workers)
	now := time.Date(2026, time.February, 6, 12, 0, 0, 0, time.UTC)

	for i := 0; i < workers; i++ {
		go func() {
			defer group.Done()
			outcome, evalErr := gate.EvaluatePolicyDetailed(policy, intent, gate.EvalOptions{ProducerVersion: "0.0.0-test"})
			if evalErr != nil {
				t.Errorf("evaluate policy: %v", evalErr)
				return
			}
			decision, enforceErr := gate.EnforceRateLimit(statePath, outcome.RateLimit, intent, now)
			if enforceErr != nil {
				if coreerrors.CategoryOf(enforceErr) == coreerrors.CategoryStateContention && coreerrors.RetryableOf(enforceErr) {
					mutex.Lock()
					blockCount++
					contentionCount++
					mutex.Unlock()
					return
				}
				t.Errorf("enforce rate limit: %v", enforceErr)
				return
			}

			mutex.Lock()
			if decision.Allowed {
				allowCount++
			} else {
				blockCount++
			}
			mutex.Unlock()
		}()
	}

	group.Wait()
	if allowCount == 0 {
		t.Fatalf("expected at least one allowed operation, got 0 (blocked=%d contention=%d)", blockCount, contentionCount)
	}
	if allowCount > 2 {
		t.Fatalf("expected at most 2 allowed operations, got %d (blocked=%d contention=%d)", allowCount, blockCount, contentionCount)
	}
	if allowCount+blockCount != workers {
		t.Fatalf("expected all workers accounted for, allowed=%d blocked=%d workers=%d", allowCount, blockCount, workers)
	}
}
