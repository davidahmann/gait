#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_ROOT="${REPO_ROOT}/.agents/skills"
CODEX_DEST_ROOT="${CODEX_HOME:-${HOME}/.codex}/skills"
CLAUDE_DEST_ROOT="${CLAUDE_HOME:-${HOME}/.claude}/skills"
PROVIDER="both"

if [[ ! -d "${SOURCE_ROOT}" ]]; then
  echo "skills source not found: ${SOURCE_ROOT}" >&2
  exit 1
fi

usage() {
  cat <<'EOF'
Usage: install_repo_skills.sh [--provider codex|claude|both]

Installs repo skills from .agents/skills into provider skill directories:
  codex  -> ${CODEX_HOME:-$HOME/.codex}/skills
  claude -> ${CLAUDE_HOME:-$HOME/.claude}/skills
  both   -> both destinations (default)
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --provider)
      if [[ $# -lt 2 ]]; then
        echo "--provider requires a value" >&2
        usage
        exit 1
      fi
      PROVIDER="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

case "${PROVIDER}" in
  codex|claude|both) ;;
  *)
    echo "invalid --provider value: ${PROVIDER}" >&2
    usage
    exit 1
    ;;
esac

install_one_provider() {
  local provider="$1"
  local dest_root="$2"
  local installed=0

  mkdir -p "${dest_root}"
  for skill_dir in "${SOURCE_ROOT}"/*; do
    [[ -d "${skill_dir}" ]] || continue
    local skill_name
    skill_name="$(basename "${skill_dir}")"
    rm -rf "${dest_root:?}/${skill_name}"
    cp -R "${skill_dir}" "${dest_root}/${skill_name}"
    installed=$((installed + 1))
    echo "[${provider}] installed ${skill_name} -> ${dest_root}/${skill_name}"
  done

  if [[ "${installed}" -eq 0 ]]; then
    return 1
  fi

  echo "[${provider}] installed ${installed} skills into ${dest_root}"
  return 0
}

installed_any=0
if [[ "${PROVIDER}" == "codex" || "${PROVIDER}" == "both" ]]; then
  install_one_provider "codex" "${CODEX_DEST_ROOT}" && installed_any=1
fi

if [[ "${PROVIDER}" == "claude" || "${PROVIDER}" == "both" ]]; then
  install_one_provider "claude" "${CLAUDE_DEST_ROOT}" && installed_any=1
fi

if [[ "${installed_any}" -eq 0 ]]; then
  echo "no skills installed from ${SOURCE_ROOT}" >&2
  exit 1
fi
