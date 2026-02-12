#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

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

echo "==> intent schema required-field conformance"
python3 - "$REPO_ROOT" <<'PY'
import json
import sys
from pathlib import Path

repo = Path(sys.argv[1])
schema = json.loads((repo / "schemas/v1/gate/intent_request.schema.json").read_text(encoding="utf-8"))
required = schema.get("required", [])
expected = [
    "schema_id",
    "schema_version",
    "created_at",
    "producer_version",
    "tool_name",
    "args",
    "targets",
    "context",
]
if sorted(required) != sorted(expected):
    raise SystemExit(f"intent_request required fields drifted: {required}")
if schema.get("properties", {}).get("schema_id", {}).get("const") != "gait.gate.intent_request":
    raise SystemExit("intent_request schema_id const mismatch")
if schema.get("properties", {}).get("schema_version", {}).get("pattern") != r"^1\.0\.0$":
    raise SystemExit("intent_request schema_version pattern mismatch")
PY

echo "==> gate digest continuity"
(
  cd "$WORK_A"
  "$BIN_PATH" demo --json > demo.json
  "$BIN_PATH" gate eval \
    --policy "$REPO_ROOT/examples/policy/base_low_risk.yaml" \
    --intent "$REPO_ROOT/examples/policy/intents/intent_read.json" \
    --trace-out "$WORK_A/trace_intent_receipt.json" \
    --json > gate.json
  "$BIN_PATH" run receipt --from run_demo --json > receipt.json
)

echo "==> runpack refs + receipt footer continuity"
python3 - "$WORK_A" "$WORK_B" "$BIN_PATH" <<'PY'
import json
import re
import subprocess
import sys
import zipfile
from pathlib import Path

work_a = Path(sys.argv[1])
work_b = Path(sys.argv[2])
bin_path = Path(sys.argv[3])

sha256_re = re.compile(r"^[a-f0-9]{64}$")

with (work_a / "gate.json").open("r", encoding="utf-8") as fh:
    gate = json.load(fh)
required_gate = {"ok", "verdict", "trace_id", "trace_path", "intent_digest", "policy_digest"}
missing_gate = sorted(required_gate.difference(gate.keys()))
if missing_gate:
    raise SystemExit(f"gate output missing keys: {missing_gate}")
if gate.get("verdict") != "allow":
    raise SystemExit(f"expected allow verdict, got {gate.get('verdict')}")
if not sha256_re.match(str(gate.get("intent_digest", "")).strip()):
    raise SystemExit("intent_digest is not 64-hex sha256")
if not sha256_re.match(str(gate.get("policy_digest", "")).strip()):
    raise SystemExit("policy_digest is not 64-hex sha256")

with (work_a / "receipt.json").open("r", encoding="utf-8") as fh:
    receipt = json.load(fh)
if not receipt.get("ok"):
    raise SystemExit("run receipt expected ok=true")
run_id = str(receipt.get("run_id", "")).strip()
manifest_digest = str(receipt.get("manifest_digest", "")).strip()
footer = str(receipt.get("ticket_footer", "")).strip()
if not run_id:
    raise SystemExit("run receipt missing run_id")
if not sha256_re.match(manifest_digest):
    raise SystemExit("run receipt manifest_digest is not 64-hex sha256")
pattern = re.compile(r'^GAIT run_id=([A-Za-z0-9_-]+) manifest=sha256:([a-f0-9]{64}) verify="gait verify ([A-Za-z0-9_-]+)"$')
match = pattern.match(footer)
if not match:
    raise SystemExit(f"ticket_footer format mismatch: {footer}")
if match.group(1) != run_id or match.group(3) != run_id:
    raise SystemExit("ticket_footer run_id continuity mismatch")
if match.group(2) != manifest_digest:
    raise SystemExit("ticket_footer manifest continuity mismatch")

zip_path_a = work_a / "gait-out" / "runpack_run_demo.zip"
if not zip_path_a.exists():
    raise SystemExit(f"missing runpack: {zip_path_a}")

with zipfile.ZipFile(zip_path_a, "r") as archive:
    refs_a = json.loads(archive.read("refs.json").decode("utf-8"))

if refs_a.get("schema_id") != "gait.runpack.refs":
    raise SystemExit("refs schema_id mismatch")
if refs_a.get("run_id") != run_id:
    raise SystemExit("refs run_id mismatch")
receipts = refs_a.get("receipts")
if not isinstance(receipts, list) or not receipts:
    raise SystemExit("refs receipts must be a non-empty list")
required_receipt_fields = {
    "ref_id",
    "source_type",
    "source_locator",
    "query_digest",
    "content_digest",
    "retrieved_at",
    "redaction_mode",
}
for index, row in enumerate(receipts):
    if not isinstance(row, dict):
        raise SystemExit(f"refs receipt[{index}] must be object")
    missing = sorted(required_receipt_fields.difference(row.keys()))
    if missing:
        raise SystemExit(f"refs receipt[{index}] missing fields: {missing}")
    if not sha256_re.match(str(row.get("query_digest", "")).strip()):
        raise SystemExit(f"refs receipt[{index}] query_digest invalid")
    if not sha256_re.match(str(row.get("content_digest", "")).strip()):
        raise SystemExit(f"refs receipt[{index}] content_digest invalid")

# Determinism check: run demo in separate workspace and ensure refs/ticket footer are byte-identical.
subprocess.run([str(bin_path), "demo", "--json"], cwd=work_b, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
subprocess.run([str(bin_path), "run", "receipt", "--from", "run_demo", "--json"], cwd=work_b, check=True, stdout=(work_b / "receipt.json").open("w", encoding="utf-8"), stderr=subprocess.PIPE)
with zipfile.ZipFile(work_b / "gait-out" / "runpack_run_demo.zip", "r") as archive:
    refs_b = json.loads(archive.read("refs.json").decode("utf-8"))

if refs_a != refs_b:
    raise SystemExit("refs.json structure is not deterministic across equivalent runs")
receipt_b = json.loads((work_b / "receipt.json").read_text(encoding="utf-8"))
if receipt_b.get("ticket_footer") != footer:
    raise SystemExit("ticket_footer is not deterministic across equivalent runs")
PY

echo "intent + receipt conformance: pass"
