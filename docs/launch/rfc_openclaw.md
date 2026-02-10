# RFC Template: OpenClaw + Gait Execution Boundary

Use this template for a GitHub Discussion or Issue in OpenClaw-adjacent communities.

## Title

RFC: Deterministic tool-call policy enforcement for OpenClaw runtime actions

## Problem

OpenClaw executes real tool actions. Existing runtime controls in most deployments are observational and content-scanning focused. Teams need deterministic execution-time allow or block behavior and portable proof artifacts for incident review.

## Proposed Integration

- install one official Gait boundary package:
  - `bash scripts/install_openclaw_skill.sh`
- route OpenClaw tool-call envelopes through:
  - `~/.openclaw/skills/gait-gate/gait_openclaw_gate.py`
- enforce fail-closed behavior on non-`allow` verdicts.

## Integration Contract

- input: OpenClaw tool-call payload
- normalization: payload -> `IntentRequest`
- decision: `gait mcp proxy --policy ... --json`
- output: deterministic decision object with:
  - `verdict`
  - `executed` (true only on `allow`)
  - `reason_codes`
  - `trace_path`

## Acceptance Checks

```bash
bash scripts/install_openclaw_skill.sh --json
bash scripts/test_openclaw_skill_install.sh
bash scripts/test_adapter_parity.sh
```

## Security And Operations Notes

- fail-closed: no side effects for `block`, `require_approval`, `dry_run`
- trace proofs: signed trace records per decision
- no hosted dependency required for local enforcement

## Out of Scope

- prompt scanning gateway replacement
- identity or vault platform replacement
- hosted control-plane dashboard
