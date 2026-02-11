#!/usr/bin/env bash
set -euo pipefail

echo "[session-soak] append/checkpoint/verify loop (count=8)"
go test ./core/runpack -run 'TestSessionJournalLifecycleAndCheckpointChain|TestSessionConcurrentAppendsAreDeterministic|TestSessionLockRecoveryAndHelpers' -count=8

echo "[session-soak] integration contention (count=5)"
go test ./internal/integration -run 'TestConcurrentSessionAppendStateIsDeterministic' -count=5
