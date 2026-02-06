#!/usr/bin/env bash
set -euo pipefail

mkdir -p gait-out
summary_path="gait-out/policy_compliance_summary.md"
results_path="gait-out/policy_compliance_results.jsonl"

cat >"$summary_path" <<'EOF'
# Policy Compliance Summary

| Case | Expected Exit | Actual Exit | Verdict | Reason Codes | Result |
| --- | --- | --- | --- | --- | --- |
EOF

: >"$results_path"

run_case() {
  local case_name="$1"
  local policy_path="$2"
  local intent_path="$3"
  local expected_exit="$4"

  local output
  local actual_exit
  set +e
  output="$(./gait policy test "$policy_path" "$intent_path" --json 2>&1)"
  actual_exit=$?
  set -e

  printf '%s\n' "$output" >>"$results_path"

  local verdict
  verdict="$(printf '%s' "$output" | sed -n 's/.*"verdict":"\([^"]*\)".*/\1/p')"
  if [[ -z "$verdict" ]]; then
    verdict="unknown"
  fi

  local reason_codes
  reason_codes="$(printf '%s' "$output" | sed -n 's/.*"reason_codes":\[\([^]]*\)\].*/\1/p')"
  if [[ -z "$reason_codes" ]]; then
    reason_codes="none"
  fi

  local result
  if [[ "$actual_exit" -eq "$expected_exit" ]]; then
    result="pass"
  else
    result="fail"
    failed_cases=1
  fi

  printf '| %s | `%s` | `%s` | `%s` | `%s` | **%s** |\n' \
    "$case_name" "$expected_exit" "$actual_exit" "$verdict" "$reason_codes" "$result" >>"$summary_path"
}

failed_cases=0

go build -o ./gait ./cmd/gait

run_case "baseline-allow" "examples/policy-test/allow.yaml" "examples/policy-test/intent.json" 0
run_case "baseline-block" "examples/policy-test/block.yaml" "examples/policy-test/intent.json" 3
run_case "baseline-require-approval" "examples/policy-test/require_approval.yaml" "examples/policy-test/intent.json" 4
run_case "prompt-injection-block" "examples/prompt-injection/policy.yaml" "examples/prompt-injection/intent_injected.json" 3
run_case "template-low-read" "examples/policy/base_low_risk.yaml" "examples/policy/intents/intent_read.json" 0
run_case "template-medium-write" "examples/policy/base_medium_risk.yaml" "examples/policy/intents/intent_write.json" 4
run_case "template-high-delete" "examples/policy/base_high_risk.yaml" "examples/policy/intents/intent_delete.json" 3

cat "$summary_path"
if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  cat "$summary_path" >>"$GITHUB_STEP_SUMMARY"
fi

if [[ "$failed_cases" -ne 0 ]]; then
  echo "policy compliance checks failed"
  exit 1
fi

echo "policy compliance checks passed"
