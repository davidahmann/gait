# v2.1 Additive Readiness Contract

Status: implementation guidance for OSS `v1.x` consumers preparing for `v2.1`-scoped extensions.

This document defines what is available now, what is planned next, and how to stay compatible while those additions land.

## Goal

Prepare integrations for long-running sessions and delegation governance without breaking current `v1.x` behavior.

## Available Now (Current OSS Surface)

Current schema and SDK surfaces already support:

- intent context session correlation:
  - `context.session_id`
  - `context.request_id`
- enterprise passthrough context:
  - `context.auth_context`
  - `context.credential_scopes`
  - `context.environment_fingerprint`
- deterministic gate traces, runpacks, and regress fixtures for single-run capture flows

Reference:

- `schemas/v1/gate/intent_request.schema.json`
- `core/schema/v1/gate/types.go`
- `sdk/python/gait/models.py`

## Planned v2.1 Additions (Not Fully Shipped Yet)

The following items are planned additive extensions and should be treated as forward-looking:

- append-only session journal and checkpoint chain artifacts
- session-chain verification and checkpoint-aware diff/regress flows
- first-class delegation metadata in intent and trace contracts
- delegation token mint/verify and gate-enforced delegation checks

These are additive targets within `v1.x` compatibility rules and should not require replacing existing artifacts.

## Compatibility Rules (Normative For Producers/Consumers)

- Producers MUST keep existing required fields and semantics unchanged.
- New fields MUST be additive and optional unless a major version is introduced.
- Consumers MUST ignore unknown additive fields they do not understand.
- Consumers MUST continue to fail closed on non-`allow` or non-evaluable high-risk decisions.

## Integration Readiness Checklist

Use this now so v2.1 additions can be adopted incrementally:

1. Always set stable `context.identity`, `context.workspace`, and `context.risk_class`.
2. Set `context.session_id` for long-running or multi-step jobs.
3. Pass through `auth_context`, `credential_scopes`, and `environment_fingerprint` when available.
4. Treat runpack + trace artifacts as durable references in CI/tickets.
5. Avoid custom parsing that rejects unknown optional fields.

## Rollout Guidance

- Keep current production enforcement on existing primitives (`IntentRequest`, `GateResult`, `TraceRecord`, `Runpack`).
- Add context passthrough first; do not wait for delegation token support to improve traceability.
- Adopt future session/delegation fields in staged CI fixtures before runtime enforcement.

See also:

- `docs/contracts/primitive_contract.md`
- `docs/contracts/artifact_graph.md`
- `docs/policy_rollout.md`
- `docs/wiki/Migration-Playbooks.md`
