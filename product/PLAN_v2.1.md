# PLAN v2.1: Long-Running Sessions, Delegation Governance, and Invisible Enforcement

Date: 2026-02-11
Source of truth: `product/PLAN_v1.md`, `product/PLAN_v2.md`, `README.md`, `docs/architecture.md`, `docs/flows.md`, current `main` codebase
Scope: OSS CLI, schemas, adapters, and docs only (no hosted control plane)

This plan is written to be executed top-to-bottom with minimal interpretation. Each story includes concrete repo paths, commands, and acceptance criteria.

---

## Global Decisions (Locked for v2.1)

- Keep Go authoritative for schema contracts, canonicalization, hashing/signing, policy evaluation, and deterministic artifact behavior.
- Preserve offline-first defaults and fail-closed behavior for non-evaluable high-risk paths.
- Keep all contract changes additive within `v1.x`; do not break existing `runpack`, `trace`, `gate_result`, or `policy_test` readers.
- Treat long-running capture as append-only and crash-tolerant by design; never require clean process termination for artifact integrity.
- Treat delegation as explicit authorization data, not inferred metadata.
- Keep wrapper/sidecar ergonomics simple enough for framework developers integrating into non-technical user experiences.
- Keep stable exit-code semantics; any new non-zero code requires explicit contract documentation and tests.
- Keep deterministic CI lanes fast; put heavy soak/perf lanes in nightly.

---

## v2.1 Objective

Ship governance primitives needed for real-world multi-day and multi-agent execution:

1. Robust long-running session capture with incremental, verifiable checkpoints.
2. First-class delegation chain modeling in intent, policy, trace, and approvals.
3. Delegation token verification parallel to approval tokens.
4. Invisible-by-default enforcement surfaces for adapter and sidecar paths.
5. Deterministic verification, diff, regression, and CI behavior across session chains.

---

## Current Baseline (Observed)

Already in place:

- Deterministic runpack recording and verification for single-shot capture paths (`core/runpack/*`, `cmd/gait/run_record.go`).
- Long-running local interception service (`gait mcp serve`) with JSON/SSE/NDJSON endpoints per request (`cmd/gait/mcp_server.go`).
- Gate intent model includes `session_id` and strong normalization/digesting (`schemas/v1/gate/intent_request.schema.json`, `core/gate/intent.go`).
- Policy matching over identity/workspace/targets/provenance/endpoint classes is deterministic (`core/gate/policy.go`).
- Approval token issuance and deterministic validation pipeline exists (`core/gate/approval.go`, `cmd/gait/approve.go`, `cmd/gait/gate.go`).
- Adapter and sidecar integration checklist and parity scripts exist (`docs/integration_checklist.md`, `scripts/test_adapter_parity.sh`).

Gaps blocking objective:

- Runpack model is snapshot-oriented and zip-at-end, not append-only session journaling.
- No first-class session checkpoint chain artifact contract.
- No delegation chain in `IntentRequest`; policy cannot express lead-agent-to-specialist delegation constraints.
- No delegation token contract parallel to approval tokens.
- MCP serve stream endpoint emits one decision payload per request, not incremental session artifact events.
- SDK models do not expose enterprise passthrough fields and delegation metadata as first-class typed surfaces.

---

## v2.1 Exit Criteria

v2.1 is complete only when all are true:

- A crash-safe append-only session journal exists with deterministic replay ordering.
- Checkpoint runpacks can be emitted incrementally and verified independently.
- A session chain verifier validates link continuity (`prev_checkpoint_digest`) and per-checkpoint integrity.
- `IntentRequest` supports explicit delegation chain metadata and that metadata is included in intent digest normalization.
- Policy can match delegation constraints and fail closed when delegation is required but missing/invalid.
- Delegation token mint/validate flow exists with deterministic reason/error codes and signed proofs.
- Gate can enforce combined constraints: policy + delegation + approval + broker credentials.
- Python SDK, sidecar, and adapter examples can pass delegation/session fields through without custom framework logic.
- CI and nightly suites cover crash recovery, multi-day session checkpointing, and delegation rule determinism.
- Docs and runbooks include migration guidance for existing integrations.

