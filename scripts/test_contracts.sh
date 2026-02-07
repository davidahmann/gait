#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [[ $# -gt 1 ]]; then
  echo "usage: $0 [path-to-gait-binary]" >&2
  exit 2
fi

if [[ $# -eq 1 ]]; then
  if [[ "$1" = /* ]]; then
    BIN_PATH="$1"
  else
    BIN_PATH="$(pwd)/$1"
  fi
else
  BIN_PATH="$REPO_ROOT/gait"
  go build -o "$BIN_PATH" ./cmd/gait
fi

if [[ ! -x "$BIN_PATH" ]]; then
  echo "binary is not executable: $BIN_PATH" >&2
  exit 2
fi

WORK_A="$(mktemp -d)"
WORK_B="$(mktemp -d)"
trap 'rm -rf "$WORK_A" "$WORK_B"' EXIT

echo "==> deterministic artifact bytes"
(
  cd "$WORK_A"
  "$BIN_PATH" demo >/dev/null
)
(
  cd "$WORK_B"
  "$BIN_PATH" demo >/dev/null
)
SUM_A="$(shasum -a 256 "$WORK_A/gait-out/runpack_run_demo.zip" | awk '{print $1}')"
SUM_B="$(shasum -a 256 "$WORK_B/gait-out/runpack_run_demo.zip" | awk '{print $1}')"
if [[ "$SUM_A" != "$SUM_B" ]]; then
  echo "determinism failure: runpack bytes differ" >&2
  exit 1
fi

echo "==> stable exit-code contract"
cp "$REPO_ROOT/examples/policy-test/intent.json" "$WORK_A/intent.json"
cp "$REPO_ROOT/examples/policy-test/allow.yaml" "$WORK_A/allow.yaml"
cp "$REPO_ROOT/examples/policy-test/block.yaml" "$WORK_A/block.yaml"
cp "$REPO_ROOT/examples/policy-test/require_approval.yaml" "$WORK_A/require_approval.yaml"

(
  cd "$WORK_A"
  "$BIN_PATH" policy test allow.yaml intent.json --json > allow.json
)

set +e
(
  cd "$WORK_A"
  "$BIN_PATH" policy test block.yaml intent.json --json > block.json
)
BLOCK_CODE=$?
(
  cd "$WORK_A"
  "$BIN_PATH" policy test require_approval.yaml intent.json --json > require_approval.json
)
APPROVAL_CODE=$?
set -e

if [[ $BLOCK_CODE -ne 3 ]]; then
  echo "unexpected block exit code: $BLOCK_CODE" >&2
  exit 1
fi
if [[ $APPROVAL_CODE -ne 4 ]]; then
  echo "unexpected require_approval exit code: $APPROVAL_CODE" >&2
  exit 1
fi

echo "==> json shape contract"
(
  cd "$WORK_A"
  "$BIN_PATH" verify run_demo --json > verify.json
  "$BIN_PATH" gate eval \
    --policy "$REPO_ROOT/examples/policy/base_low_risk.yaml" \
    --intent "$REPO_ROOT/examples/policy/intents/intent_read.json" \
    --trace-out "$WORK_A/trace_contract.json" \
    --json > gate.json
)

python3 - "$WORK_A" <<'PY'
import json
import sys
from pathlib import Path

work = Path(sys.argv[1])
verify = json.loads((work / "verify.json").read_text(encoding="utf-8"))
gate = json.loads((work / "gate.json").read_text(encoding="utf-8"))

verify_required = {"ok", "run_id", "manifest_digest", "files_checked", "signature_status"}
gate_required = {"ok", "verdict", "reason_codes", "trace_id", "trace_path", "policy_digest", "intent_digest"}

missing_verify = sorted(verify_required.difference(verify.keys()))
missing_gate = sorted(gate_required.difference(gate.keys()))

if missing_verify:
    raise SystemExit(f"verify shape missing keys: {missing_verify}")
if missing_gate:
    raise SystemExit(f"gate shape missing keys: {missing_gate}")
if verify.get("ok") is not True:
    raise SystemExit("verify expected ok=true")
if gate.get("verdict") != "allow":
    raise SystemExit(f"gate expected allow verdict, got={gate.get('verdict')}")
if not isinstance(gate.get("reason_codes"), list):
    raise SystemExit("gate reason_codes must be an array")
PY

echo "==> compatibility roundtrip from legacy fixture"
cp "$REPO_ROOT/internal/integration/testdata/legacy_run_record_v1.json" "$WORK_A/legacy_run_record_v1.json"
(
  cd "$WORK_A"
  "$BIN_PATH" run record --input legacy_run_record_v1.json --out-dir "$WORK_A/gait-out" --json > legacy_record.json
  "$BIN_PATH" migrate --input "$WORK_A/gait-out/runpack_run_legacy_fixture.zip" --json > legacy_migrate.json
  "$BIN_PATH" verify "$WORK_A/gait-out/runpack_run_legacy_fixture_migrated.zip" --json > legacy_verify.json
)

python3 - "$WORK_A" <<'PY'
import json
import sys
from pathlib import Path

work = Path(sys.argv[1])
record = json.loads((work / "legacy_record.json").read_text(encoding="utf-8"))
migrate = json.loads((work / "legacy_migrate.json").read_text(encoding="utf-8"))
verify = json.loads((work / "legacy_verify.json").read_text(encoding="utf-8"))

if not record.get("ok"):
    raise SystemExit("legacy record failed")
if not migrate.get("ok"):
    raise SystemExit("legacy migrate failed")
if migrate.get("status") not in {"migrated", "up_to_date", "no_change"}:
    raise SystemExit(f"unexpected migrate status: {migrate.get('status')}")
if not verify.get("ok"):
    raise SystemExit("legacy migrated verify failed")
PY

echo "contracts: pass"
