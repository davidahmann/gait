#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

if [[ $# -gt 1 ]]; then
  echo "usage: $0 [path-to-gait-binary]" >&2
  exit 2
fi

if [[ $# -eq 1 ]]; then
  if [[ "$1" = /* ]]; then
    BIN_PATH="$1"
  else
    BIN_PATH="$(pwd)/$1"
  fi
else
  BIN_PATH="$REPO_ROOT/gait"
  go build -o "$BIN_PATH" ./cmd/gait
fi

if [[ ! -x "$BIN_PATH" ]]; then
  echo "binary is not executable: $BIN_PATH" >&2
  exit 2
fi

WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/gait-voice-acceptance-XXXXXX")"
trap 'rm -rf "${WORK_DIR}"' EXIT

cd "$WORK_DIR"

echo "==> voice: seed runpack"
"$BIN_PATH" demo --json > "$WORK_DIR/demo.json"
RUNPACK_PATH="$(python3 - <<'PY' "$WORK_DIR/demo.json"
import json
import os
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is True, payload
assert payload.get("run_id") == "run_demo", payload
bundle = payload.get("bundle")
assert isinstance(bundle, str) and bundle, payload
resolved = bundle if os.path.isabs(bundle) else os.path.join(os.getcwd(), bundle)
assert os.path.exists(resolved), resolved
print(resolved)
PY
)"

cat > "$WORK_DIR/policy_allow.yaml" <<'EOF'
schema_id: gait.gate.policy
schema_version: 1.0.0
default_verdict: allow
EOF

cat > "$WORK_DIR/policy_block.yaml" <<'EOF'
schema_id: gait.gate.policy
schema_version: 1.0.0
default_verdict: block
EOF

cat > "$WORK_DIR/policy_require_approval.yaml" <<'EOF'
schema_id: gait.gate.policy
schema_version: 1.0.0
default_verdict: require_approval
EOF

python3 - <<'PY' "$WORK_DIR/commitment_intent.json"
import json
import sys
from pathlib import Path

intent = {
    "schema_id": "gait.voice.commitment_intent",
    "schema_version": "1.0.0",
    "created_at": "2026-02-15T00:00:00Z",
    "producer_version": "acceptance",
    "call_id": "call_voice_acceptance",
    "turn_index": 2,
    "call_seq": 2,
    "commitment_class": "quote",
    "currency": "USD",
    "quote_min_cents": 1200,
    "quote_max_cents": 1800,
    "context": {
        "identity": "agent.voice",
        "workspace": "/srv/voice",
        "risk_class": "high",
        "session_id": "sess_voice_acceptance",
        "request_id": "req_voice_acceptance",
        "environment_fingerprint": "voice:acceptance",
    },
}
Path(sys.argv[1]).write_text(json.dumps(intent, indent=2) + "\n", encoding="utf-8")
PY

echo "==> voice: key material"
"$BIN_PATH" keys init --out-dir "$WORK_DIR/keys" --prefix voice --json > "$WORK_DIR/keys.json"
PRIVATE_KEY_PATH="$(python3 - <<'PY' "$WORK_DIR/keys.json"
import json
import os
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is True, payload
private_key = payload.get("private_key_path")
assert private_key and os.path.exists(private_key), payload
print(private_key)
PY
)"

echo "==> voice: allow token mint"
"$BIN_PATH" voice token mint \
  --intent "$WORK_DIR/commitment_intent.json" \
  --policy "$WORK_DIR/policy_allow.yaml" \
  --trace-out "$WORK_DIR/trace_allow.json" \
  --out "$WORK_DIR/say_token_allow.json" \
  --key-mode prod \
  --private-key "$PRIVATE_KEY_PATH" \
  --json > "$WORK_DIR/token_allow.json"

read -r TOKEN_PATH TOKEN_ID INTENT_DIGEST POLICY_DIGEST <<<"$(python3 - <<'PY' "$WORK_DIR/token_allow.json"
import json
import os
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is True, payload
assert payload.get("verdict") == "allow", payload
token_path = payload.get("token_path")
token_id = payload.get("token_id")
intent_digest = payload.get("intent_digest")
policy_digest = payload.get("policy_digest")
assert token_path and os.path.exists(token_path), payload
assert token_id, payload
assert intent_digest and len(intent_digest) == 64, payload
assert policy_digest and len(policy_digest) == 64, payload
print(token_path, token_id, intent_digest, policy_digest)
PY
)"

echo "==> voice: token verify and mismatch guard"
"$BIN_PATH" voice token verify \
  --token "$TOKEN_PATH" \
  --private-key "$PRIVATE_KEY_PATH" \
  --intent-digest "$INTENT_DIGEST" \
  --policy-digest "$POLICY_DIGEST" \
  --call-id "call_voice_acceptance" \
  --turn-index 2 \
  --call-seq 2 \
  --commitment-class quote \
  --json > "$WORK_DIR/token_verify_ok.json"
python3 - <<'PY' "$WORK_DIR/token_verify_ok.json"
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is True, payload
assert payload.get("token_id"), payload
PY

set +e
"$BIN_PATH" voice token verify \
  --token "$TOKEN_PATH" \
  --private-key "$PRIVATE_KEY_PATH" \
  --policy-digest "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" \
  --json > "$WORK_DIR/token_verify_fail.json"
VERIFY_FAIL_CODE=$?
set -e
if [[ "$VERIFY_FAIL_CODE" -eq 0 ]]; then
  echo "expected token verify mismatch to fail" >&2
  exit 1
fi
python3 - <<'PY' "$WORK_DIR/token_verify_fail.json"
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is False, payload
assert payload.get("error_code"), payload
PY

echo "==> voice: block and approval gate outcomes"
set +e
"$BIN_PATH" voice token mint \
  --intent "$WORK_DIR/commitment_intent.json" \
  --policy "$WORK_DIR/policy_block.yaml" \
  --trace-out "$WORK_DIR/trace_block.json" \
  --out "$WORK_DIR/say_token_block.json" \
  --key-mode prod \
  --private-key "$PRIVATE_KEY_PATH" \
  --json > "$WORK_DIR/token_block.json"
BLOCK_CODE=$?
set -e
if [[ "$BLOCK_CODE" -eq 0 ]]; then
  echo "expected block policy to prevent token mint" >&2
  exit 1
fi
python3 - <<'PY' "$WORK_DIR/token_block.json"
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is False, payload
assert payload.get("verdict") == "block", payload
assert not payload.get("token_path"), payload
PY

set +e
"$BIN_PATH" voice token mint \
  --intent "$WORK_DIR/commitment_intent.json" \
  --policy "$WORK_DIR/policy_require_approval.yaml" \
  --trace-out "$WORK_DIR/trace_require_approval.json" \
  --out "$WORK_DIR/say_token_require_approval.json" \
  --key-mode prod \
  --private-key "$PRIVATE_KEY_PATH" \
  --json > "$WORK_DIR/token_require_approval.json"
APPROVAL_CODE=$?
set -e
if [[ "$APPROVAL_CODE" -eq 0 ]]; then
  echo "expected require_approval policy to prevent token mint" >&2
  exit 1
fi
python3 - <<'PY' "$WORK_DIR/token_require_approval.json"
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is False, payload
assert payload.get("verdict") == "require_approval", payload
assert not payload.get("token_path"), payload
PY

python3 - <<'PY' "$WORK_DIR/call_record_allow.json" "$RUNPACK_PATH" "$TOKEN_ID" "$INTENT_DIGEST" "$POLICY_DIGEST"
import json
import sys
from pathlib import Path

target = Path(sys.argv[1])
runpack_path = sys.argv[2]
token_id = sys.argv[3]
intent_digest = sys.argv[4]
policy_digest = sys.argv[5]

record = {
    "schema_id": "gait.voice.call_record",
    "schema_version": "1.0.0",
    "created_at": "2026-02-15T00:00:00Z",
    "producer_version": "acceptance",
    "call_id": "call_voice_acceptance",
    "runpack_path": runpack_path,
    "privacy_mode": "hash_only",
    "environment_fingerprint": "voice:acceptance",
    "events": [
        {"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": "2026-02-15T00:00:00Z", "call_id": "call_voice_acceptance", "call_seq": 1, "turn_index": 1, "event_type": "asr.final", "payload_digest": "1" * 64},
        {"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": "2026-02-15T00:00:01Z", "call_id": "call_voice_acceptance", "call_seq": 2, "turn_index": 2, "event_type": "commitment.declared", "commitment_class": "quote"},
        {"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": "2026-02-15T00:00:02Z", "call_id": "call_voice_acceptance", "call_seq": 3, "turn_index": 2, "event_type": "gate.decision", "commitment_class": "quote", "intent_digest": intent_digest, "policy_digest": policy_digest},
        {"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": "2026-02-15T00:00:03Z", "call_id": "call_voice_acceptance", "call_seq": 4, "turn_index": 2, "event_type": "tts.request", "commitment_class": "quote"},
        {"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": "2026-02-15T00:00:04Z", "call_id": "call_voice_acceptance", "call_seq": 5, "turn_index": 2, "event_type": "tts.emitted", "commitment_class": "quote", "say_token_id": token_id},
        {"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": "2026-02-15T00:00:05Z", "call_id": "call_voice_acceptance", "call_seq": 6, "turn_index": 2, "event_type": "tool.intent"},
        {"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": "2026-02-15T00:00:06Z", "call_id": "call_voice_acceptance", "call_seq": 7, "turn_index": 2, "event_type": "tool.result"},
        {"schema_id": "gait.voice.call_event", "schema_version": "1.0.0", "created_at": "2026-02-15T00:00:07Z", "call_id": "call_voice_acceptance", "call_seq": 8, "turn_index": 2, "event_type": "approval.granted", "commitment_class": "quote"},
    ],
    "commitments": [
        {
            "schema_id": "gait.voice.commitment_intent",
            "schema_version": "1.0.0",
            "created_at": "2026-02-15T00:00:01Z",
            "producer_version": "acceptance",
            "call_id": "call_voice_acceptance",
            "turn_index": 2,
            "call_seq": 2,
            "commitment_class": "quote",
            "currency": "USD",
            "quote_min_cents": 1200,
            "quote_max_cents": 1800,
            "context": {
                "identity": "agent.voice",
                "workspace": "/srv/voice",
                "risk_class": "high",
                "session_id": "sess_voice_acceptance",
                "request_id": "req_voice_acceptance",
                "environment_fingerprint": "voice:acceptance",
            },
        }
    ],
    "gate_decisions": [
        {
            "call_id": "call_voice_acceptance",
            "call_seq": 3,
            "turn_index": 2,
            "commitment_class": "quote",
            "verdict": "allow",
            "reason_codes": ["allow_default"],
            "intent_digest": intent_digest,
            "policy_digest": policy_digest,
        }
    ],
    "speak_receipts": [
        {
            "call_id": "call_voice_acceptance",
            "call_seq": 5,
            "turn_index": 2,
            "commitment_class": "quote",
            "say_token_id": token_id,
            "spoken_digest": "2" * 64,
            "emitted_at": "2026-02-15T00:00:04Z",
        }
    ],
    "reference_digests": [{"ref_id": "kb.quote", "sha256": "3" * 64}],
}
target.write_text(json.dumps(record, indent=2) + "\n", encoding="utf-8")
PY

echo "==> voice: callpack build/verify/inspect/diff"
"$BIN_PATH" voice pack build --from "$WORK_DIR/call_record_allow.json" --out "$WORK_DIR/callpack_allow.zip" --json > "$WORK_DIR/callpack_build.json"
CALLPACK_PATH="$(python3 - <<'PY' "$WORK_DIR/callpack_build.json"
import json
import os
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is True, payload
assert payload.get("pack_type") == "call", payload
path = payload.get("path")
assert path and os.path.exists(path), payload
print(path)
PY
)"

"$BIN_PATH" voice pack verify "$CALLPACK_PATH" --json > "$WORK_DIR/callpack_verify.json"
python3 - <<'PY' "$WORK_DIR/callpack_verify.json"
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is True, payload
verify = payload.get("verify") or {}
assert verify.get("pack_type") == "call", payload
assert not verify.get("missing_files"), payload
assert not verify.get("hash_mismatches"), payload
PY

"$BIN_PATH" voice pack inspect "$CALLPACK_PATH" --json > "$WORK_DIR/callpack_inspect.json"
python3 - <<'PY' "$WORK_DIR/callpack_inspect.json"
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is True, payload
inspect = payload.get("inspect") or {}
call_payload = inspect.get("call_payload") or {}
assert call_payload.get("privacy_mode") == "hash_only", payload
assert call_payload.get("call_id") == "call_voice_acceptance", payload
PY

"$BIN_PATH" voice pack diff "$CALLPACK_PATH" "$CALLPACK_PATH" --json > "$WORK_DIR/callpack_diff.json"
python3 - <<'PY' "$WORK_DIR/callpack_diff.json"
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is True, payload
diff = payload.get("diff") or {}
result = diff.get("result") or diff.get("Result") or {}
summary = result.get("summary") or {}
assert summary.get("changed") is False, payload
PY

echo "==> voice: regress bootstrap from callpack"
"$BIN_PATH" regress bootstrap --from "$CALLPACK_PATH" --name voice_acceptance --json > "$WORK_DIR/regress_bootstrap.json"
python3 - <<'PY' "$WORK_DIR/regress_bootstrap.json"
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is True, payload
assert payload.get("status") == "pass", payload
fixture_dir = payload.get("fixture_dir")
runpack_path = payload.get("runpack_path")
assert fixture_dir and runpack_path, payload
assert Path(fixture_dir).exists(), payload
assert Path(runpack_path).exists(), payload
PY

echo "==> voice: tamper detection"
python3 - <<'PY' "$CALLPACK_PATH" "$WORK_DIR/callpack_tampered.zip"
import sys
from pathlib import Path

source = Path(sys.argv[1]).read_bytes()
mutated = bytearray(source)
mutated[len(mutated) // 2] ^= 0x01
Path(sys.argv[2]).write_bytes(bytes(mutated))
PY
set +e
"$BIN_PATH" voice pack verify "$WORK_DIR/callpack_tampered.zip" --json > "$WORK_DIR/callpack_tampered_verify.json"
TAMPERED_CODE=$?
set -e
if [[ "$TAMPERED_CODE" -eq 0 ]]; then
  echo "expected tampered callpack verification to fail" >&2
  exit 1
fi

echo "==> voice: non-bypassable speak boundary"
python3 - <<'PY' "$WORK_DIR/call_record_allow.json" "$WORK_DIR/call_record_missing_token.json"
import json
import sys
from pathlib import Path

record = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
record["call_id"] = "call_voice_missing_token"
for key in ("events", "commitments", "gate_decisions", "speak_receipts"):
    for row in record[key]:
        row["call_id"] = "call_voice_missing_token"
for row in record["speak_receipts"]:
    row["say_token_id"] = ""
Path(sys.argv[2]).write_text(json.dumps(record, indent=2) + "\n", encoding="utf-8")
PY
set +e
"$BIN_PATH" voice pack build --from "$WORK_DIR/call_record_missing_token.json" --json > "$WORK_DIR/callpack_missing_token.json"
MISSING_TOKEN_CODE=$?
set -e
if [[ "$MISSING_TOKEN_CODE" -eq 0 ]]; then
  echo "expected callpack build to fail with missing say_token_id" >&2
  exit 1
fi

python3 - <<'PY' "$WORK_DIR/call_record_allow.json" "$WORK_DIR/call_record_missing_receipt.json"
import json
import sys
from pathlib import Path

record = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
record["call_id"] = "call_voice_missing_receipt"
for key in ("events", "commitments", "gate_decisions", "speak_receipts"):
    for row in record[key]:
        row["call_id"] = "call_voice_missing_receipt"
record["speak_receipts"] = []
Path(sys.argv[2]).write_text(json.dumps(record, indent=2) + "\n", encoding="utf-8")
PY
set +e
"$BIN_PATH" voice pack build --from "$WORK_DIR/call_record_missing_receipt.json" --json > "$WORK_DIR/callpack_missing_receipt.json"
MISSING_RECEIPT_CODE=$?
set -e
if [[ "$MISSING_RECEIPT_CODE" -eq 0 ]]; then
  echo "expected callpack build to fail with missing speak receipt for gated tts.emitted" >&2
  exit 1
fi

echo "==> voice: privacy mode dispute_encrypted"
python3 - <<'PY' "$WORK_DIR/call_record_allow.json" "$WORK_DIR/call_record_dispute.json"
import json
import sys
from pathlib import Path

record = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
record["call_id"] = "call_voice_dispute"
record["privacy_mode"] = "dispute_encrypted"
for key in ("events", "commitments", "gate_decisions", "speak_receipts"):
    for row in record[key]:
        row["call_id"] = "call_voice_dispute"
Path(sys.argv[2]).write_text(json.dumps(record, indent=2) + "\n", encoding="utf-8")
PY
"$BIN_PATH" voice pack build --from "$WORK_DIR/call_record_dispute.json" --out "$WORK_DIR/callpack_dispute.zip" --json > "$WORK_DIR/callpack_dispute_build.json"
"$BIN_PATH" voice pack inspect "$WORK_DIR/callpack_dispute.zip" --json > "$WORK_DIR/callpack_dispute_inspect.json"
python3 - <<'PY' "$WORK_DIR/callpack_dispute_inspect.json"
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
assert payload.get("ok") is True, payload
inspect = payload.get("inspect") or {}
call_payload = inspect.get("call_payload") or {}
assert call_payload.get("privacy_mode") == "dispute_encrypted", payload
PY

echo "==> voice: deterministic changed diff"
set +e
"$BIN_PATH" voice pack diff "$CALLPACK_PATH" "$WORK_DIR/callpack_dispute.zip" --json > "$WORK_DIR/callpack_changed_diff_a.json"
DIFF_A_CODE=$?
"$BIN_PATH" voice pack diff "$CALLPACK_PATH" "$WORK_DIR/callpack_dispute.zip" --json > "$WORK_DIR/callpack_changed_diff_b.json"
DIFF_B_CODE=$?
set -e
if [[ "$DIFF_A_CODE" -eq 0 || "$DIFF_B_CODE" -eq 0 ]]; then
  echo "expected changed diff to return non-zero exit" >&2
  exit 1
fi
python3 - <<'PY' "$WORK_DIR/callpack_changed_diff_a.json" "$WORK_DIR/callpack_changed_diff_b.json"
import json
import sys
from pathlib import Path

a = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
b = json.loads(Path(sys.argv[2]).read_text(encoding="utf-8"))
assert a == b, "changed diff output must be stable across runs"
diff = a.get("diff") or {}
result = diff.get("result") or diff.get("Result") or {}
summary = result.get("summary") or {}
assert summary.get("changed") is True, a
PY

echo "voice acceptance: pass"