---

## Epic V2.1.0: Contract Strategy and Compatibility Guard

Objective: add new capabilities without breaking existing OSS artifact consumers.

### Story V2.1.0.1: Define additive session and delegation contract surface

Tasks:

- Define new schema artifacts for session capture:
  - `schemas/v1/runpack/session_journal.schema.json`
  - `schemas/v1/runpack/session_checkpoint.schema.json`
  - `schemas/v1/runpack/session_chain.schema.json`
- Extend existing schemas additively for delegation metadata:
  - `schemas/v1/gate/intent_request.schema.json` (new optional delegation object)
  - `schemas/v1/gate/trace_record.schema.json` (new optional delegation summary/reference)
  - `schemas/v1/gate/approval_token.schema.json` (optional delegation binding digest fields if needed)
- Add matching Go types under:
  - `core/schema/v1/runpack/`
  - `core/schema/v1/gate/`

Acceptance criteria:

- Existing fixtures for current v1 schema contracts continue to validate unchanged.
- New schema fixtures validate with Draft 2020-12 validator in CI.
- Unknown additive fields are tolerated by consumers where contract says append-only.

### Story V2.1.0.2: Compatibility harness for v1 readers and new artifacts

Tasks:

- Extend compatibility tests in `core/schema/validate/validate_test.go` and relevant package tests to assert:
  - v1 readers ignore unknown additive fields
  - new artifacts do not alter old verify/diff behavior when not present
- Add contract notes to:
  - `docs/contracts/primitive_contract.md`
  - `docs/contracts/artifact_graph.md`

Acceptance criteria:

- Contract tests prove backward compatibility for existing runpack and trace consumers.
- Compatibility notes are explicit about additive-only behavior within `v1.x`.

---

## Epic V2.1.1: Long-Running Session Capture (Append-Only + Crash-Tolerant)

Objective: capture multi-day execution safely without requiring clean shutdown.

### Story V2.1.1.1: Session journal writer and append protocol

Tasks:

- Implement append-only session journal in `core/runpack/`:
  - deterministic event ordering by sequence id
  - atomic append/flush semantics for crash consistency
  - explicit session metadata: `session_id`, `run_id`, `started_at`, `producer_version`
- Add CLI helpers under `cmd/gait/run.go` and/or new `cmd/gait/run_session.go`:
  - `gait run session start`
  - `gait run session append`
  - `gait run session status`
- Ensure journal writes use existing atomic-write and lock strategy patterns (`core/fsx`, ADR alignment).

Repo paths:

- `core/runpack/record.go`
- `core/runpack/read.go`
- `core/runpack/replay.go`
- `cmd/gait/run.go`
- `cmd/gait/main_test.go`

Acceptance criteria:

- Simulated crash during append preserves all previously committed entries.
- Re-open and continue append on same session yields stable sequence continuity.
- Empty/partial line corruption is rejected deterministically with actionable errors.

### Story V2.1.1.2: Periodic sealed checkpoint runpacks

Tasks:

- Implement checkpoint emitter in `core/runpack/`:
  - materialize checkpoint runpack from journal prefix
  - include checkpoint metadata (session id, checkpoint index, covered sequence range)
  - include `prev_checkpoint_digest` linkage
- Add CLI command:
  - `gait run session checkpoint --session <id> --out <path>`
- Keep deterministic zip guarantees using `core/zipx`.

Repo paths:

- `core/runpack/record.go`
- `core/runpack/verify.go`
- `core/zipx/zipx.go`
- `cmd/gait/run.go`

Acceptance criteria:

- Same journal prefix produces identical checkpoint zip bytes.
- Checkpoint `N` cryptographically links to `N-1` digest.
- If session dies before final checkpoint, earlier checkpoints remain independently verifiable.

### Story V2.1.1.3: MCP serve integration for session emission

Tasks:

- Extend `gait mcp serve` request model with session-aware fields:
  - `session_id`
  - optional `checkpoint_interval` or explicit checkpoint trigger semantics
