# PLAN v2.2: Gait OSS Prime-Time Hardening (Zero-Ambiguity Execution Plan)

Date: 2026-02-11
Source of truth: `product/PLAN_v1.md`, `product/PLAN_v2.1.md`, `gait-out/gtm_1/CHAOS.md`, `docs/architecture.md`, `docs/slo/runtime_slo.md`
Scope: OSS only (runtime safety, durability, operability, and adoption hardening)

This plan is written to be executed top-to-bottom with minimal interpretation. Each story includes concrete repo paths, commands, and acceptance criteria.

---

## Global Decisions (Locked for v2.2)

- v2.2 is a hardening release; no hosted UI/control-plane scope is introduced.
- Go remains authoritative for policy decisions, schemas, canonicalization, signing, verification, and runtime boundary behavior.
- Python remains a thin adoption layer; no policy logic duplication in SDK/adapters.
- Artifact/schema evolution remains additive in `v1.x`; no breaking schema changes.
- Default posture for production guidance is fail-closed and explicit (`oss-prod`, strict context, production key mode).
- `mcp serve` is treated as a security boundary component, not only a convenience transport.
- Long-running session behavior must remain deterministic and crash-tolerant at multi-day scale.
- Evidence quality (trace uniqueness + timeline semantics) is product-critical, not cosmetic.
- Chaos regressions become release-blocking for prime-time paths.

---

## v2.2 Objective

Close all OSS gaps identified in `gait-out/gtm_1/CHAOS.md` (F-01 through F-12) and ship a release that is safe for broad OSS production adoption without requiring enterprise features.

---

## Finding Coverage Matrix (Must Close)

| Finding | Risk Theme | v2.2 Stories |
|---|---|---|
| F-01 | Concurrent JSONL corruption | 2.1, 2.2, 8.1 |
| F-02 | Unauthenticated/unsafe service boundary | 1.1, 1.3, 1.5, 8.2 |
| F-03 | Unbounded request body | 1.2, 8.3 |
| F-04 | Session lock timeout under swarm load | 3.2, 3.4, 8.4 |
| F-05 | Session append scaling drift | 3.1, 3.3, 8.5 |
| F-06 | Trace overwrite on repeated identical actions | 4.1, 4.2, 8.6 |
| F-07 | Synthetic historical timestamp default (`1980`) in runtime traces | 4.1, 4.3 |
| F-08 | Easy fallback to weak runtime context in `standard` | 5.1, 5.2, 5.3 |
| F-09 | HTTP 200 on non-allow easy to misuse | 1.4, 6.2 |
| F-10 | SDK subprocess has no timeout budget | 6.1, 6.3 |
| F-11 | Telemetry append failures silently dropped | 2.3, 5.4 |
| F-12 | No retention/rotation controls for service-mode artifacts | 7.1, 7.2, 7.3 |

---

## Epic 0: Program Guardrails and Compatibility Envelope

Objective: define hardening boundaries before touching runtime behavior.

### Story 0.1: Add v2.2 hardening contract doc

Tasks:
- Add `docs/hardening/v2_2_contract.md` with:
  - threat model assumptions for OSS runtime deployment
  - mandatory/optional hardening knobs
  - compatibility policy for new flags/fields
- Add cross-links from `README.md`, `docs/architecture.md`, `docs/integration_checklist.md`.

Acceptance criteria:
- Hardening contract is discoverable from README within two clicks.
- Contract explicitly states which defaults are dev convenience vs production posture.

### Story 0.2: Backward-compatibility test gate for additive changes

Tasks:
- Extend contract tests to ensure old consumers tolerate new additive fields in:
  - trace records
  - service responses
  - operational telemetry outputs
- Update `scripts/test_contracts.sh` and `scripts/test_ent_consumer_contract.sh` as needed.

Acceptance criteria:
- Contract suite passes on all OS CI targets.
- No existing schema fixture or parser path breaks in `v1.x` compatibility mode.

---

## Epic 1: MCP Serve Security Boundary Hardening

Objective: remove high-risk boundary weaknesses in service mode.

### Story 1.1: Add explicit service authentication modes

Tasks:
- Extend `gait mcp serve` with auth mode flags:
  - `--auth-mode off|token` (default remains safe for local loopback)
  - `--auth-token-env <VAR>` for bearer token verification
