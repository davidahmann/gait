#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

tmp_root="$(mktemp -d)"
trap 'rm -rf "$tmp_root"' EXIT

target_dir="$tmp_root/openclaw/skills/gait-gate"
policy_path="$repo_root/examples/policy/base_high_risk.yaml"

echo "==> openclaw skill install smoke"
output="$(bash scripts/install_openclaw_skill.sh --target-dir "$target_dir" --policy "$policy_path" --json)"
printf '%s\n' "$output"

TARGET_DIR="$target_dir" POLICY_PATH="$policy_path" INSTALL_OUTPUT="$output" python3 - <<'PY'
import json
import os
from pathlib import Path

payload = json.loads(os.environ["INSTALL_OUTPUT"])
target_dir = Path(os.environ["TARGET_DIR"])
policy_path = Path(os.environ["POLICY_PATH"])

if not payload.get("ok"):
    raise SystemExit("install output missing ok=true")
if Path(payload.get("target_dir", "")) != target_dir:
    raise SystemExit("target_dir mismatch")
if Path(payload.get("policy_path", "")) != policy_path:
    raise SystemExit("policy_path mismatch")

required_files = [
    target_dir / "gait_openclaw_gate.py",
    target_dir / "skill_manifest.json",
    target_dir / "skill_config.json",
    target_dir / "README.md",
]
for file_path in required_files:
    if not file_path.exists():
        raise SystemExit(f"missing installed file: {file_path}")

entrypoint = target_dir / "gait_openclaw_gate.py"
if not os.access(entrypoint, os.X_OK):
    raise SystemExit("entrypoint is not executable")

config = json.loads((target_dir / "skill_config.json").read_text(encoding="utf-8"))
if config.get("policy_path") != str(policy_path):
    raise SystemExit("skill_config policy_path mismatch")
if config.get("key_mode") != "dev":
    raise SystemExit("skill_config key_mode mismatch")

manifest = json.loads((target_dir / "skill_manifest.json").read_text(encoding="utf-8"))
if manifest.get("name") != "gait-gate":
    raise SystemExit("manifest name mismatch")
if manifest.get("framework") != "openclaw":
    raise SystemExit("manifest framework mismatch")
PY

echo "==> installer idempotent overwrite with --force"
bash scripts/install_openclaw_skill.sh --target-dir "$target_dir" --policy "$policy_path" --force --json > /dev/null

echo "openclaw skill install checks: pass"
