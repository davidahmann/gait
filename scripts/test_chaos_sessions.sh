#!/usr/bin/env bash
set -euo pipefail

echo "[chaos-sessions] swarm contention budget"
go test ./internal/integration -run 'TestSessionSwarmContentionBudget' -count=3

echo "[chaos-sessions] session append latency drift + compaction"
go test ./core/runpack -run 'TestSessionAppendLatencyDriftBudget|TestSessionCompactionPreservesCheckpointVerification' -count=2
