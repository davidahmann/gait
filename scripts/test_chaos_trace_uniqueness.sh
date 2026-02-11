#!/usr/bin/env bash
set -euo pipefail

echo "[chaos-trace] runtime event identity uniqueness"
go test ./core/gate -run 'TestEmitSignedTraceRuntimeEventIdentity' -count=5

echo "[chaos-trace] proxy default trace path uniqueness"
go test ./cmd/gait -run 'TestRunMCPProxyDefaultTracePathIsUniquePerEmission' -count=3
