package scout

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotCoverageAndDiff(t *testing.T) {
	workDir := t.TempDir()
	pythonPath := filepath.Join(workDir, "agent.py")
	pythonSource := `
from langchain.tools import tool
from agents import function_tool

@tool
def delete_user():
    return "ok"

@function_tool
def list_orders():
    return "ok"
`
	if err := os.WriteFile(pythonPath, []byte(pythonSource), 0o600); err != nil {
		t.Fatalf("write python source: %v", err)
	}
	mcpPath := filepath.Join(workDir, "mcp.json")
	mcpSource := `{"mcpServers":{"database":{"command":"db"},"email":{"command":"mail"}}}`
	if err := os.WriteFile(mcpPath, []byte(mcpSource), 0o600); err != nil {
		t.Fatalf("write mcp config: %v", err)
	}

	provider := DefaultProvider{}
	snapshot, err := provider.Snapshot(context.Background(), SnapshotRequest{Roots: []string{workDir}})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snapshot.SchemaID != "gait.scout.inventory_snapshot" {
		t.Fatalf("unexpected schema id: %s", snapshot.SchemaID)
	}
	if len(snapshot.Items) < 4 {
		t.Fatalf("expected at least 4 inventory items, got %d", len(snapshot.Items))
	}

	policyPath := filepath.Join(workDir, "policy.yaml")
	policySource := `default_verdict: block
rules:
  - name: allow-delete
    priority: 1
    effect: allow
    match:
      tool_names: [delete_user]
`
	if err := os.WriteFile(policyPath, []byte(policySource), 0o600); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	coverage, err := BuildCoverage(snapshot, []string{policyPath})
	if err != nil {
		t.Fatalf("coverage: %v", err)
	}
	if coverage.DiscoveredTools < 3 {
		t.Fatalf("expected discovered tools >= 3 got %d", coverage.DiscoveredTools)
	}
	if coverage.GatedTools != 1 {
		t.Fatalf("expected gated tools 1 got %d", coverage.GatedTools)
	}

	left := snapshot
	right := snapshot
	right.SnapshotID = "snap_other"
	if len(right.Items) == 0 {
		t.Fatalf("expected inventory items")
	}
	right.Items = append(right.Items[:0:0], right.Items...)
	right.Items[0].Locator = right.Items[0].Locator + ":99"
	diff := DiffSnapshots(left, right)
	if diff.ChangedCount != 1 {
		t.Fatalf("expected 1 changed item, got %d", diff.ChangedCount)
	}
}