- Persist per-decision events into session journal while preserving current trace outputs.
- Keep endpoint parity across `/v1/evaluate`, `/v1/evaluate/sse`, `/v1/evaluate/stream`.

Repo paths:

- `cmd/gait/mcp_server.go`
- `cmd/gait/mcp_server_test.go`
- `core/mcp/exporters.go`

Acceptance criteria:

- Session-aware requests produce journal updates and traces in one deterministic flow.
- Stream/SSE/JSON endpoints return consistent verdict semantics.
- Non-session requests preserve current behavior.

---

## Epic V2.1.2: Session Chain Verify, Diff, and Regress Compatibility

Objective: make session checkpoints first-class citizens in existing verification and CI workflows.

### Story V2.1.2.1: Session chain verification command

Tasks:

- Add `gait verify session-chain --chain <session_chain.json> [--require-signature] ...`.
- Verify:
  - each checkpoint runpack integrity/signature
  - digest linkage continuity
  - monotonic checkpoint sequence and covered ranges
- Emit stable JSON with machine-readable failure reason codes.

Repo paths:

- `cmd/gait/verify.go`
- `core/runpack/verify.go`
- `core/runpack/verify_test.go`

Acceptance criteria:

- Any missing/altered checkpoint causes deterministic verify failure.
- Valid chain returns `ok=true` with stable summary fields.

### Story V2.1.2.2: Diff and inspect support for checkpoint sequences

Tasks:

- Extend `gait run inspect` to optionally inspect a session chain summary.
- Add diff support for checkpoint-to-checkpoint and chain-to-chain comparisons.
- Keep deterministic ordering and output for mixed old/new artifacts.

Repo paths:

- `cmd/gait/run_inspect.go`
- `core/runpack/diff.go`
- `core/runpack/diff_test.go`

Acceptance criteria:

- Same compare inputs produce identical diff output across OSes.
- Existing `runpack.zip` diff behavior remains unchanged.

### Story V2.1.2.3: Regress fixture initialization from session chains

Tasks:

- Extend `gait regress init` to accept session-chain source and choose checkpoint pinning strategy.
- Store fixture metadata with explicit checkpoint reference.
- Ensure `gait regress run` remains deterministic when based on checkpoint fixtures.

Repo paths:

- `core/regress/init.go`
- `core/regress/run.go`
- `schemas/v1/regress/regress_result.schema.json` (additive fields if required)

Acceptance criteria:

- Regress fixtures initialized from checkpoints pass deterministic replay graders.
- Drift reasons identify checkpoint id and run/session reference clearly.

---

## Epic V2.1.3: Delegation-Aware Intent and Policy Model

Objective: model multi-agent delegation as first-class policy input.

### Story V2.1.3.1: Add delegation chain fields to IntentRequest

Tasks:

- Add optional `delegation` object to intent schema and Go type containing:
  - requester agent identity
  - delegator chain (ordered)
  - delegation scope class
  - delegation token ref(s)
  - issued/expiry metadata where applicable
- Extend normalization/digest logic so delegation is canonicalized and bound into `intent_digest`.
- Add fixtures for equivalent vs non-equivalent delegation chains.

Repo paths:

- `schemas/v1/gate/intent_request.schema.json`
- `core/schema/v1/gate/types.go`
- `core/gate/intent.go`
- `core/gate/intent_test.go`

Acceptance criteria:

- Equivalent delegation payloads normalize to identical digest.
- Reordered or modified chain elements produce different digest.

### Story V2.1.3.2: Policy matcher support for delegation constraints

Tasks:

- Extend policy schema and evaluator match surface for delegation:
  - allowed delegator identities
  - allowed delegate identities
  - max delegation depth
  - allowed delegation scopes
  - optional cross-workspace delegation rules
- Add fail-closed toggle for missing delegation metadata on high-risk multi-agent rules.

Repo paths:

- `schemas/v1/gate/policy.schema.json`
- `core/gate/policy.go`
- `core/gate/policy_test.go`
- `examples/policy/` (new delegation fixtures)

