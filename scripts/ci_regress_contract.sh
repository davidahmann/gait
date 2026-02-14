#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

BIN_PATH="${GAIT_BIN:-$REPO_ROOT/gait}"
ARTIFACT_ROOT="${ARTIFACT_ROOT:-$REPO_ROOT/gait-out/adoption_regress}"

if [[ ! -x "$BIN_PATH" ]]; then
  go build -o "$BIN_PATH" ./cmd/gait
fi

mkdir -p "$ARTIFACT_ROOT"

if [[ ! -f "$REPO_ROOT/fixtures/run_demo/runpack.zip" || ! -f "$REPO_ROOT/gait.yaml" ]]; then
  "$BIN_PATH" demo --json >/dev/null
  "$BIN_PATH" regress init --from run_demo --json >"$ARTIFACT_ROOT/regress_init_result.json"
fi

set +e
"$BIN_PATH" regress run --json --junit "$ARTIFACT_ROOT/junit.xml" >"$ARTIFACT_ROOT/regress_result.json"
regress_status=$?
set -e

if [[ "$regress_status" -eq 5 ]]; then
  echo "regress fail (stable exit code 5)"
  exit 5
fi
if [[ "$regress_status" -ne 0 ]]; then
  echo "unexpected regress exit code: $regress_status" >&2
  exit "$regress_status"
fi

"$BIN_PATH" policy test examples/policy/endpoint/allow_safe_endpoints.yaml examples/policy/endpoint/fixtures/intent_allow.json --json >/dev/null
set +e
"$BIN_PATH" policy test examples/policy/endpoint/block_denied_endpoints.yaml examples/policy/endpoint/fixtures/intent_block.json --json >/dev/null
block_status=$?
"$BIN_PATH" policy test examples/policy/endpoint/require_approval_destructive.yaml examples/policy/endpoint/fixtures/intent_destructive.json --json >/dev/null
approval_status=$?
set -e
if [[ "$block_status" -ne 3 ]]; then
  echo "endpoint block fixture exit mismatch: $block_status" >&2
  exit 1
fi
if [[ "$approval_status" -ne 4 ]]; then
  echo "endpoint approval fixture exit mismatch: $approval_status" >&2
  exit 1
fi

echo "ci regress contract: pass"
