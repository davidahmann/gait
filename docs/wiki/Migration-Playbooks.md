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

## Playbook 4: v2.1 Session + Delegation Adoption

1. Add additive intent fields in your runtime adapter:
   - `context.session_id`
   - `context.auth_context`
   - `context.credential_scopes`
   - `context.environment_fingerprint`
   - `delegation.*` when delegated execution is used
2. Validate adapter parity:
   - `bash scripts/test_adapter_parity.sh`
3. Enable long-running checkpoint capture for runtime lanes that execute continuously:
   - `gait run session start ...`
   - `gait run session append ...`
   - `gait run session checkpoint ...`
   - `gait verify session-chain --chain ... --json`
4. Add policy fixtures for delegation-valid, delegation-invalid, and tainted-egress cases.

References:

- `docs/integration_checklist.md`
- `docs/policy_rollout.md`

## Exit Criteria

- No side-effecting path bypasses Gate.
- Deterministic artifacts generated for every guarded action.
- Regress and policy tests run on every PR.