Acceptance criteria:

- Delegation-policy fixtures produce deterministic allow/block/require_approval outcomes.
- Missing required delegation metadata fails closed for configured high-risk classes.

### Story V2.1.3.3: Delegation visibility in traces and test outputs

Tasks:

- Add additive delegation summary/reference fields to trace output.
- Expose matched delegation rule context in policy test output where relevant.

Repo paths:

- `schemas/v1/gate/trace_record.schema.json`
- `core/gate/trace.go`
- `core/policytest/run.go`
- `schemas/v1/policytest/policy_test_result.schema.json`

Acceptance criteria:

- Trace records include delegation metadata when provided.
- Existing trace readers remain compatible with absent delegation fields.

---

## Epic V2.1.4: Delegation Tokens and Authorization Chain Enforcement

Objective: add cryptographic proof for agent-to-agent delegation.

### Story V2.1.4.1: Delegation token schema, mint, and verify

Tasks:

- Add schema and type:
  - `schemas/v1/gate/delegation_token.schema.json`
  - `core/schema/v1/gate/types.go` additions
- Implement mint/verify in `core/gate/` mirroring approval-token rigor:
  - binds delegator, delegate, scope, intent/policy context where required, TTL
  - signed with existing signing primitives
  - deterministic error codes for expired/mismatch/scope/signature issues
- Add CLI surface:
  - `gait delegate mint ...`
  - `gait delegate verify ...`

Repo paths:

- `core/gate/approval.go` (refactor shared token helpers as needed)
- `cmd/gait/` (new `delegate.go`)
- `cmd/gait/main_test.go`

Acceptance criteria:

- Invalid/expired/mismatched tokens are rejected with stable reason codes.
- Valid token verifies offline deterministically.

### Story V2.1.4.2: Gate evaluation integration for delegation tokens

Tasks:

- Extend `gait gate eval` with delegation token inputs similar to approval token chain handling.
- Validate delegation token chain before honoring delegated high-risk actions.
- Emit delegation audit artifact (parallel to approval audit) when delegation is in scope.

Repo paths:

- `cmd/gait/gate.go`
- `core/gate/` (new delegation audit helpers)
- `schemas/v1/gate/` (delegation audit schema if needed)

Acceptance criteria:

- Gate blocks deterministically when required delegation token is missing/invalid.
- Gate allows only when policy + delegation + approval requirements are all satisfied.

---

## Epic V2.1.5: Fail-Closed Multi-Agent Enforcement in MCP and Runtime Boundary

Objective: harden runtime boundary behavior for autonomous multi-agent workloads.

### Story V2.1.5.1: Strict context requirements in production profile

Tasks:

- In `oss-prod`, require explicit context identity/workspace/session metadata for MCP proxy/serve calls.
- Remove implicit permissive defaults for high-risk profiles.
- Emit actionable errors and stable reason codes for missing context.

Repo paths:

- `core/mcp/proxy.go`
- `cmd/gait/mcp.go`
- `cmd/gait/mcp_server.go`
- `cmd/gait/mcp_test.go`

Acceptance criteria:

- Production profile fails closed when identity/workspace/session context is absent.
- Standard profile remains backward-compatible for local demo/test flows.

### Story V2.1.5.2: Deterministic decision/event stream semantics

Tasks:

- Ensure exported JSONL/OTEL event streams include session and delegation references when available.
- Keep event ordering deterministic and line-oriented append behavior stable.

Repo paths:

- `core/mcp/exporters.go`
- `docs/siem_ingestion_recipes.md`

Acceptance criteria:

- Exported events can be correlated to trace/session/delegation artifacts by immutable IDs.
- Existing ingestion recipes remain valid; new fields are additive.

---

## Epic V2.1.6: Invisible Governance Integration Surface (SDK + Adapters + Sidecar)

Objective: make secure integration trivial for framework developers and invisible to end users.

### Story V2.1.6.1: Python SDK model and adapter surface expansion

Tasks:

- Extend SDK models for:
  - `auth_context`, `credential_scopes`, `environment_fingerprint`
  - delegation metadata and session metadata
