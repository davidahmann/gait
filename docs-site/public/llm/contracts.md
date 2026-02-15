# Gait Contracts

Stable OSS contracts include:

- **PackSpec v1**: Unified portable artifact envelope for run, job, and call evidence with Ed25519 signatures and SHA-256 manifest. Schema: `schemas/v1/pack/manifest.schema.json`.
- **ContextSpec v1**: Deterministic context evidence envelopes with privacy-aware modes and fail-closed enforcement.
- **Primitive Contract**: Four deterministic primitives — capture, enforce, regress, diagnose.
- **Intent+Receipt Spec**: Structured tool-call intent with deterministic receipt generation.
- **Endpoint Action Model**: Maps tool-call intent to policy-evaluated action outcomes.
- Artifact schemas (`schemas/v1/*`)
- Stable CLI exit codes (0 success, 2 verification failure, 5 regression drift, 6 invalid input)
- Backward-compatible readers within major version
- Deterministic zip entry ordering, fixed timestamps, canonical JSON (RFC 8785 / JCS)

References:

- `docs/contracts/packspec_v1.md`
- `docs/contracts/contextspec_v1.md`
- `docs/contracts/primitive_contract.md`
- `docs/contracts/intent_receipt_spec.md`
- `docs/contracts/endpoint_action_model.md`
