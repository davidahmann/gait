#!/usr/bin/env bash
set -euo pipefail

if [[ "${GAIT_SKIP_CODEQL:-0}" == "1" ]]; then
  echo "[codeql] skipped (GAIT_SKIP_CODEQL=1)"
  exit 0
fi

if ! command -v codeql >/dev/null 2>&1; then
  echo "[codeql] CodeQL CLI not found in PATH"
  echo "[codeql] install: https://codeql.github.com/docs/codeql-cli/getting-started-with-the-codeql-cli/"
  echo "[codeql] optional bypass: GAIT_SKIP_CODEQL=1"
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "[codeql] python3 is required to parse SARIF output"
  exit 1
fi

langs_raw="${GAIT_CODEQL_LANGS:-go,python}"
fail_level="${GAIT_CODEQL_FAIL_LEVEL:-warning}"
threads="${GAIT_CODEQL_THREADS:-0}"
ram_mb="${GAIT_CODEQL_RAM_MB:-6000}"

IFS=',' read -r -a langs <<<"${langs_raw}"
if [[ "${#langs[@]}" -eq 0 ]]; then
  echo "[codeql] no languages configured"
  exit 2
fi

tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/gait-codeql-XXXXXX")"
cleanup() {
  rm -rf "${tmp_dir}"
}
trap cleanup EXIT

sarif_paths=()
for lang in "${langs[@]}"; do
  lang="$(echo "${lang}" | tr '[:upper:]' '[:lower:]' | xargs)"
  if [[ -z "${lang}" ]]; then
    continue
  fi

  db_path="${tmp_dir}/${lang}-db"
  sarif_path="${tmp_dir}/${lang}.sarif"

  case "${lang}" in
    go)
      echo "[codeql] creating ${lang} database"
      codeql database create "${db_path}" --language=go --source-root . --command "go build ./..."
      echo "[codeql] analyzing ${lang} database"
      codeql database analyze "${db_path}" \
        codeql/go-queries:codeql-suites/go-code-scanning.qls \
        --format=sarif-latest \
        --output "${sarif_path}" \
        --threads="${threads}" \
        --ram="${ram_mb}"
      ;;
    python)
      echo "[codeql] creating ${lang} database"
      codeql database create "${db_path}" --language=python --source-root .
      echo "[codeql] analyzing ${lang} database"
      codeql database analyze "${db_path}" \
        codeql/python-queries:codeql-suites/python-code-scanning.qls \
        --format=sarif-latest \
        --output "${sarif_path}" \
        --threads="${threads}" \
        --ram="${ram_mb}"
      ;;
    *)
      echo "[codeql] unsupported language '${lang}' (supported: go, python)"
      exit 2
      ;;
  esac

  sarif_paths+=("${sarif_path}")
done

if [[ "${#sarif_paths[@]}" -eq 0 ]]; then
  echo "[codeql] no analyses were executed"
  exit 2
fi

python3 - "${fail_level}" "${sarif_paths[@]}" <<'PY'
import json
import sys
from pathlib import Path

if len(sys.argv) < 3:
    print("[codeql] internal error: missing parser args", file=sys.stderr)
    sys.exit(2)

min_level = sys.argv[1].strip().lower()
paths = [Path(arg) for arg in sys.argv[2:]]

level_rank = {"note": 0, "warning": 1, "error": 2}
threshold = level_rank.get(min_level, 1)

findings: list[dict[str, object]] = []

for sarif_path in paths:
    with sarif_path.open("r", encoding="utf-8") as handle:
        data = json.load(handle)

    for run in data.get("runs", []):
        rule_meta: dict[str, dict[str, object]] = {}
        driver = run.get("tool", {}).get("driver", {})
        for rule in driver.get("rules", []):
            rule_id = str(rule.get("id", ""))
            if rule_id:
                rule_meta[rule_id] = rule

        for result in run.get("results", []):
            rule_id = str(result.get("ruleId", "unknown"))
            level = str(result.get("level", "")).lower().strip()
            if not level:
                level = str(
                    rule_meta.get(rule_id, {})
                    .get("defaultConfiguration", {})
                    .get("level", "warning")
                ).lower().strip() or "warning"

            if level_rank.get(level, 1) < threshold:
                continue

            message = str(result.get("message", {}).get("text", ""))
            location = (
                result.get("locations", [{}])[0]
                .get("physicalLocation", {})
            )
            uri = str(location.get("artifactLocation", {}).get("uri", ""))
            line = location.get("region", {}).get("startLine")

            findings.append(
                {
                    "sarif": sarif_path.name,
                    "rule": rule_id,
                    "level": level,
                    "uri": uri,
                    "line": line,
                    "message": message,
                }
            )

if findings:
    print(f"[codeql] findings >= {min_level}: {len(findings)}")
    for item in findings:
        line = "?"
        if isinstance(item["line"], int):
            line = str(item["line"])
        print(
            f"  - [{item['level']}] {item['rule']} at {item['uri']}:{line}"
        )
    sys.exit(1)

print(f"[codeql] no findings >= {min_level}")
PY
