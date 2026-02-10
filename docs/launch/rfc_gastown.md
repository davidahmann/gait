# RFC Template: Gas Town + Gait Worker Boundary

Use this template for proposing Gait integration into Gas Town worker execution paths.

## Title

RFC: Policy-gated worker execution for Gas Town polecat actions

## Problem

Gas Town runs parallel worker actions across code and system boundaries. Teams need deterministic execution gates and auditable evidence for each action attempt, not only post-hoc observation.

## Proposed Integration

- add wrapper path using:
  - `examples/integrations/gastown/quickstart.py`
- enforce action boundary:
  - normalize worker call -> `IntentRequest`
  - evaluate via `gait gate eval`
  - execute only when verdict is `allow`

## Integration Contract

- deterministic artifacts under:
  - `gait-out/integrations/gastown/intent_<scenario>.json`
  - `gait-out/integrations/gastown/trace_<scenario>.json`
- parity behavior:
  - `allow` -> `executed=true`
  - non-`allow` -> `executed=false`

## Acceptance Checks

```bash
python3 examples/integrations/gastown/quickstart.py --scenario allow
python3 examples/integrations/gastown/quickstart.py --scenario block
bash scripts/test_adapter_parity.sh
```

## Security And Operations Notes

- non-`allow` paths remain fail-closed
- traces are signed and verifiable
- incident traces can be bridged into deterministic work items

## Out of Scope

- orchestrator feature ownership
- centralized fleet dashboard
- prompt/content scanner replacement
