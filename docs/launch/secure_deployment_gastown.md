# Secure Deployment Guide: Gas Town + Gait

This guide defines the minimum safe OSS deployment shape for Gas Town worker execution.

## Goal

Ensure each worker action is policy-evaluated before execution and emits signed trace evidence.

## 1) Wire Worker Wrapper

Reference artifact:

- `examples/integrations/gastown/quickstart.py`

Required behavior:

1. normalize worker call to `IntentRequest`
2. evaluate with `gait gate eval --json`
3. execute side effects only on `allow`

## 2) Set Policy Baseline

For production-adjacent worker paths:

- prefer `default_verdict: block` + explicit allow rules
- use approval requirements for destructive operations

## 3) Preserve Deterministic Artifacts

Persist under stable paths:

- `gait-out/integrations/gastown/intent_<scenario>.json`
- `gait-out/integrations/gastown/trace_<scenario>.json`

## 4) Bridge Operational Follow-Up

For blocked traces:

```bash
bash scripts/bridge_trace_to_beads.sh \
  --trace gait-out/integrations/gastown/trace_block.json \
  --dry-run \
  --json
```

## 5) SIEM/Observability Integration

Export structured events and ingest into existing systems:

- `docs/siem_ingestion_recipes.md`

## Validation

```bash
python3 examples/integrations/gastown/quickstart.py --scenario allow
python3 examples/integrations/gastown/quickstart.py --scenario block
bash scripts/test_adapter_parity.sh
```