- Extend client/adapter methods to pass through new fields without changing enforcement semantics.

Repo paths:

- `sdk/python/gait/models.py`
- `sdk/python/gait/client.py`
- `sdk/python/gait/adapter.py`
- `sdk/python/tests/*`

Acceptance criteria:

- SDK can emit valid extended intents without manual dict surgery.
- Existing SDK usage remains source-compatible for current integrations.

### Story V2.1.6.2: Sidecar and adapter parity for delegation/session metadata

Tasks:

- Update canonical sidecar and integration examples to include optional delegation/session fields.
- Extend adapter parity tests for these new paths.

Repo paths:

- `examples/sidecar/gate_sidecar.py`
- `examples/integrations/*/quickstart.py`
- `scripts/test_adapter_parity.sh`
- `docs/integration_checklist.md`

Acceptance criteria:

- Parity suite confirms same non-`allow` execution semantics across adapters with delegation/session payloads.
- Non-Python sidecar path remains one-command integration pattern.

---

## Epic V2.1.7: Policy Defaults for Non-Technical User Workloads

Objective: provide safe default governance packs for "describe outcome, agent executes" usage.

### Story V2.1.7.1: Add opinionated policy templates for vibe-working workloads

Tasks:

- Add baseline templates emphasizing:
  - strict egress controls
  - explicit approval for destructive or integration-wiring operations
  - dataflow checks for external/tool-output taint to sensitive destinations
  - delegation requirement controls for multi-agent tool classes
- Add template command support if needed through `gait policy init`.

Repo paths:

- `cmd/gait/policy_templates/`
- `cmd/gait/policy.go`
- `examples/policy/`

Acceptance criteria:

- Template-driven policy packs block known prompt-injection-style exfil paths in fixtures.
- Templates are deterministic and pass validate/fmt/test/simulate workflows.

### Story V2.1.7.2: Fixture corpus for injection + delegation cross-cases

Tasks:

- Add fixture matrix covering:
  - benign user outcome requests
  - injected exfiltration attempts
  - valid delegation with constrained scope
  - invalid delegation escalation attempts
- Wire into policy compliance script and acceptance suites.

Repo paths:

- `examples/prompt-injection/`
- `examples/policy/intents/`
- `scripts/policy_compliance_ci.sh`

Acceptance criteria:

- CI fails on regression in block/approval behavior for this fixture matrix.
- Reason code output remains stable and documented.

---

## Epic V2.1.8: CI, Soak, Performance, and UAT Expansion

Objective: prove deterministic behavior under multi-day and concurrent conditions.

### Story V2.1.8.1: Unit/integration suite expansion for sessions and delegation

Tasks:

- Add unit tests for:
  - session append ordering
  - checkpoint linkage
  - delegation normalization/digesting
  - delegation token verification errors
- Add integration tests for:
  - session chain verify
  - regress from checkpoint fixtures
  - gate eval with combined approval + delegation requirements

Repo paths:

- `core/runpack/*_test.go`
- `core/gate/*_test.go`
- `cmd/gait/*_test.go`
- `internal/integration/` (if used)

Acceptance criteria:

- Deterministic test outcomes across Linux/macOS/Windows CI matrix.
- Coverage for touched core packages remains `>= 85%`.

### Story V2.1.8.2: Nightly soak and contention tests

Tasks:

- Add nightly soak tests for long-running session append/checkpoint loops.
- Add contention tests for concurrent session writers and shared state lock paths.
- Keep bounded deterministic timeout behavior aligned with existing ADR lock strategy.

Repo paths:

- `scripts/test_hardening_acceptance.sh`
- `docs/test_cadence.md`
- `.github/workflows/adoption-nightly.yml` (or equivalent)

Acceptance criteria:

- Nightly lane detects session corruption, lock starvation, or non-deterministic ordering regressions.
- Failures are machine-classifiable and actionable.

### Story V2.1.8.3: Runtime SLO budget updates

Tasks:

- Add budget checks for new critical paths:
  - session checkpoint emit
  - session chain verify
  - gate eval with delegation token verification