- Enforce auth for non-loopback listen addresses by default unless explicitly overridden.

Repo paths:
- `cmd/gait/mcp_server.go`
- `cmd/gait/mcp_server_test.go`
- `docs/flows.md`

Acceptance criteria:
- Non-loopback bind without auth mode/token fails closed with invalid input error.
- Loopback local dev path remains simple and documented.

### Story 1.2: Enforce body size and decode safety limits

Tasks:
- Wrap request body with max-byte enforcement (`http.MaxBytesReader`).
- Add CLI flag: `--max-request-bytes` with conservative default.
- Return deterministic `413` payload on oversized body.

Repo paths:
- `cmd/gait/mcp_server.go`
- `cmd/gait/mcp_server_test.go`

Acceptance criteria:
- Oversized request receives `413` and structured error JSON.
- Normal requests unaffected; latency budget impact remains acceptable.

### Story 1.3: Lock down client-provided output paths

Tasks:
- Add path policy for service outputs:
  - allow only server-managed dirs by default (`trace-dir`, `session-dir`, `runpack-dir`)
  - optional explicit escape hatch flag for trusted local testing
- Reject absolute/untrusted client `trace_path`, `session_journal`, `runpack_out` by default.

Repo paths:
- `cmd/gait/mcp_server.go`
- `cmd/gait/mcp_server_test.go`
- `core/gate/trace.go`
- `core/runpack/session.go`

Acceptance criteria:
- Caller cannot write artifacts outside configured server directories in default mode.
- Error reason is stable and actionable.

### Story 1.4: Add strict HTTP semantics mode for non-allow verdicts

Tasks:
- Add optional flag: `--http-verdict-status strict|compat`.
- `compat`: preserve current `200` behavior.
- `strict`: map non-allow verdicts to non-2xx status while preserving JSON contract payload.
- Add explicit `executed=false` field in response contract.

Repo paths:
- `cmd/gait/mcp_server.go`
- `cmd/gait/mcp_server_test.go`
- `schemas/v1/gate/gate_result.schema.json` (if needed additively via service response schema)

Acceptance criteria:
- Strict mode is fully tested for `block`, `require_approval`, `dry_run`.
- Existing integrations using compat mode keep current behavior.

### Story 1.5: Add service deployment safety checks in doctor

Tasks:
- Extend doctor checks to detect risky service configs:
  - non-loopback without auth
  - weak request size limit
  - path override mode enabled in production profile

Repo paths:
- `cmd/gait/doctor.go`
- `core/doctor/doctor.go`
- `core/doctor/doctor_test.go`

Acceptance criteria:
- `gait doctor --production-readiness` fails when unsafe service settings are present.

---

## Epic 2: Append Durability and Telemetry Integrity

Objective: make JSONL append paths corruption-safe under concurrent writes.

### Story 2.1: Introduce shared concurrency-safe JSONL append utility

Tasks:
- Add shared append helper in `core/fsx` or `core/mcp` that guarantees atomic line writes under concurrency.
- Ensure one syscall per event record (`line + '\n'`), plus optional file lock strategy.

Repo paths:
- `core/fsx/`
- `core/mcp/exporters.go`

Acceptance criteria:
- High-concurrency stress test yields zero malformed lines.

### Story 2.2: Migrate all JSONL append call sites

Tasks:
- Migrate:
  - `core/mcp/exporters.go`
  - `core/scout/adoption.go`
  - `core/scout/operational.go`
- Preserve file permissions and local-path constraints.

Acceptance criteria:
- Existing unit tests pass unchanged where behavior is contract-compatible.
- New concurrency tests added for each module.

### Story 2.3: Stop silently dropping telemetry write failures

Tasks:
- Replace ignored append errors in command entrypoint with deterministic surfaced signals:
  - structured warning emission
  - optional debug stderr output
  - internal counters for dropped events

Repo paths:
- `cmd/gait/main.go`
- `core/scout/*.go`

Acceptance criteria:
- Failure to write telemetry is observable without breaking main command correctness.
- No panic/crash introduced on telemetry backpressure.

---

## Epic 3: Session Scaling and Contention Hardening

Objective: keep long-running sessions reliable under high throughput and long duration.

