# RFC Template: Agent Framework Integration

Use this template for proposing a new framework integration while preserving Gait contract parity.

## Title

RFC: Add `<framework>` execution-boundary integration for deterministic policy enforcement

## Problem

`<framework>` executes tool actions, but deployments need deterministic policy decisions and signed evidence per action attempt.

## Proposed Integration Shape

1. Map framework payload -> `IntentRequest`.
2. Evaluate through `gait gate eval` or `gait mcp proxy`.
3. Execute side effects only on verdict `allow`.
4. Persist deterministic trace outputs in `gait-out/integrations/<framework>/`.

## Required Artifacts

- `examples/integrations/<framework>/README.md`
- `examples/integrations/<framework>/quickstart.py` (or equivalent)
- `policy_allow.yaml`
- `policy_block.yaml`

## Acceptance Checks

```bash
bash scripts/test_adapter_parity.sh
bash scripts/test_adoption_smoke.sh
```

## Security Requirements

- fail-closed on non-`allow`
- include `skill_provenance` when skills are involved
- include endpoint taxonomy fields for side-effecting actions

## Optional Enterprise Passthrough Context

When available, include:

- `context.auth_context`
- `context.credential_scopes`
- `context.environment_fingerprint`

## Out of Scope

- custom policy engines outside Go core
- bypass channels around Gait evaluation
- hosted fleet control-plane behavior in OSS path
