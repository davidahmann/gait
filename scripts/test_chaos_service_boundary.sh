#!/usr/bin/env bash
set -euo pipefail

echo "[chaos-service] auth, path override, strict verdict semantics"
go test ./cmd/gait -run 'TestMCPServeHandlerRequiresBearerAuthorizationWhenEnabled|TestMCPServeHandlerAcceptsBearerAuthorizationWhenEnabled|TestMCPServeHandlerRejectsClientArtifactPathOverridesByDefault|TestMCPServeHandlerStrictVerdictStatusForBlock|TestMCPServeHandlerTraceRetentionMaxCount|TestMCPServeHandlerTraceRetentionMaxAge' -count=2