### Story 3.1: Remove per-append full journal read from hot path

Tasks:
- Add session state index/cache file with deterministic fields:
  - last sequence
  - checkpoint cursor
  - stable digest anchors
- Use lock-protected incremental updates instead of full journal replay on each append.

Repo paths:
- `core/runpack/session.go`
- `core/runpack/session_test.go`

Acceptance criteria:
- Append latency remains flat within target budget at long journal lengths.

### Story 3.2: Improve lock strategy for swarm contention

Tasks:
- Make lock timeout/retry tunable and profile-aware.
- Introduce fairer lock acquisition/backoff policy.
- Return structured contention diagnostics for operators.

Repo paths:
- `core/runpack/session.go`
- `internal/integration/contention_test.go`

Acceptance criteria:
- Contention failure rate under defined swarm workload is below budget threshold.

### Story 3.3: Add deterministic compaction/checkpoint pruning flow

Tasks:
- Implement session compaction command that preserves chain verifiability while trimming hot journal size.
- Document safe compaction cadence for long-running runs.

Repo paths:
- `cmd/gait/run.go` (session subcommands)
- `core/runpack/session.go`
- `docs/flows.md`

Acceptance criteria:
- Compaction does not change verification outcome.
- Compacted and uncompacted chains verify identically.

### Story 3.4: Strengthen session soak and contention integration tests

Tasks:
- Expand `scripts/test_session_soak.sh` and integration tests with swarm profiles and thresholds.
- Add pass/fail budgets for contention and append latency drift.

Acceptance criteria:
- Soak suite is deterministic and reproducible in CI/nightly.

---

## Epic 4: Trace Artifact Quality and Timeline Semantics

Objective: improve evidence integrity and operational usefulness without breaking determinism contracts.

### Story 4.1: Add distinct runtime event identity and observed timestamp

Tasks:
- Keep deterministic `trace_id` as decision identity.
- Add additive fields to trace schema/type:
  - `event_id` (unique per emission)
  - `observed_at` (runtime wall clock)
- Preserve existing digest/signature semantics.

Repo paths:
- `schemas/v1/gate/trace_record.schema.json`
- `core/schema/v1/gate/`
- `core/gate/trace.go`

Acceptance criteria:
- Repeated identical decisions keep same `trace_id` but distinct `event_id`.
- `observed_at` is present for runtime-emitted traces.

### Story 4.2: Prevent default trace overwrite

Tasks:
- Change default trace output naming to include per-event uniqueness while preserving trace digest linkage.
- Keep explicit user-provided `--trace-out` behavior unchanged.

Repo paths:
- `core/gate/trace.go`
- `cmd/gait/mcp.go`
- `cmd/gait/mcp_server.go`

Acceptance criteria:
- Repeated identical calls no longer overwrite each other by default.

### Story 4.3: Clarify deterministic vs operational time fields in docs

Tasks:
- Update artifact contracts and docs to explain:
  - deterministic fields used for cryptographic binding
  - operational timeline fields used for incident reconstruction

Repo paths:
- `docs/contracts/primitive_contract.md`
- `docs/architecture.md`
- `docs/flows.md`

Acceptance criteria:
- Contract docs explicitly prevent confusion about `created_at` vs runtime observation semantics.

---

## Epic 5: Default-Safe Governance Posture (OSS)

Objective: reduce accidental weak production configuration.

### Story 5.1: Add production-readiness doctor profile

Tasks:
- Implement `gait doctor --production-readiness` checks for:
  - profile strictness (`oss-prod`)
  - key mode (`prod`)
  - context completeness
  - service boundary hardening
  - retention settings present

Repo paths:
- `cmd/gait/doctor.go`
- `core/doctor/doctor.go`

Acceptance criteria:
- Command exits non-zero when critical readiness gates fail.
- Output includes actionable remediation commands.

### Story 5.2: Add explicit warnings on weak defaults in risky paths

Tasks:
- Emit structured warnings when `standard` profile falls back to synthetic context for high-risk tools.
- Improve warning text for dev signing mode in service/runtime paths.

Repo paths:
- `core/mcp/proxy.go`
- `cmd/gait/mcp.go`
- `cmd/gait/gate.go`

Acceptance criteria:
- Warning fields are deterministic and machine-parseable in `--json` output.

