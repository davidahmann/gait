#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

go build -o ./gait ./cmd/gait
export PATH="$repo_root:$PATH"

mkdir -p "$repo_root/gait-out"

had_fixtures=0
if [[ -d "$repo_root/fixtures" ]]; then
  had_fixtures=1
fi

created_before="$(mktemp)"
touch "$created_before"

cleanup() {
  find "$repo_root" -maxdepth 1 -type f \
    \( -name 'trace_*.json' -o -name 'approval_audit_*.json' -o -name 'gait.yaml' -o -name 'regress_result.json' \) \
    -newer "$created_before" -delete
  if [[ "$had_fixtures" -eq 0 && -d "$repo_root/fixtures" ]]; then
    rm -rf "$repo_root/fixtures"
  fi
  rm -f "$created_before"
}
trap cleanup EXIT

echo "==> quickstart smoke"
quickstart_output="$(bash scripts/quickstart.sh)"
printf '%s\n' "$quickstart_output"
QUICKSTART_OUTPUT="$quickstart_output" python3 - <<'PY'
import os
from pathlib import Path

raw = os.environ["QUICKSTART_OUTPUT"]
parsed: dict[str, str] = {}
for line in raw.splitlines():
    if "=" not in line:
        continue
    key, value = line.split("=", 1)
    parsed[key.strip()] = value.strip()

required = [
    "quickstart_status",
    "run_id",
    "runpack",
    "verify_json",
    "regress_init_json",
    "regress_run_json",
    "junit",
]
missing = [field for field in required if field not in parsed]
if missing:
    raise SystemExit(f"quickstart missing required output fields: {missing}")
if parsed["quickstart_status"] != "pass":
    raise SystemExit(f"quickstart_status expected pass, got {parsed['quickstart_status']}")
for field in ("runpack", "verify_json", "regress_init_json", "regress_run_json", "junit"):
    if not Path(parsed[field]).exists():
        raise SystemExit(f"quickstart artifact missing: {field}={parsed[field]}")
PY

echo "==> integration adapter smoke"
bash scripts/test_adapter_parity.sh
python3 examples/integrations/template/quickstart.py --scenario allow
python3 examples/integrations/template/quickstart.py --scenario block
python3 examples/integrations/template/quickstart.py --scenario require_approval
python3 examples/integrations/openai_agents/quickstart.py --scenario require_approval
python3 - <<'PY'
from pathlib import Path

required = [
    Path("gait-out/integrations/template/trace_allow.json"),
    Path("gait-out/integrations/template/trace_block.json"),
    Path("gait-out/integrations/template/trace_require_approval.json"),
    Path("gait-out/integrations/openai_agents/trace_require_approval.json"),
]
missing = [str(path) for path in required if not path.exists()]
if missing:
    raise SystemExit(f"missing adapter first-win trace artifacts: {missing}")
PY

echo "==> sidecar smoke"
python3 examples/sidecar/gate_sidecar.py \
  --policy examples/policy-test/allow.yaml \
  --intent-file core/schema/testdata/gate_intent_request_valid.json \
  --trace-out "$repo_root/gait-out/trace_sidecar_allow.json" \
  > "$repo_root/gait-out/sidecar_allow.json"
python3 examples/sidecar/gate_sidecar.py \
  --policy examples/policy-test/block.yaml \
  --intent-file core/schema/testdata/gate_intent_request_valid.json \
  --trace-out "$repo_root/gait-out/trace_sidecar_block.json" \
  > "$repo_root/gait-out/sidecar_block.json"
python3 - <<'PY'
import json
from pathlib import Path

allow = json.loads(Path("gait-out/sidecar_allow.json").read_text(encoding="utf-8"))
block = json.loads(Path("gait-out/sidecar_block.json").read_text(encoding="utf-8"))

if allow.get("gate_result", {}).get("verdict") != "allow":
    raise SystemExit("sidecar allow verdict mismatch")
if block.get("gate_result", {}).get("verdict") != "block":
    raise SystemExit("sidecar block verdict mismatch")
if not allow.get("trace_path"):
    raise SystemExit("sidecar allow trace_path missing")
if not block.get("trace_path"):
    raise SystemExit("sidecar block trace_path missing")
PY

echo "==> scenario pack smoke"
bash examples/scenarios/incident_reproduction.sh
bash examples/scenarios/prompt_injection_block.sh
bash examples/scenarios/approval_flow.sh

echo "adoption smoke: pass"
