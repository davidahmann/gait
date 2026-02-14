#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

go build -o ./gait ./cmd/gait

required_files=(
  "scripts/ci_regress_contract.sh"
  "examples/ci/portability/README.md"
  "examples/ci/portability/gitlab/.gitlab-ci.yml"
  "examples/ci/portability/jenkins/Jenkinsfile"
  "examples/ci/portability/circleci/config.yml"
)

for path in "${required_files[@]}"; do
  if [[ ! -f "$path" ]]; then
    echo "missing portability template artifact: $path" >&2
    exit 1
  fi
done

python3 - <<'PY'
from pathlib import Path

checks = {
    "examples/ci/portability/gitlab/.gitlab-ci.yml": [
        "scripts/ci_regress_contract.sh",
        "gait-out/adoption_regress/",
    ],
    "examples/ci/portability/jenkins/Jenkinsfile": [
        "scripts/ci_regress_contract.sh",
        "gait-out/adoption_regress",
    ],
    "examples/ci/portability/circleci/config.yml": [
        "scripts/ci_regress_contract.sh",
        "gait-out/adoption_regress",
    ],
}

for path, fragments in checks.items():
    text = Path(path).read_text(encoding="utf-8")
    missing = [fragment for fragment in fragments if fragment not in text]
    if missing:
        raise SystemExit(f"{path} missing required fragments: {missing}")
PY

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

ARTIFACT_ROOT_PATH="$WORK_DIR/adoption_regress"
ARTIFACT_ROOT="$ARTIFACT_ROOT_PATH" GAIT_BIN="$REPO_ROOT/gait" bash "$REPO_ROOT/scripts/ci_regress_contract.sh"

if [[ ! -f "$ARTIFACT_ROOT_PATH/regress_result.json" ]]; then
  echo "missing regress_result.json from contract script" >&2
  exit 1
fi
if [[ ! -f "$ARTIFACT_ROOT_PATH/junit.xml" ]]; then
  echo "missing junit.xml from contract script" >&2
  exit 1
fi

python3 - "$ARTIFACT_ROOT_PATH/regress_result.json" <<'PY'
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
if payload.get("status") != "pass":
    raise SystemExit(f"expected regress status=pass, got {payload.get('status')}")
if payload.get("ok") is not True:
    raise SystemExit("expected regress ok=true")
PY

echo "ci portability templates: pass"
