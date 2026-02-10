# Secure Deployment Guide: OpenClaw + Gait

This guide defines the minimum safe OSS deployment shape for OpenClaw users.

## Goal

Route OpenClaw actions through deterministic policy decisions with signed traces, fail-closed by default.

## 1) Install Boundary Package

```bash
bash scripts/install_openclaw_skill.sh --json
```

## 2) Set Policy Baseline

Start with explicit defaults:

- `default_verdict: block` for high-risk environments
- or `default_verdict: allow` plus targeted blocking while instrumenting

Always test policy before rollout:

```bash
gait policy test <policy.yaml> <intent.json> --json
```

## 3) Enforce Fail-Closed Runtime Path

Use skill entrypoint for all tool paths:

```bash
python3 ~/.openclaw/skills/gait-gate/gait_openclaw_gate.py \
  --policy <policy.yaml> \
  --call <tool_call.json> \
  --json
```

Non-`allow` outcomes must not execute side effects.

### Hook Point In OpenClaw Runtime

Enforcement is implemented at the tool-dispatch chokepoint, not inside the model.

- Route each tool envelope through `gait_openclaw_gate.py`.
- Execute the real tool only when returned `verdict` is `allow`.
- Treat `block`, `require_approval`, and evaluation errors as non-executable.

Reference implementation:

- `examples/integrations/openclaw/skill/gait_openclaw_gate.py`

Behavior contract:

- `executed=true` only when `verdict=allow`.
- `executed=false` for all non-`allow` outcomes.

## 4) Require Provenance For Skill-Driven Calls

Ensure intents include `skill_provenance` fields:

- `skill_name`
- `source`
- `publisher`
- digest/signature fields when available

## 5) Export Evidence For Existing Monitoring

Use local exports and ingest in existing stack:

- `docs/siem_ingestion_recipes.md`

## 6) Incident Loop

For blocked or suspicious actions:

1. preserve trace path
2. bridge to deterministic work item:
   - `bash scripts/bridge_trace_to_beads.sh --trace <trace.json> --dry-run --json`
3. convert incidents to regress fixtures

## Validation

```bash
bash scripts/test_openclaw_skill_install.sh
bash scripts/test_adapter_parity.sh
```
