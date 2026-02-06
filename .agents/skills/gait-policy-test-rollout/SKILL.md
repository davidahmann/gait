---
name: gait-policy-test-rollout
description: Evaluate Gait policy changes safely with deterministic fixture tests and rollout simulation. Use when asked to validate allow or block behavior, check reason codes, compare policy outcomes, or plan enforce-by-stage rollout.
disable-model-invocation: true
---
# Policy Test Rollout

Execute this workflow to validate policy behavior before enforcement.

## Workflow

1. Require both files:
   - `<policy.yaml>`
   - `<intent_fixture.json>`
2. Run deterministic policy test:
   - `gait policy test <policy.yaml> <intent_fixture.json> --json`
3. Parse and report fields:
   - `ok`, `policy_digest`, `intent_digest`, `verdict`, `reason_codes`, `violations`, `summary`
4. If rollout simulation is requested, run:
   - `gait gate eval --policy <policy.yaml> --intent <intent_fixture.json> --simulate --json`
5. Return structured rollout recommendation:
   - current verdict
   - blocking reasons or required approvals
   - suggested next stage (`observe`, `require_approval`, `enforce`)

## Exit Code Contract

- `0`: allow
- `3`: block
- `4`: require approval
- `6`: invalid input

## Safety Rules

- Never bypass policy test by inferring results from YAML text.
- For replay workflows, prefer `gait run replay` (stub mode default); require explicit unsafe flags for real tool replay.
- Never claim a policy digest or verdict without command output.
- Keep simulation and enforcement clearly separated in reporting.

## Determinism Rules

- Always use `--json` outputs.
- Report exact `reason_codes` and `violations` as emitted.
- Preserve fixture-based evaluation flow for repeatability.
