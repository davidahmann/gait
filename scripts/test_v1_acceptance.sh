#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <path-to-gait-binary>" >&2
  exit 2
fi

if [[ "$1" = /* ]]; then
  BIN_PATH="$1"
else
  BIN_PATH="$(pwd)/$1"
fi
if [[ ! -x "$BIN_PATH" ]]; then
  echo "binary is not executable: $BIN_PATH" >&2
  exit 2
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

cp "$REPO_ROOT/examples/policy-test/intent.json" "$WORK_DIR/intent.json"
cp "$REPO_ROOT/examples/policy-test/allow.yaml" "$WORK_DIR/allow.yaml"
cp "$REPO_ROOT/examples/policy-test/block.yaml" "$WORK_DIR/block.yaml"
cp "$REPO_ROOT/examples/policy-test/require_approval.yaml" "$WORK_DIR/require_approval.yaml"

cd "$WORK_DIR"

DEMO_OUT="$("$BIN_PATH" demo)"
echo "$DEMO_OUT"
[[ "$DEMO_OUT" == *"run_id=run_demo"* ]]
[[ "$DEMO_OUT" == *"ticket_footer=GAIT run_id=run_demo"* ]]
[[ "$DEMO_OUT" == *"verify=ok"* ]]
python3 - "$DEMO_OUT" <<'PY'
import re
import sys

output = sys.argv[1]
ticket_footer = ""
for line in output.splitlines():
    if line.startswith("ticket_footer="):
        ticket_footer = line.removeprefix("ticket_footer=").strip()
        break
if not ticket_footer:
    raise SystemExit("ticket_footer line missing")

pattern = re.compile(
    r'^GAIT run_id=([A-Za-z0-9_-]+) manifest=sha256:([a-f0-9]{64}) verify="gait verify ([A-Za-z0-9_-]+)"$'
)
match = pattern.match(ticket_footer)
if match is None:
    raise SystemExit(f"ticket_footer format mismatch: {ticket_footer}")
if match.group(1) != match.group(3):
    raise SystemExit("ticket_footer run_id mismatch")
PY

DEMO_JSON="$("$BIN_PATH" demo --json)"
python3 - "$DEMO_JSON" <<'PY'
import json
import sys

payload = json.loads(sys.argv[1])
if not payload.get("ok"):
    raise SystemExit("demo --json returned ok=false")
if payload.get("run_id") != "run_demo":
    raise SystemExit(f"unexpected demo run_id: {payload.get('run_id')}")
if payload.get("verify") != "ok":
    raise SystemExit(f"unexpected demo verify status: {payload.get('verify')}")
PY

VERIFY_OUT="$("$BIN_PATH" verify run_demo)"
echo "$VERIFY_OUT"
[[ "$VERIFY_OUT" == *"verify ok"* ]]

RECEIPT_OUT="$("$BIN_PATH" run receipt --from run_demo)"
python3 - "$RECEIPT_OUT" <<'PY'
import re
import sys

ticket_footer = sys.argv[1].strip()
pattern = re.compile(
    r'^GAIT run_id=([A-Za-z0-9_-]+) manifest=sha256:([a-f0-9]{64}) verify="gait verify ([A-Za-z0-9_-]+)"$'
)
match = pattern.match(ticket_footer)
if match is None:
    raise SystemExit(f"run receipt output mismatch: {ticket_footer}")
if match.group(1) != match.group(3):
    raise SystemExit("run receipt run_id mismatch")
PY

REPLAY_A="$("$BIN_PATH" run replay --json run_demo)"
REPLAY_B="$("$BIN_PATH" run replay --json run_demo)"
python3 - "$REPLAY_A" "$REPLAY_B" <<'PY'
import json
import sys

first = json.loads(sys.argv[1])
second = json.loads(sys.argv[2])
if first != second:
    raise SystemExit("stub replay is not deterministic")
if first.get("mode") != "stub":
    raise SystemExit("expected replay mode stub")
PY

"$BIN_PATH" regress init --from run_demo --json > regress_init.json
python3 - <<'PY'
import json
from pathlib import Path

payload = json.loads(Path("regress_init.json").read_text(encoding="utf-8"))
if not payload.get("ok"):
    raise SystemExit("regress init returned ok=false")
if payload.get("run_id") != "run_demo":
    raise SystemExit("unexpected run_id from regress init")
PY

"$BIN_PATH" regress run --json > regress_run.json
python3 - <<'PY'
import json
from pathlib import Path

payload = json.loads(Path("regress_run.json").read_text(encoding="utf-8"))
if not payload.get("ok"):
    raise SystemExit("regress run returned ok=false")
if payload.get("status") != "pass":
    raise SystemExit(f"unexpected regress status: {payload.get('status')}")
PY

"$BIN_PATH" policy test allow.yaml intent.json --json > allow.json
ALLOW_CODE=$?
if [[ $ALLOW_CODE -ne 0 ]]; then
  echo "unexpected allow exit code: $ALLOW_CODE" >&2
  exit 1
fi

"$BIN_PATH" policy init baseline-highrisk --out generated_policy.yaml --json > policy_init.json
python3 - <<'PY'
import json
from pathlib import Path

init_payload = json.loads(Path("policy_init.json").read_text(encoding="utf-8"))

if not init_payload.get("ok"):
    raise SystemExit("policy init returned ok=false")
if init_payload.get("template") != "baseline-highrisk":
    raise SystemExit(f"unexpected template: {init_payload.get('template')}")
if init_payload.get("policy_path") != "generated_policy.yaml":
    raise SystemExit(f"unexpected policy path: {init_payload.get('policy_path')}")
PY

if "$BIN_PATH" policy validate --help >/dev/null 2>&1 && "$BIN_PATH" policy fmt --help >/dev/null 2>&1; then
  "$BIN_PATH" policy validate generated_policy.yaml --json > policy_validate.json
  "$BIN_PATH" policy fmt generated_policy.yaml --write --json > policy_fmt_first.json
  "$BIN_PATH" policy fmt generated_policy.yaml --write --json > policy_fmt_second.json
  python3 - <<'PY'
import json
from pathlib import Path

validate_payload = json.loads(Path("policy_validate.json").read_text(encoding="utf-8"))
fmt_first = json.loads(Path("policy_fmt_first.json").read_text(encoding="utf-8"))
fmt_second = json.loads(Path("policy_fmt_second.json").read_text(encoding="utf-8"))

if not validate_payload.get("ok"):
    raise SystemExit("policy validate returned ok=false")
if validate_payload.get("default_verdict") != "block":
    raise SystemExit(f"unexpected validated default verdict: {validate_payload.get('default_verdict')}")
if not fmt_first.get("ok"):
    raise SystemExit("first policy fmt returned ok=false")
if not fmt_second.get("ok"):
    raise SystemExit("second policy fmt returned ok=false")
if fmt_second.get("changed", False) is not False:
    raise SystemExit("policy fmt second run should be idempotent")
PY

  if "$BIN_PATH" policy simulate --help >/dev/null 2>&1; then
    "$BIN_PATH" policy simulate --baseline allow.yaml --policy block.yaml --fixtures intent.json --json > policy_simulate.json
    python3 - <<'PY'
import json
from pathlib import Path

simulate_payload = json.loads(Path("policy_simulate.json").read_text(encoding="utf-8"))
if not simulate_payload.get("ok"):
    raise SystemExit("policy simulate returned ok=false")
if simulate_payload.get("fixtures_total") != 1:
    raise SystemExit(f"unexpected fixtures_total: {simulate_payload.get('fixtures_total')}")
if simulate_payload.get("changed_fixtures", 0) < 1:
    raise SystemExit("expected changed_fixtures >= 1")
if simulate_payload.get("recommendation") != "require_approval":
    raise SystemExit(f"unexpected recommendation: {simulate_payload.get('recommendation')}")
PY
  fi
else
  echo "skip policy validate/fmt checks (binary does not support these commands yet)"
fi

if "$BIN_PATH" keys --help >/dev/null 2>&1; then
  "$BIN_PATH" keys init --out-dir "$WORK_DIR/keys" --prefix acceptance --json > keys_init.json
  python3 - <<'PY'
import json
from pathlib import Path

payload = json.loads(Path("keys_init.json").read_text(encoding="utf-8"))
if not payload.get("ok"):
    raise SystemExit("keys init returned ok=false")
private_path = Path(payload.get("private_key_path", ""))
public_path = Path(payload.get("public_key_path", ""))
if not private_path.exists():
    raise SystemExit(f"missing private key path: {private_path}")
if not public_path.exists():
    raise SystemExit(f"missing public key path: {public_path}")
PY

  private_key_path="$(python3 - <<'PY'
import json
from pathlib import Path
payload = json.loads(Path("keys_init.json").read_text(encoding="utf-8"))
print(payload["private_key_path"])
PY
)"
  public_key_path="$(python3 - <<'PY'
import json
from pathlib import Path
payload = json.loads(Path("keys_init.json").read_text(encoding="utf-8"))
print(payload["public_key_path"])
PY
)"

  "$BIN_PATH" keys verify --private-key "$private_key_path" --public-key "$public_key_path" --json > keys_verify.json
  python3 - <<'PY'
import json
from pathlib import Path

payload = json.loads(Path("keys_verify.json").read_text(encoding="utf-8"))
if not payload.get("ok"):
    raise SystemExit("keys verify returned ok=false")
if payload.get("key_id", "") == "":
    raise SystemExit("keys verify missing key_id")
PY
fi

set +e
"$BIN_PATH" policy test block.yaml intent.json --json > block.json
BLOCK_CODE=$?
"$BIN_PATH" policy test require_approval.yaml intent.json --json > require_approval.json
APPROVAL_CODE=$?
set -e

if [[ $BLOCK_CODE -ne 3 ]]; then
  echo "unexpected block exit code: $BLOCK_CODE" >&2
  exit 1
fi
if [[ $APPROVAL_CODE -ne 4 ]]; then
  echo "unexpected require_approval exit code: $APPROVAL_CODE" >&2
  exit 1
fi

python3 - <<'PY'
import json
from pathlib import Path

allow = json.loads(Path("allow.json").read_text(encoding="utf-8"))
block = json.loads(Path("block.json").read_text(encoding="utf-8"))
approval = json.loads(Path("require_approval.json").read_text(encoding="utf-8"))

if allow.get("verdict") != "allow":
    raise SystemExit("allow policy verdict mismatch")
if block.get("verdict") != "block":
    raise SystemExit("block policy verdict mismatch")
if approval.get("verdict") != "require_approval":
    raise SystemExit("require_approval policy verdict mismatch")
PY

echo "v1 acceptance checks passed"
