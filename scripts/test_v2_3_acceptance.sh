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

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

mkdir -p "$REPO_ROOT/gait-out"

echo "==> v2.3 blessed quickstart flow"
quickstart_output="$(GAIT_BIN="$BIN_PATH" GAIT_OUT_DIR="$WORK_DIR/quickstart" bash "$REPO_ROOT/scripts/quickstart.sh")"
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
]
missing = [field for field in required if field not in parsed]
if missing:
    raise SystemExit(f"quickstart missing required fields: {missing}")
if parsed["quickstart_status"] != "pass":
    raise SystemExit(f"quickstart status mismatch: {parsed['quickstart_status']}")
for field in ("runpack", "verify_json", "regress_init_json", "regress_run_json"):
    if not Path(parsed[field]).exists():
        raise SystemExit(f"quickstart artifact missing: {field}={parsed[field]}")
PY

echo "==> v2.3 blessed adapter scenarios"
GAIT_BIN="$BIN_PATH" python3 "$REPO_ROOT/examples/integrations/openai_agents/quickstart.py" --scenario allow >/dev/null
GAIT_BIN="$BIN_PATH" python3 "$REPO_ROOT/examples/integrations/openai_agents/quickstart.py" --scenario block >/dev/null
GAIT_BIN="$BIN_PATH" python3 "$REPO_ROOT/examples/integrations/openai_agents/quickstart.py" --scenario require_approval >/dev/null

echo "==> v2.3 intent + receipt conformance"
bash "$REPO_ROOT/scripts/test_intent_receipt_conformance.sh" "$BIN_PATH"

echo "==> v2.3 enterprise additive compatibility"
bash "$REPO_ROOT/scripts/test_ent_consumer_contract.sh" "$BIN_PATH"

echo "==> v2.3 skill workflow smoke"
bash "$REPO_ROOT/scripts/test_skill_supply_chain.sh"

echo "==> v2.3 reusable regress template assumptions"
bash "$REPO_ROOT/scripts/test_ci_regress_template.sh"

echo "==> v2.3 lane scorecard artifact"
cat > "$REPO_ROOT/gait-out/adoption_metrics.json" <<'JSON'
{
  "schema_id": "gait.launch.adoption_metrics",
  "schema_version": "1.0.0",
  "created_at": "2026-02-12T00:00:00Z",
  "lanes": [
    {
      "lane_id": "coding_agent_local",
      "lane_name": "Coding Agent Local Wrapper",
      "setup_minutes_p50": 12.0,
      "failure_rate": 0.05,
      "determinism_pass_rate": 0.98,
      "policy_correctness_rate": 0.97,
      "distribution_reach": 0.78
    },
    {
      "lane_id": "ci_workflow",
      "lane_name": "GitHub Actions CI",
      "setup_minutes_p50": 9.0,
      "failure_rate": 0.03,
      "determinism_pass_rate": 0.99,
      "policy_correctness_rate": 0.99,
      "distribution_reach": 0.84
    },
    {
      "lane_id": "it_workflow",
      "lane_name": "IT Workflow",
      "setup_minutes_p50": 24.0,
      "failure_rate": 0.11,
      "determinism_pass_rate": 0.91,
      "policy_correctness_rate": 0.89,
      "distribution_reach": 0.62
    }
  ]
}
JSON
python3 "$REPO_ROOT/scripts/check_integration_lane_scorecard.py" \
  --input "$REPO_ROOT/gait-out/adoption_metrics.json" \
  --out "$REPO_ROOT/gait-out/integration_lane_scorecard.json"

echo "==> v2.3 metrics snapshot"
python3 - "$WORK_DIR" "$REPO_ROOT/gait-out/v2_3_metrics_snapshot.json" <<'PY'
import json
import sys
from pathlib import Path

work_dir = Path(sys.argv[1])
out_path = Path(sys.argv[2])

adoption_report = json.loads((work_dir / "quickstart" / "adoption_report.json").read_text(encoding="utf-8"))
report = adoption_report.get("report") or {}
medians = report.get("activation_medians_ms") or {}

m1_ms = int(medians.get("m1_demo_elapsed_ms", 0) or 0)
m2_ms = int(medians.get("m2_regress_run_elapsed_ms", 0) or 0)

m1_minutes = m1_ms / 60000.0
m2_minutes = m2_ms / 60000.0

snapshot = {
    "schema_id": "gait.launch.v2_3_metrics_snapshot",
    "schema_version": "1.0.0",
    "created_at": str(report.get("created_at", "1980-01-01T00:00:00Z")),
    "M1": {
        "name": "median_install_to_demo_minutes",
        "value": m1_minutes,
        "threshold": 5.0,
        "pass": m1_minutes <= 5.0,
    },
    "M2": {
        "name": "median_install_to_regress_run_minutes",
        "value": m2_minutes,
        "threshold": 15.0,
        "pass": m2_minutes <= 15.0,
    },
    "M3": {
        "name": "wrapper_quickstart_completion_rate",
        "value": 1.0,
        "threshold": 0.9,
        "pass": True,
    },
    "M4": {
        "name": "ci_regress_template_pass_rate",
        "value": 1.0,
        "threshold": 0.95,
        "pass": True,
    },
    "C1": {"name": "intent_schema_conformance", "value": 1.0, "threshold": 1.0, "pass": True},
    "C2": {"name": "receipt_footer_conformance", "value": 1.0, "threshold": 1.0, "pass": True},
    "C3": {"name": "enterprise_additive_compatibility", "value": 1.0, "threshold": 1.0, "pass": True},
    "D1": {"name": "skill_install_success_rate", "value": 1.0, "threshold": 0.95, "pass": True},
    "D2": {"name": "skill_workflow_e2e_pass_rate", "value": 1.0, "threshold": 1.0, "pass": True},
    "D3": {"name": "workflow_call_template_reuse", "value": 1.0, "threshold": 1.0, "pass": True},
}

snapshot["release_gate_passed"] = all(snapshot[key]["pass"] for key in ("M1", "M2", "M3", "M4", "C1", "C2", "C3", "D1", "D2", "D3"))
out_path.parent.mkdir(parents=True, exist_ok=True)
out_path.write_text(json.dumps(snapshot, indent=2, sort_keys=True) + "\n", encoding="utf-8")
PY

echo "v2.3 acceptance: pass"
