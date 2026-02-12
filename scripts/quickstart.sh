#!/usr/bin/env bash
set -euo pipefail

OUTPUT_DIR="${GAIT_OUT_DIR:-./gait-out/quickstart}"
GAIT_BIN="${GAIT_BIN:-}"
SKIP_REGRESS_RUN="${GAIT_QUICKSTART_SKIP_REGRESS_RUN:-0}"

resolve_gait_bin() {
  if [[ -n "${GAIT_BIN}" ]]; then
    return
  fi
  if command -v gait >/dev/null 2>&1; then
    GAIT_BIN="$(command -v gait)"
    return
  fi
  if [[ -x "./gait" ]]; then
    GAIT_BIN="$(pwd)/gait"
    return
  fi
  if [[ -f "./cmd/gait/main.go" ]] && command -v go >/dev/null 2>&1; then
    echo "==> building local gait binary"
    go build -o ./gait ./cmd/gait
    GAIT_BIN="$(pwd)/gait"
    return
  fi
  echo "error: gait binary not found" >&2
  echo "action: set GAIT_BIN, add gait to PATH, or run from repo root with Go installed" >&2
  exit 1
}

checkpoint() {
  local name="$1"
  local started="$2"
  local now
  now="$(python3 - <<'PY'
import time
print(int(time.time() * 1000))
PY
)"
  local elapsed_ms=$(( now - started ))
  echo "checkpoint=${name} elapsed_ms=${elapsed_ms}"
}

resolve_gait_bin

if [[ "${OUTPUT_DIR}" != /* ]]; then
  OUTPUT_DIR="$(pwd)/${OUTPUT_DIR#./}"
fi

if [[ "${GAIT_BIN}" != /* ]]; then
  GAIT_BIN="$(cd "$(dirname "${GAIT_BIN}")" && pwd)/$(basename "${GAIT_BIN}")"
fi

mkdir -p "${OUTPUT_DIR}"
if ! : > "${OUTPUT_DIR}/.writecheck"; then
  echo "error: output directory is not writable: ${OUTPUT_DIR}" >&2
  exit 1
fi
rm -f "${OUTPUT_DIR}/.writecheck"

WORKSPACE="${OUTPUT_DIR}/workspace"
mkdir -p "${WORKSPACE}"

ADOPTION_LOG_PATH="${GAIT_ADOPTION_LOG:-${OUTPUT_DIR}/adoption.jsonl}"
export GAIT_ADOPTION_LOG="${ADOPTION_LOG_PATH}"

DEMO_JSON="${OUTPUT_DIR}/demo.json"
VERIFY_JSON="${OUTPUT_DIR}/verify.json"
REGRESS_INIT_JSON="${OUTPUT_DIR}/regress_init.json"
REGRESS_RUN_JSON="${OUTPUT_DIR}/regress_run.json"
ADOPTION_REPORT_JSON="${OUTPUT_DIR}/adoption_report.json"
JUNIT_PATH="${OUTPUT_DIR}/junit.xml"

start_all="$(python3 - <<'PY'
import time
print(int(time.time() * 1000))
PY
)"

step_start="${start_all}"
echo "==> gait demo"
(
  cd "${WORKSPACE}"
  "${GAIT_BIN}" demo --json > "${DEMO_JSON}"
)
RUN_ID="$(python3 - "${DEMO_JSON}" <<'PY'
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
run_id = str(payload.get("run_id", "")).strip()
if not run_id:
    raise SystemExit("missing run_id in demo output")
print(run_id)
PY
)"
checkpoint "demo" "${step_start}"

step_start="$(python3 - <<'PY'
import time
print(int(time.time() * 1000))
PY
)"
echo "==> gait verify ${RUN_ID}"
(
  cd "${WORKSPACE}"
  "${GAIT_BIN}" verify "${RUN_ID}" --json > "${VERIFY_JSON}"
)
checkpoint "verify" "${step_start}"

step_start="$(python3 - <<'PY'
import time
print(int(time.time() * 1000))
PY
)"
echo "==> gait regress init --from ${RUN_ID}"
(
  cd "${WORKSPACE}"
  "${GAIT_BIN}" regress init --from "${RUN_ID}" --json > "${REGRESS_INIT_JSON}"
)
checkpoint "regress_init" "${step_start}"

if [[ "${SKIP_REGRESS_RUN}" == "1" ]]; then
  echo "==> skip regress run (GAIT_QUICKSTART_SKIP_REGRESS_RUN=1)"
else
  step_start="$(python3 - <<'PY'
import time
print(int(time.time() * 1000))
PY
)"
  echo "==> gait regress run"
  (
    cd "${WORKSPACE}"
    "${GAIT_BIN}" regress run --json --junit "${JUNIT_PATH}" > "${REGRESS_RUN_JSON}"
  )
  checkpoint "regress_run" "${step_start}"
fi

if [[ -f "${ADOPTION_LOG_PATH}" ]]; then
  "${GAIT_BIN}" doctor adoption --from "${ADOPTION_LOG_PATH}" --json > "${ADOPTION_REPORT_JSON}"
fi

end_all="$(python3 - <<'PY'
import time
print(int(time.time() * 1000))
PY
)"
total_ms=$(( end_all - start_all ))
RUNPACK_PATH="${WORKSPACE}/gait-out/runpack_${RUN_ID}.zip"

if [[ ! -f "${RUNPACK_PATH}" ]]; then
  echo "error: expected runpack not found at ${RUNPACK_PATH}" >&2
  echo "action: rerun quickstart and inspect ${DEMO_JSON}" >&2
  exit 1
fi

echo "quickstart_status=pass"
echo "run_id=${RUN_ID}"
echo "runpack=${RUNPACK_PATH}"
echo "verify_json=${VERIFY_JSON}"
echo "regress_init_json=${REGRESS_INIT_JSON}"
if [[ -f "${REGRESS_RUN_JSON}" ]]; then
  echo "regress_run_json=${REGRESS_RUN_JSON}"
fi
if [[ -f "${JUNIT_PATH}" ]]; then
  echo "junit=${JUNIT_PATH}"
fi
if [[ -f "${ADOPTION_REPORT_JSON}" ]]; then
  echo "adoption_report_json=${ADOPTION_REPORT_JSON}"
fi
echo "elapsed_ms=${total_ms}"
