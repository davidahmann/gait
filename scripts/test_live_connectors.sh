#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

go build -o ./gait ./cmd/gait
export GAIT_BIN="$REPO_ROOT/gait"

if [[ "${GAIT_ENABLE_LIVE_CONNECTOR_TESTS:-0}" != "1" ]]; then
  echo "live connectors: skipped (set GAIT_ENABLE_LIVE_CONNECTOR_TESTS=1 to enable)"
  exit 0
fi

echo "==> adapter contract smoke"
python3 examples/integrations/openai_agents/quickstart.py --scenario allow
python3 examples/integrations/openai_agents/quickstart.py --scenario block
python3 examples/integrations/langchain/quickstart.py --scenario allow
python3 examples/integrations/langchain/quickstart.py --scenario block
python3 examples/integrations/autogen/quickstart.py --scenario allow
python3 examples/integrations/autogen/quickstart.py --scenario block

echo "==> provider reachability probes"
if [[ -n "${OPENAI_API_KEY:-}" ]]; then
  HTTP_CODE="$(curl -sS -o /tmp/gait_openai_models.json -w '%{http_code}' \
    https://api.openai.com/v1/models \
    -H "Authorization: Bearer ${OPENAI_API_KEY}")"
  if [[ "$HTTP_CODE" != "200" ]]; then
    echo "openai probe failed: HTTP $HTTP_CODE" >&2
    exit 1
  fi
  python3 - <<'PY'
import json
from pathlib import Path

payload = json.loads(Path("/tmp/gait_openai_models.json").read_text(encoding="utf-8"))
if "data" not in payload:
    raise SystemExit("openai models payload missing data")
PY
else
  echo "openai probe: skipped (OPENAI_API_KEY not set)"
fi

if [[ -n "${ANTHROPIC_API_KEY:-}" ]]; then
  HTTP_CODE="$(curl -sS -o /tmp/gait_anthropic_models.json -w '%{http_code}' \
    https://api.anthropic.com/v1/models \
    -H "x-api-key: ${ANTHROPIC_API_KEY}" \
    -H "anthropic-version: 2023-06-01")"
  if [[ "$HTTP_CODE" != "200" ]]; then
    echo "anthropic probe failed: HTTP $HTTP_CODE" >&2
    exit 1
  fi
  python3 - <<'PY'
import json
from pathlib import Path

payload = json.loads(Path("/tmp/gait_anthropic_models.json").read_text(encoding="utf-8"))
if "data" not in payload:
    raise SystemExit("anthropic models payload missing data")
PY
else
  echo "anthropic probe: skipped (ANTHROPIC_API_KEY not set)"
fi

echo "live connectors: pass"
