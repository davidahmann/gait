#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
package_root="$repo_root/examples/integrations/openclaw/skill"

target_dir="${OPENCLAW_HOME:-$HOME/.openclaw}/skills/gait-gate"
policy_path="$repo_root/examples/policy/base_high_risk.yaml"
force=0
json_output=0

usage() {
  cat <<'EOF'
Usage:
  install_openclaw_skill.sh [--target-dir <path>] [--policy <path>] [--force] [--json]

Description:
  Installs the official Gait OpenClaw skill package into a local OpenClaw skills directory.

Options:
  --target-dir  install destination (default: ${OPENCLAW_HOME:-$HOME/.openclaw}/skills/gait-gate)
  --policy      default policy path to write into skill_config.json
  --force       overwrite existing target directory
  --json        emit JSON output
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --target-dir)
      target_dir="${2:?missing value for --target-dir}"
      shift 2
      ;;
    --policy)
      policy_path="${2:?missing value for --policy}"
      shift 2
      ;;
    --force)
      force=1
      shift
      ;;
    --json)
      json_output=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "install_openclaw_skill: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ ! -d "$package_root" ]]; then
  echo "install_openclaw_skill: package root not found: $package_root" >&2
  exit 2
fi

if [[ ! -f "$policy_path" ]]; then
  echo "install_openclaw_skill: policy path not found: $policy_path" >&2
  exit 2
fi

if [[ -e "$target_dir" && "$force" -ne 1 ]]; then
  echo "install_openclaw_skill: target already exists (use --force): $target_dir" >&2
  exit 2
fi

rm -rf "$target_dir"
mkdir -p "$target_dir"
cp "$package_root"/gait_openclaw_gate.py "$target_dir"/
cp "$package_root"/skill_manifest.json "$target_dir"/
cp "$package_root"/README.md "$target_dir"/

chmod 0755 "$target_dir/gait_openclaw_gate.py"

cat > "$target_dir/skill_config.json" <<EOF
{
  "policy_path": "${policy_path}",
  "adapter": "mcp",
  "key_mode": "dev",
  "identity": "openclaw-user",
  "risk_class": "high"
}
EOF

if [[ "$json_output" -eq 1 ]]; then
  cat <<EOF
{"ok":true,"package":"gait-gate","framework":"openclaw","target_dir":"${target_dir}","entrypoint":"${target_dir}/gait_openclaw_gate.py","policy_path":"${policy_path}"}
EOF
else
  echo "openclaw skill installed"
  echo "package=gait-gate"
  echo "framework=openclaw"
  echo "target_dir=$target_dir"
  echo "entrypoint=$target_dir/gait_openclaw_gate.py"
  echo "policy_path=$policy_path"
fi
