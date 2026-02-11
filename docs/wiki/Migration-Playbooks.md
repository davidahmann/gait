# Migration Playbooks

## Playbook 1: Demo to Team Pilot (30-120 min)

1. Run first-win:
   - `gait demo --json`
   - `gait verify run_demo --json`
2. Pick one adapter example and copy boundary pattern.
3. Add policy fixtures (`allow`, `block`, `require_approval`).
4. Enable CI checks:
   - `make test-adapter-parity`
   - `make test-skill-supply-chain`
   - `make test-acceptance`

## Playbook 2: Incident to Regression

1. Capture ticket-safe footer from run artifacts.
2. Initialize regress fixture:
   - `gait regress init --from <run_id> --json`
3. Enforce in CI:
   - `gait regress run --json --junit ./gait-out/junit.xml`

Reference: `docs/ci_regress_kit.md`

## Playbook 3: High-Risk Tool Rollout

1. Start with simulate/dry-run policy.
2. Review reasons and traces.
3. Move to enforce after deterministic pass.

Reference: `docs/policy_rollout.md`

## Exit Criteria

- No side-effecting path bypasses Gate.
- Deterministic artifacts generated for every guarded action.
- Regress and policy tests run on every PR.
