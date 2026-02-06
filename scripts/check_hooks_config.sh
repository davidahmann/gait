#!/usr/bin/env bash
set -euo pipefail

if [[ ! -d ".githooks" ]]; then
  exit 0
fi

if ! command -v git >/dev/null 2>&1; then
  echo "[hooks] git is required to validate core.hooksPath"
  echo "[hooks] remediation: make hooks"
  exit 1
fi

configured="$(git config --get core.hooksPath || true)"
configured="${configured#"${configured%%[![:space:]]*}"}"
configured="${configured%"${configured##*[![:space:]]}"}"

if [[ "${configured}" == ".githooks" ]]; then
  exit 0
fi

echo "[hooks] core.hooksPath is not configured to .githooks"
if [[ -n "${configured}" ]]; then
  echo "[hooks] current value: ${configured}"
else
  echo "[hooks] current value: <unset>"
fi
echo "[hooks] remediation: make hooks"
exit 1