### Story 5.3: Ship hardened project default templates

Tasks:
- Add hardened `.gait/config.yaml` templates and examples for OSS production deployments.
- Include migration notes from permissive to strict mode.

Repo paths:
- `docs/project_defaults.md`
- `examples/` (new profile template files)

Acceptance criteria:
- Teams can adopt strict defaults with one copy/paste config path.

### Story 5.4: Operational telemetry health indicator

Tasks:
- Add lightweight health signal for telemetry pipeline status (writes ok/dropped/error count).

Repo paths:
- `cmd/gait/main.go`
- `core/scout/operational.go`

Acceptance criteria:
- Operators can detect observability degradation early.

---

## Epic 6: SDK and Adapter Fail-Closed Robustness

Objective: make thin adoption surfaces resilient and hard to misuse.

### Story 6.1: Add subprocess timeout and retry budget to Python client

Tasks:
- Add default timeout to `_run_command` in SDK.
- Add explicit timeout exception category and message.
- Keep retries conservative and deterministic.

Repo paths:
- `sdk/python/gait/client.py`
- `sdk/python/tests/test_client.py`

Acceptance criteria:
- Hung subprocess is surfaced as deterministic SDK error within timeout budget.

### Story 6.2: Tighten adapter service-response contract guidance

Tasks:
- Standardize adapter behavior around:
  - `executed=false` for any non-allow
  - strict handling of service-mode status + payload
- Update all reference integration quickstarts and docs.

Repo paths:
- `sdk/python/gait/adapter.py`
- `examples/integrations/*`
- `docs/integration_checklist.md`

Acceptance criteria:
- Adapter parity suite confirms no reference adapter can fail-open on non-allow verdict.

### Story 6.3: Expand SDK/adapter tests for timeout and failure semantics

Tasks:
- Add tests for timeout, malformed output, service 4xx/5xx handling, and non-allow behavior.

Repo paths:
- `sdk/python/tests/test_client.py`
- `sdk/python/tests/test_adapter.py`
- `scripts/test_adapter_parity.sh`

Acceptance criteria:
- New negative-path tests pass across supported environments.

---

## Epic 7: Retention, Rotation, and Artifact Lifecycle Controls

Objective: prevent unbounded growth and improve operational safety for long-running service deployments.

### Story 7.1: Add retention and rotation flags for service outputs

Tasks:
- Add rotation controls for:
  - trace files
  - session journals/chains
  - export logs
- Add configurable max-age and max-count policies.

Repo paths:
- `cmd/gait/mcp_server.go`
- `core/mcp/exporters.go`
- `core/runpack/session.go`

Acceptance criteria:
- Service mode can run continuously without unbounded artifact growth.

### Story 7.2: Add deterministic prune/report command

Tasks:
- Implement prune/report commands with dry-run and deterministic summary output.

Repo paths:
- `cmd/gait/guard.go` or new lifecycle command surface under `cmd/gait/`
- `core/guard/retention.go`

Acceptance criteria:
- Prune operations produce auditable reports and are reversible by policy (dry-run first).

### Story 7.3: Document recommended retention profiles

Tasks:
- Add OSS guidance for short/medium/long retention tiers by workload profile.

Repo paths:
- `docs/slo/`
- `docs/README.md`

Acceptance criteria:
- Ops teams have clear defaults without enterprise dependencies.

---

## Epic 8: Chaos and Release-Gate Enforcement

Objective: make prime-time failure modes regression-proof.

### Story 8.1: Exporter concurrency corruption gate

Tasks:
- Add chaos test script and Go integration tests that simulate high-concurrency appends.
- Fail if any malformed JSONL line appears.

Repo paths:
- `scripts/test_chaos_exporters.sh`
- `core/mcp/proxy_test.go`
- `core/scout/*_test.go`

Acceptance criteria:
- 0 malformed lines in stress scenario threshold.

### Story 8.2: Service boundary abuse gate

Tasks:
- Add tests for unauthenticated access in non-loopback, forbidden path override attempts, and strict auth enforcement.

Repo paths:
- `scripts/test_chaos_service_boundary.sh`
- `cmd/gait/mcp_server_test.go`

Acceptance criteria:
- Boundary abuse attempts fail closed deterministically.

