#!/usr/bin/env bash
set -euo pipefail

if [[ $# -gt 1 ]]; then
  echo "usage: $0 [path-to-gait-binary]" >&2
  exit 2
fi

GAIT_BIN="${1:-gait}"
if [[ -x "$GAIT_BIN" ]]; then
  GAIT_CMD="$(cd "$(dirname "$GAIT_BIN")" && pwd)/$(basename "$GAIT_BIN")"
elif command -v "$GAIT_BIN" >/dev/null 2>&1; then
  GAIT_CMD="$(command -v "$GAIT_BIN")"
else
  echo "gait binary not found: $GAIT_BIN" >&2
  exit 2
fi

WORK_DIR="${GAIT_SIGNATURE_WORKDIR:-$(mktemp -d)}"
mkdir -p "$WORK_DIR"
cd "$WORK_DIR"

echo "==> signature demo workspace: $WORK_DIR"

echo "==> 1) gait demo"
"$GAIT_CMD" demo --json > demo.json

echo "==> 2) gait verify run_demo"
"$GAIT_CMD" verify run_demo --json > verify.json

echo "==> 3) create changed runpack fixture"
cat > run_record_changed.json <<'JSON'
{
  "run": {
    "schema_id": "gait.runpack.run",
    "schema_version": "1.0.0",
    "created_at": "2026-01-01T00:00:00Z",
    "producer_version": "signature-demo",
    "run_id": "run_demo_changed",
    "env": {
      "os": "linux",
      "arch": "amd64",
      "runtime": "shell"
    },
    "timeline": [
      { "event": "start", "ts": "2026-01-01T00:00:00Z" },
      { "event": "finish", "ts": "2026-01-01T00:00:02Z" }
    ]
  },
  "intents": [
    {
      "schema_id": "gait.runpack.intent",
      "schema_version": "1.0.0",
      "created_at": "2026-01-01T00:00:00Z",
      "producer_version": "signature-demo",
      "run_id": "run_demo_changed",
      "intent_id": "intent_changed_1",
      "tool_name": "tool.write",
      "args_digest": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      "args": {
        "path": "./changed.txt",
        "content": "changed"
      }
    }
  ],
  "results": [
    {
      "schema_id": "gait.runpack.result",
      "schema_version": "1.0.0",
      "created_at": "2026-01-01T00:00:01Z",
      "producer_version": "signature-demo",
      "run_id": "run_demo_changed",
      "intent_id": "intent_changed_1",
      "status": "ok",
      "result_digest": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
      "result": {
        "ok": true,
        "bytes": 7
      }
    }
  ],
  "refs": {
    "schema_id": "gait.runpack.refs",
    "schema_version": "1.0.0",
    "created_at": "2026-01-01T00:00:02Z",
    "producer_version": "signature-demo",
    "run_id": "run_demo_changed",
    "receipts": []
  },
  "capture_mode": "reference"
}
JSON
"$GAIT_CMD" run record --input run_record_changed.json --json > changed.json

echo "==> 4) gait run diff run_demo run_demo_changed"
set +e
"$GAIT_CMD" run diff run_demo run_demo_changed --json > diff.json
diff_status=$?
set -e
if [[ "$diff_status" != "0" && "$diff_status" != "2" ]]; then
  echo "unexpected diff exit code: $diff_status" >&2
  exit "$diff_status"
fi

echo "==> 5) gait run receipt --from run_demo"
"$GAIT_CMD" run receipt --from run_demo --json > receipt.json

runpack_path="$(python3 -c 'import json; print(json.load(open("demo.json", "r", encoding="utf-8")).get("bundle",""))')"
ticket_footer="$(python3 -c 'import json; print(json.load(open("receipt.json", "r", encoding="utf-8")).get("ticket_footer",""))')"

echo
echo "runpack generated: $runpack_path"
echo "verify passed: run_demo"
echo "diff generated: run_demo vs run_demo_changed"
echo "ticket footer:"
echo "$ticket_footer"

if [[ -n "$ticket_footer" ]]; then
  if command -v pbcopy >/dev/null 2>&1; then
    printf '%s' "$ticket_footer" | pbcopy
    echo "copied ticket footer to clipboard (pbcopy)"
  elif command -v xclip >/dev/null 2>&1; then
    printf '%s' "$ticket_footer" | xclip -selection clipboard
    echo "copied ticket footer to clipboard (xclip)"
  fi
fi
