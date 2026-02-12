#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

index_path="docs/ecosystem/community_index.json"
output_path="gait-out/ecosystem_release_notes.md"
metrics_path="gait-out/v2_3_metrics_snapshot.json"

python3 scripts/validate_community_index.py "$index_path"
mkdir -p "$(dirname "$output_path")"
cat > "$metrics_path" <<'JSON'
{
  "schema_id": "gait.launch.v2_3_metrics_snapshot",
  "schema_version": "1.0.0",
  "release_gate_passed": true,
  "M1": {"name": "median_install_to_demo_minutes", "value": 1.0, "threshold": 5.0, "pass": true}
}
JSON

python3 scripts/render_ecosystem_release_notes.py "$index_path" "$output_path" "$metrics_path"

if [[ ! -s "$output_path" ]]; then
  echo "ecosystem release notes output missing or empty: $output_path" >&2
  exit 1
fi
if ! grep -Eq '^# Ecosystem Release Notes$' "$output_path"; then
  echo "ecosystem release notes heading missing in $output_path" >&2
  exit 1
fi
if ! grep -q 'adapter-openai-agents-official' "$output_path"; then
  echo "expected adapter entry missing from ecosystem release notes" >&2
  exit 1
fi
if ! grep -q '## v2.3 Metrics Snapshot' "$output_path"; then
  echo "expected v2.3 metrics snapshot section missing from ecosystem release notes" >&2
  exit 1
fi

echo "ecosystem release automation: pass"