### Story 8.3: Oversized payload rejection gate

Tasks:
- Add tests that send payloads above configured limit.

Acceptance criteria:
- Response is `413` with stable error payload.

### Story 8.4: Session swarm contention gate

Tasks:
- Add CI/nightly scenario with parallel requests to shared session IDs.
- Define contention/error budget thresholds.

Repo paths:
- `scripts/test_session_soak.sh`
- `internal/integration/contention_test.go`

Acceptance criteria:
- Swarm scenario remains within budget.

### Story 8.5: Long-run latency drift gate

Tasks:
- Add synthetic long-run session append benchmark with drift threshold.

Repo paths:
- `perf/runtime_slo_budgets.json`
- `scripts/check_resource_budgets.py`

Acceptance criteria:
- Long-run append latency growth stays below defined tolerance.

### Story 8.6: Trace uniqueness/evidence retention gate

Tasks:
- Add tests ensuring repeated identical decisions preserve multiple artifacts without overwrite.

Repo paths:
- `core/gate/trace_test.go`
- `cmd/gait/mcp_test.go`

Acceptance criteria:
- Repeated identical decisions produce distinct retained trace files by default.

---

## Epic 9: Documentation, Launch Positioning, and PM Readiness

Objective: make OSS hardening legible to both engineers and product/security buyers.

### Story 9.1: Publish prime-time hardening runbook

Tasks:
- Create `docs/hardening/prime_time_runbook.md` with:
  - baseline deployment profile
  - must-pass checks
  - chaos gate suite
  - escalation flow

Acceptance criteria:
- Runbook is executable by a new team without tribal knowledge.

### Story 9.2: Update messaging for OSS vs enterprise boundary

Tasks:
- Clarify in docs that OSS includes strong runtime hardening while enterprise adds fleet governance.
- Ensure no wording implies hosted dependency for safety-critical OSS features.

Repo paths:
- `README.md`
- `docs/positioning.md`
- `product/ROADMAP.md`

Acceptance criteria:
- Positioning remains consistent across README, docs, and roadmap.

### Story 9.3: Add production rollout checklists for framework maintainers

Tasks:
- Provide framework-agnostic checklist focused on non-technical end-user safety:
  - enforce non-allow block path
  - strict service mode
  - timeout/retention settings

Repo paths:
- `docs/integration_checklist.md`
- `examples/integrations/README.md`

Acceptance criteria:
- Checklist can be used as acceptance criteria for community adapter PRs.

---

## Epic 10: Delivery Sequence and Release Cut Criteria

Objective: ship hardening in safe order with measurable readiness gates.

### Phase sequencing

- Phase A (Security boundary first): Epics 1 + 2
- Phase B (Scale and evidence quality): Epics 3 + 4
- Phase C (Defaults and adoption robustness): Epics 5 + 6 + 7
- Phase D (Hard gates and launch docs): Epics 8 + 9

### v2.2 Release exit criteria

All must be true:

1. All `CHAOS.md` findings F-01 to F-12 are closed or explicitly downgraded with documented rationale.
2. New chaos suites are green in CI/nightly for two consecutive cycles.
3. `gait doctor --production-readiness` exists and is stable.
4. Service boundary hardening defaults are documented and tested.
5. No regressions in deterministic verify/diff/replay/session-chain contract tests.
6. Coverage remains >= 85% for Go core/CLI and Python SDK.
7. `docs/hardening/prime_time_runbook.md` and migration notes are published.

### Commands required before v2.2 tag

- `make fmt`
- `make lint`
- `make test`
- `make test-e2e`
- `make test-hardening-acceptance`
- `bash scripts/test_session_soak.sh`
- `make bench-budgets`
- `bash scripts/test_adapter_parity.sh`

---

## Non-Goals (v2.2)

- No enterprise fleet policy distribution, SSO/RBAC management workflows, or central audit data plane.
- No hosted dashboard as a correctness dependency.
- No language expansion beyond existing OSS-supported adoption surfaces.

---

## Final Success Definition

v2.2 succeeds when Gait OSS can be operated as a hardened local/runtime boundary for long-running agent systems with deterministic evidence, fail-closed semantics, and chaos-tested reliability under realistic concurrency and abuse conditions.
