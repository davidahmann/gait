#!/usr/bin/env bash
set -euo pipefail

echo "[chaos-payload] oversized body rejection"
go test ./cmd/gait -run 'TestMCPServeHandlerRejectsOversizedRequest' -count=5