- Update perf budget config and bench harness.

Repo paths:

- `perf/runtime_slo_budgets.json`
- `docs/slo/runtime_slo.md`
- benchmark harness files under `perf/`

Acceptance criteria:

- New budgets enforced in CI/nightly where appropriate.
- No silent latency regressions on governance-critical paths.

---

## Epic V2.1.9: Documentation, Migration, and Operator Playbooks

Objective: make rollout low-risk for existing adopters.

### Story V2.1.9.1: Contract docs and architecture updates

Tasks:

- Update architecture and flow docs with session-chain and delegation-token flows.
- Update primitive/artifact contracts to include new optional fields/artifacts.
- Document fail-closed semantics for missing delegation data.

Repo paths:

- `docs/architecture.md`
- `docs/flows.md`
- `docs/contracts/primitive_contract.md`
- `docs/contracts/artifact_graph.md`

Acceptance criteria:

- Docs accurately reflect command and schema behavior.
- No conflict between docs and command help/tests.

### Story V2.1.9.2: Integration and rollout playbooks

Tasks:

- Update integration checklist with session/delegation onboarding steps.
- Update policy rollout guide with delegation rollout stages and fixture gates.
- Add migration playbook for existing adapters adopting additive fields.

Repo paths:

- `docs/integration_checklist.md`
- `docs/policy_rollout.md`
- `docs/wiki/Migration-Playbooks.md`
- `README.md` (surgical overview updates)

Acceptance criteria:

- Existing users can adopt v2.1 features incrementally without breaking existing production lanes.
- New users can integrate with one canonical secure boundary path in under two hours.

---

## Validation Plan (Local)

Run in order:

1. `gofmt -w .`
2. `go test ./core/runpack ./core/gate ./core/mcp ./core/regress ./core/schema/validate ./cmd/gait`
3. `go test ./...`
4. `(cd sdk/python && PYTHONPATH=. uv run --python 3.13 --extra dev pytest -q)`
5. `bash scripts/policy_compliance_ci.sh`
6. `bash scripts/test_adapter_parity.sh`
7. `bash scripts/test_v1_acceptance.sh`
8. `bash scripts/test_v1_7_acceptance.sh`
9. `bash scripts/test_v1_8_acceptance.sh`
10. `bash scripts/test_hardening_acceptance.sh`
11. `make bench-budgets`
12. `bash scripts/test_uat_local.sh --skip-brew`

If any command fails, fix and rerun from the failing step.

---

## Release/Delivery Checklist

- [ ] `product/PLAN_v2.1.md` committed with implementation cycle.
- [ ] Schema/type changes include fixtures and validator coverage.
- [ ] Command help and tests updated for any new flags/subcommands.
- [ ] CI lanes green (main + nightly where touched).
- [ ] UAT local run green after CI green.
- [ ] Changelog updated only if release/tag cut is requested in this cycle.

---

## Execution Order (Strict)

1. Land contract/schema/type additions with compatibility tests.
2. Implement session journal + checkpoint pipeline.
3. Implement session-chain verify/diff/regress compatibility.
4. Implement delegation fields + policy matcher extensions.
5. Implement delegation token mint/verify + gate integration.
6. Update MCP/runtime enforcement strictness and event exports.
7. Update SDK/adapters/sidecar and parity tests.
8. Add policy templates + injection/delegation fixture matrix.
9. Expand CI/soak/perf/UAT lanes.
10. Complete docs + migration playbooks.

---

## Definition of Done (applies to every story)

- Code is formatted and linted (`make fmt`, `make lint`).
- Tests are added/updated and passing (`make test`, integration/e2e as relevant).
- Any new artifact/schema has:
  - JSON Schema under `schemas/v1/`
  - matching Go type under `core/schema/v1/`
  - validator + valid/invalid fixtures
- `--json` outputs and reason codes are stable and tested.
- Core workflows remain offline-first.
- Production/high-risk failure modes remain fail-closed and deterministic.
- No policy or authorization logic is moved into Python/adapters.
