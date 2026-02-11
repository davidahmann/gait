#!/usr/bin/env bash
set -euo pipefail

echo "[session-soak] append/checkpoint/verify loop (count=8)"
go test ./core/runpack -run 'TestSessionJournalLifecycleAndCheckpointChain|TestSessionConcurrentAppendsAreDeterministic|TestSessionLockRecoveryAndHelpers|TestSessionCompactionPreservesCheckpointVerification|TestSessionAppendLatencyDriftBudget' -count=4

echo "[session-soak] integration contention (count=5)"
GAIT_SESSION_LOCK_PROFILE=swarm GAIT_SESSION_LOCK_TIMEOUT=5s GAIT_SESSION_LOCK_RETRY=10ms go test ./internal/integration -run 'TestConcurrentSessionAppendStateIsDeterministic|TestSessionSwarmContentionBudget' -count=3
