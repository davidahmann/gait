#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GAIT_BIN="${GAIT_BIN:-}"
DEMO_WORKDIR="${DEMO_WORKDIR:-${REPO_ROOT}/gait-out/demo_90s/workspace}"

if [[ -z "${GAIT_BIN}" ]]; then
  if command -v gait >/dev/null 2>&1; then
    GAIT_BIN="$(command -v gait)"
  elif [[ -x "${REPO_ROOT}/gait" ]]; then
    GAIT_BIN="${REPO_ROOT}/gait"
  else
    echo "==> building local gait binary"
    (cd "${REPO_ROOT}" && go build -o ./gait ./cmd/gait)
    GAIT_BIN="${REPO_ROOT}/gait"
  fi
fi

export PATH="$(dirname "${GAIT_BIN}"):${PATH}"
mkdir -p "${DEMO_WORKDIR}"

if [[ ! -e "${DEMO_WORKDIR}/schemas" ]]; then
  ln -s "${REPO_ROOT}/schemas" "${DEMO_WORKDIR}/schemas"
fi

echo "==> demo workspace: ${DEMO_WORKDIR}"
cd "${DEMO_WORKDIR}"

PROOF_BUNDLE_DIR="${DEMO_WORKDIR}/proof_bundle"
mkdir -p "${PROOF_BUNDLE_DIR}"

echo "==> gait demo (Runpack in seconds)"
"${GAIT_BIN}" demo

echo "==> gait verify run_demo (offline integrity proof)"
"${GAIT_BIN}" verify run_demo --json | tee "${PROOF_BUNDLE_DIR}/verify_output.json"

echo "==> gait gate eval (blocked prompt injection + signed trace)"
TRACE_PATH="./trace_block.json"
EVAL_PATH="./gate_eval_block.json"
POLICY_PATH="./policy_prompt_injection.yaml"
INTENT_PATH="./intent_prompt_injection.json"
cp "${REPO_ROOT}/examples/prompt-injection/policy.yaml" "${POLICY_PATH}"
cp "${REPO_ROOT}/examples/prompt-injection/intent_injected.json" "${INTENT_PATH}"
SIGNING_PREFIX="demo_signing"
SIGNING_PRIVATE_KEY="./${SIGNING_PREFIX}_private.key"
if [[ ! -f "${SIGNING_PRIVATE_KEY}" ]]; then
  "${GAIT_BIN}" keys init --out-dir "." --prefix "${SIGNING_PREFIX}" --json >/dev/null
fi
"${GAIT_BIN}" gate eval \
  --policy "${POLICY_PATH}" \
  --intent "${INTENT_PATH}" \
  --trace-out "${TRACE_PATH}" \
  --key-mode prod \
  --private-key "${SIGNING_PRIVATE_KEY}" \
  --json | tee "${EVAL_PATH}"

python3 - "${EVAL_PATH}" "${TRACE_PATH}" <<'PY'
import json
import sys
from pathlib import Path

eval_path = Path(sys.argv[1])
trace_path = Path(sys.argv[2])

eval_payload = json.loads(eval_path.read_text(encoding="utf-8"))
if eval_payload.get("verdict") != "block":
    raise SystemExit("expected verdict=block from gait gate eval")

trace_payload = json.loads(trace_path.read_text(encoding="utf-8"))
if not isinstance(trace_payload.get("signature"), dict):
    raise SystemExit("expected signature object in trace output")

summary = {
    "verdict": eval_payload.get("verdict"),
    "reason_codes": eval_payload.get("reason_codes", []),
    "trace_path": str(trace_path),
    "trace_key_id": trace_payload["signature"].get("key_id", ""),
}
print(json.dumps(summary))
PY

echo "==> gait run inspect (readable artifact narrative)"
"${GAIT_BIN}" run inspect --from run_demo

echo "==> gait regress bootstrap (incident -> CI test)"
"${GAIT_BIN}" regress bootstrap --from run_demo --json --junit "./junit.xml" | tee "${PROOF_BUNDLE_DIR}/regress_result.json"

echo "==> build adoption proof bundle"
if [[ -f "./gait-out/runpack_run_demo.zip" ]]; then
  cp "./gait-out/runpack_run_demo.zip" "${PROOF_BUNDLE_DIR}/runpack_run_demo.zip"
fi
if [[ -f "./junit.xml" ]]; then
  cp "./junit.xml" "${PROOF_BUNDLE_DIR}/junit.xml"
fi

cat > "${PROOF_BUNDLE_DIR}/adoption_metrics.json" <<'JSON'
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

python3 "${REPO_ROOT}/scripts/check_integration_lane_scorecard.py" \
  --input "${PROOF_BUNDLE_DIR}/adoption_metrics.json" \
  --out "${PROOF_BUNDLE_DIR}/integration_lane_scorecard.json"

if bash "${REPO_ROOT}/scripts/test_intent_receipt_conformance.sh" "${GAIT_BIN}" >"${PROOF_BUNDLE_DIR}/intent_receipt_conformance.txt" 2>&1; then
  printf '%s\n' '{"ok":true,"report":"intent_receipt_conformance.txt"}' > "${PROOF_BUNDLE_DIR}/conformance_report.json"
else
  printf '%s\n' '{"ok":false,"report":"intent_receipt_conformance.txt"}' > "${PROOF_BUNDLE_DIR}/conformance_report.json"
  echo "conformance check failed; see ${PROOF_BUNDLE_DIR}/intent_receipt_conformance.txt" >&2
  exit 1
fi

echo "demo_90s: pass"
