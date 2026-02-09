# CI Regress Kit (Epic A5.1)

This kit makes incident-to-regression checks turnkey in CI.

Canonical default path:

- `.github/workflows/adoption-regress-template.yml`

## GitHub Actions Template

Use `.github/workflows/adoption-regress-template.yml` as the baseline workflow.

Template flow:

1. Build local `gait` binary.
2. Restore fixture (`fixtures/run_demo/runpack.zip` + `gait.yaml`) or initialize from a run artifact.
3. Run `gait regress run --json --junit=...`.
4. Run endpoint policy fixture checks (`allow`, `block`, `require_approval`).
5. Run skill provenance verification checks.
6. Fail with stable exit codes (`5` for deterministic regression failure).
7. Upload `regress_result.json`, `junit.xml`, and fixture artifacts.

## Generic Shell CI Snippet (Compatibility Only)

Use this in non-GitHub providers (Jenkins, Buildkite, CircleCI, etc.):

```bash
set -euo pipefail

go build -o ./gait ./cmd/gait
mkdir -p gait-out

if [[ ! -f fixtures/run_demo/runpack.zip || ! -f gait.yaml ]]; then
  ./gait demo
  ./gait regress init --from run_demo --json
fi

set +e
./gait regress run --json --junit=./gait-out/junit.xml > ./gait-out/regress_result.json
status=$?
set -e

if [[ "$status" -eq 0 ]]; then
  echo "regress pass"
elif [[ "$status" -eq 5 ]]; then
  echo "regress fail (stable exit code 5)"
  exit 5
else
  echo "unexpected regress exit code: $status"
  exit "$status"
fi

# Endpoint policy fixture checks.
./gait policy test examples/policy/endpoint/allow_safe_endpoints.yaml examples/policy/endpoint/fixtures/intent_allow.json --json
set +e
./gait policy test examples/policy/endpoint/block_denied_endpoints.yaml examples/policy/endpoint/fixtures/intent_block.json --json
block_status=$?
./gait policy test examples/policy/endpoint/require_approval_destructive.yaml examples/policy/endpoint/fixtures/intent_destructive.json --json
approval_status=$?
set -e
if [[ "$block_status" -ne 3 ]]; then
  echo "endpoint block fixture exit mismatch: $block_status"
  exit 1
fi
if [[ "$approval_status" -ne 4 ]]; then
  echo "endpoint approval fixture exit mismatch: $approval_status"
  exit 1
fi

# Skill provenance verification path.
bash scripts/test_skill_supply_chain.sh
```

Artifacts to retain:

- `gait-out/regress_result.json`
- `gait-out/junit.xml`
- `gait.yaml`
- `fixtures/`

## Recommended PR Gate

- Run regress on pull requests that touch policy, runpack, regress, gate, and fixtures.
- Require pass (`0`) for merge.
- Treat `5` as deterministic regression failure requiring either code fix or fixture update with explicit review.

Note:

- Keep this file as the only source for CI snippet variants.
- Prefer the workflow template above to avoid drift across docs.
