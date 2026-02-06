#!/usr/bin/env bash
set -euo pipefail

OUTPUT_DIR="${GAIT_OUT_DIR:-./gait-out}"

if ! command -v gait >/dev/null 2>&1; then
  echo "error: gait binary not found on PATH"
  echo "install gait first, then rerun this script"
  exit 1
fi

if ! mkdir -p "${OUTPUT_DIR}"; then
  echo "error: output directory is not writable: ${OUTPUT_DIR}"
  exit 1
fi

WRITE_CHECK_PATH="${OUTPUT_DIR}/.gait-quickstart-writecheck"
if ! : >"${WRITE_CHECK_PATH}"; then
  echo "error: output directory is not writable: ${OUTPUT_DIR}"
  exit 1
fi
rm -f "${WRITE_CHECK_PATH}"

echo "==> Running gait demo"
DEMO_OUTPUT="$(gait demo)"
printf '%s\n' "${DEMO_OUTPUT}"

RUN_ID="$(printf '%s\n' "${DEMO_OUTPUT}" | awk -F'=' '/^run_id=/{print $2; exit}')"
if [ -z "${RUN_ID}" ]; then
  echo "error: unable to parse run_id from gait demo output"
  exit 1
fi

echo "==> Running gait verify ${RUN_ID}"
gait verify "${RUN_ID}"

echo "next: gait regress init --from ${RUN_ID}"
