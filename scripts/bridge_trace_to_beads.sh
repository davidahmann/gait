#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
usage: bridge_trace_to_beads.sh --trace <trace.json> [--dry-run|--live] [--bd-bin <path>] [--type <task|bug|...>] [--priority <0-4>] [--json] [--description-out <path>]

Converts a signed Gait trace record into a deterministic Beads issue payload.

Examples:
  bash scripts/bridge_trace_to_beads.sh --trace gait-out/trace_block.json --dry-run --json
  bash scripts/bridge_trace_to_beads.sh --trace gait-out/trace_block.json --live --type bug --priority 1
EOF
}

TRACE_PATH=""
BD_BIN="${BD_BIN:-bd}"
ISSUE_TYPE="task"
PRIORITY="2"
MODE="dry-run"
JSON_OUTPUT=0
DESCRIPTION_OUT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --trace)
      TRACE_PATH="${2:-}"
      shift 2
      ;;
    --bd-bin)
      BD_BIN="${2:-}"
      shift 2
      ;;
    --type)
      ISSUE_TYPE="${2:-}"
      shift 2
      ;;
    --priority)
      PRIORITY="${2:-}"
      shift 2
      ;;
    --dry-run)
      MODE="dry-run"
      shift
      ;;
    --live)
      MODE="live"
      shift
      ;;
    --description-out)
      DESCRIPTION_OUT="${2:-}"
      shift 2
      ;;
    --json)
      JSON_OUTPUT=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$TRACE_PATH" ]]; then
  echo "--trace is required" >&2
  usage >&2
  exit 2
fi
if [[ ! -f "$TRACE_PATH" ]]; then
  echo "trace file not found: $TRACE_PATH" >&2
  exit 2
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT
summary_json="$tmp_dir/summary.json"

python3 - "$TRACE_PATH" "$summary_json" <<'PY'
from __future__ import annotations

import json
import pathlib
import sys

trace_path = pathlib.Path(sys.argv[1]).expanduser().resolve()
summary_path = pathlib.Path(sys.argv[2]).expanduser().resolve()
payload = json.loads(trace_path.read_text(encoding="utf-8"))

trace_id = str(payload.get("trace_id", "")).strip()
verdict = str(payload.get("verdict", "")).strip()
tool_name = str(payload.get("tool_name", "")).strip()
intent_digest = str(payload.get("intent_digest", "")).strip()
policy_digest = str(payload.get("policy_digest", "")).strip()

if not trace_id:
    raise SystemExit("trace_id is required in trace payload")
if not verdict:
    raise SystemExit("verdict is required in trace payload")
if not tool_name:
    raise SystemExit("tool_name is required in trace payload")
if not intent_digest:
    raise SystemExit("intent_digest is required in trace payload")
if not policy_digest:
    raise SystemExit("policy_digest is required in trace payload")

violations_raw = payload.get("violations", [])
if not isinstance(violations_raw, list):
    raise SystemExit("violations must be an array")
violations = sorted({str(item).strip() for item in violations_raw if str(item).strip()})

reason_codes_raw = payload.get("reason_codes", [])
if isinstance(reason_codes_raw, list):
    reason_codes = sorted({str(item).strip() for item in reason_codes_raw if str(item).strip()})
else:
    reason_codes = []

title = f"Gait {verdict}: {tool_name} ({trace_id})"
description_lines = [
    "Gait trace bridge report",
    "",
    f"Trace Path: {trace_path}",
    f"Trace ID: {trace_id}",
    f"Verdict: {verdict}",
    f"Tool Name: {tool_name}",
    f"Policy Digest: {policy_digest}",
    f"Intent Digest: {intent_digest}",
    f"Violations: {', '.join(violations) if violations else 'none'}",
    f"Reason Codes: {', '.join(reason_codes) if reason_codes else 'none'}",
    "",
    "Follow-up:",
    "- if blocked unexpectedly, validate policy match and reason codes",
    "- if expected block, convert incident to regression fixture",
]

summary = {
    "trace_path": str(trace_path),
    "trace_id": trace_id,
    "verdict": verdict,
    "tool_name": tool_name,
    "policy_digest": policy_digest,
    "intent_digest": intent_digest,
    "violations": violations,
    "reason_codes": reason_codes,
    "title": title,
    "description": "\n".join(description_lines).strip() + "\n",
}
summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
PY

if [[ -n "$DESCRIPTION_OUT" ]]; then
  python3 - "$summary_json" "$DESCRIPTION_OUT" <<'PY'
from __future__ import annotations

import json
import pathlib
import sys

summary = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
description = str(summary["description"])
output_path = pathlib.Path(sys.argv[2]).expanduser().resolve()
output_path.parent.mkdir(parents=True, exist_ok=True)
output_path.write_text(description, encoding="utf-8")
PY
fi

if [[ "$MODE" == "live" ]] && ! command -v "$BD_BIN" >/dev/null 2>&1; then
  echo "bd command not available: $BD_BIN" >&2
  exit 2
fi

if [[ "$MODE" == "live" ]]; then
  create_json="$tmp_dir/create.json"
  title="$(python3 - "$summary_json" <<'PY'
from __future__ import annotations

import json
import pathlib
import sys

summary = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
print(summary["title"])
PY
)"
  description="$(python3 - "$summary_json" <<'PY'
from __future__ import annotations

import json
import pathlib
import sys

summary = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
print(summary["description"], end="")
PY
)"

  "$BD_BIN" create "$title" --type "$ISSUE_TYPE" --priority "$PRIORITY" --description "$description" --json >"$create_json"

  python3 - "$summary_json" "$create_json" "$JSON_OUTPUT" <<'PY'
from __future__ import annotations

import json
import pathlib
import sys

summary = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
create_payload = json.loads(pathlib.Path(sys.argv[2]).read_text(encoding="utf-8"))
json_output = int(sys.argv[3]) == 1
issue_id = ""
if isinstance(create_payload, dict):
    issue_id = str(create_payload.get("id", "")).strip()
result = {
    "ok": True,
    "mode": "live",
    "issue_id": issue_id,
    **summary,
}
if json_output:
    print(json.dumps(result, indent=2))
else:
    print(f"mode=live")
    print(f"issue_id={issue_id}")
    print(f"trace_id={summary['trace_id']}")
    print(f"verdict={summary['verdict']}")
    print(f"title={summary['title']}")
PY
  exit 0
fi

python3 - "$summary_json" "$JSON_OUTPUT" <<'PY'
from __future__ import annotations

import json
import pathlib
import sys

summary = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
json_output = int(sys.argv[2]) == 1
result = {
    "ok": True,
    "mode": "dry_run",
    **summary,
}
if json_output:
    print(json.dumps(result, indent=2))
else:
    print("mode=dry_run")
    print(f"trace_id={summary['trace_id']}")
    print(f"verdict={summary['verdict']}")
    print(f"title={summary['title']}")
PY
