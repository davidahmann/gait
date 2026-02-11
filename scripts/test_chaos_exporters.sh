#!/usr/bin/env bash
set -euo pipefail

echo "[chaos-exporters] mcp exporter concurrent append integrity"
go test ./core/mcp -run 'TestAppendJSONLConcurrentIntegrity' -count=3

echo "[chaos-exporters] scout adoption/operational concurrent append integrity"
go test ./core/scout -run 'TestAppendAdoptionEventConcurrentIntegrity|TestAppendOperationalEventConcurrentIntegrity' -count=3
