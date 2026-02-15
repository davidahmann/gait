---
title: "Intent+Receipt Conformance"
description: "Conformance test vectors for intent request and gate receipt producers and consumers."
---

# Intent + Receipt Conformance (Normative v1.x)

Status: normative for OSS v1.x producers/consumers that claim boundary-intent and receipt continuity.

This contract proves a full chain:

1. intent shape conformance
2. gate digest continuity (`intent_digest`, `policy_digest`)
3. runpack receipt continuity (`refs.receipts`)
4. ticket-footer continuity (`gait run receipt`)

See also:

- `docs/contracts/primitive_contract.md`
- `docs/contracts/artifact_graph.md`

## Required Schemas

- `schemas/v1/gate/intent_request.schema.json`
- `schemas/v1/gate/trace_record.schema.json`
- `schemas/v1/runpack/refs.schema.json`

## Minimal Normative Command Sequence

From repo root:

```bash
go build -o ./gait ./cmd/gait
./gait demo --json > ./gait-out/intent_receipt_demo.json
./gait gate eval \
  --policy examples/policy/base_low_risk.yaml \
  --intent examples/policy/intents/intent_read.json \
  --trace-out ./gait-out/intent_receipt_trace.json \
  --json > ./gait-out/intent_receipt_gate.json
./gait run receipt --from run_demo --json > ./gait-out/intent_receipt_footer.json
```

## Required Fields And Continuity Checks

### Intent request conformance

`intent_request.schema.json` required fields MUST remain present:

- `schema_id`
- `schema_version`
- `created_at`
- `producer_version`
- `tool_name`
- `args`
- `targets`
- `context`

### Gate digest continuity

`gait gate eval --json` output MUST include:

- `verdict`
- `intent_digest`
- `policy_digest`
- `trace_id`
- `trace_path`

`intent_digest` and `policy_digest` MUST be 64-char sha256 hex.

### Runpack receipt continuity

`runpack.zip` -> `refs.json` MUST include:

- `schema_id = gait.runpack.refs`
- `run_id`
- `receipts` array

Each receipt object MUST include:

- `ref_id`
- `source_type`
- `source_locator`
- `query_digest`
- `content_digest`
- `retrieved_at`
- `redaction_mode`

### Ticket footer continuity

`gait run receipt --from <run> --json` MUST include:

- `ok=true`
- `run_id`
- `manifest_digest`
- `ticket_footer`

`ticket_footer` MUST match contract form:

```text
GAIT run_id=<run_id> manifest=sha256:<64-hex> verify="gait verify <run_id>"
```

## CI Conformance Gate

Canonical script:

```bash
bash scripts/test_intent_receipt_conformance.sh ./gait
```

This script is wired into `make test-contracts` and CI contract lanes.
