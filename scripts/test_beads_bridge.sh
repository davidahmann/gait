#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

trace_fixture="$repo_root/scripts/testdata/trace_block_violation.json"
if [[ ! -f "$trace_fixture" ]]; then
  echo "missing fixture: $trace_fixture" >&2
  exit 1
fi

work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT

echo "==> beads bridge dry-run"
bash scripts/bridge_trace_to_beads.sh \
  --trace "$trace_fixture" \
  --dry-run \
  --description-out "$work_dir/description.md" \
  --json >"$work_dir/dry_run.json"

python3 - "$work_dir/dry_run.json" "$work_dir/description.md" <<'PY'
from __future__ import annotations

import json
import pathlib
import sys

payload = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
description_path = pathlib.Path(sys.argv[2])

if payload.get("mode") != "dry_run":
    raise SystemExit(f"expected dry_run mode, got {payload.get('mode')}")
if payload.get("verdict") != "block":
    raise SystemExit(f"expected block verdict, got {payload.get('verdict')}")
if payload.get("trace_id") != "trace_blocked_001":
    raise SystemExit(f"trace_id mismatch: {payload.get('trace_id')}")
if "destructive_action" not in payload.get("violations", []):
    raise SystemExit("expected destructive_action violation")
if not str(payload.get("title", "")).startswith("Gait block:"):
    raise SystemExit(f"unexpected title: {payload.get('title')}")
if not description_path.exists():
    raise SystemExit("expected description-out file")
if "Policy Digest:" not in description_path.read_text(encoding="utf-8"):
    raise SystemExit("description-out missing policy digest line")
PY

echo "==> beads bridge live simulation"
fake_bd="$work_dir/fake-bd.sh"
cat >"$fake_bd" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ "$1" != "create" ]]; then
  echo "unexpected command: $*" >&2
  exit 2
fi
if [[ -z "${FAKE_BD_ARGS_OUT:-}" ]]; then
  echo "FAKE_BD_ARGS_OUT not set" >&2
  exit 2
fi
printf '%s\n' "$*" >"$FAKE_BD_ARGS_OUT"
cat <<'JSON'
{"id":"bd-1234"}
JSON
EOF
chmod +x "$fake_bd"

FAKE_BD_ARGS_OUT="$work_dir/fake_bd_args.txt" \
bash scripts/bridge_trace_to_beads.sh \
  --trace "$trace_fixture" \
  --live \
  --bd-bin "$fake_bd" \
  --type bug \
  --priority 1 \
  --json >"$work_dir/live.json"

python3 - "$work_dir/live.json" "$work_dir/fake_bd_args.txt" <<'PY'
from __future__ import annotations

import json
import pathlib
import sys

payload = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
captured_args = pathlib.Path(sys.argv[2]).read_text(encoding="utf-8")

if payload.get("mode") != "live":
    raise SystemExit(f"expected live mode, got {payload.get('mode')}")
if payload.get("issue_id") != "bd-1234":
    raise SystemExit(f"expected issue_id bd-1234, got {payload.get('issue_id')}")
if "--type bug" not in captured_args:
    raise SystemExit(f"expected --type bug in command args: {captured_args}")
if "--priority 1" not in captured_args:
    raise SystemExit(f"expected --priority 1 in command args: {captured_args}")
PY

echo "beads bridge: pass"
