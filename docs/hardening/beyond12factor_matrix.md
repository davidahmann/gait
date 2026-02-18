# Beyond 12-Factor Alignment Matrix

This matrix maps Beyond 12-Factor operational principles to concrete Gait controls and explicit gaps.

| Principle | Current control(s) | Evidence in repo | Gap / next action |
| --- | --- | --- | --- |
| Dependency declaration and isolation | Reproducible toolchain via `go.mod`, `uv.lock`, pinned CI tooling in workflows | `go.mod`, `sdk/python/uv.lock`, `.github/workflows/release.yml` | Add lockfile freshness policy check in CI (fail when lockfiles drift from manifests). |
| Configuration as environment | Runtime controls via flags/env (`GAIT_ADOPTION_LOG`, `GAIT_OPERATIONAL_LOG`, key env vars, broker env prefixes) | `cmd/gait/*.go`, `README.md` | Add central config reference doc with precedence tables for all commands. |
| Backing services parity | Offline-first core paths; registry remote optional and fail-closed for trust | `core/registry/install.go`, `docs/hardening/contracts.md` | Add explicit "offline test mode required" assertion in CI for integration suites. |
| Build/release/run separation | Distinct lint/test/release workflows and release integrity verification | `.github/workflows/ci.yml`, `.github/workflows/release.yml` | Add release dry-run workflow that validates end-to-end publishing logic without tagging. |
| Stateless processes with durable state stores | CLI processes are short-lived; state persisted via deterministic files | `core/gate/rate_limit.go`, `core/fsx/fsx.go` | Add configurable state-root isolation per environment profile (dev/stage/prod). |
| Logs as event streams | Structured, opt-in local JSONL adoption and operational events | `core/scout/adoption.go`, `core/scout/operational.go` | Add schema version migration notes for event stream consumers. |
| Admin processes as first-class | `doctor`, `policy test`, `regress`, `incident pack` are explicit operational commands | `cmd/gait/doctor.go`, `cmd/gait/policy.go`, `cmd/gait/regress.go`, `cmd/gait/incident.go` | Add documented incident triage playbook linking these commands in a single flow. |
| Disposability and graceful shutdown | Time-bounded command broker execution, lock timeout handling, deterministic retries | `core/credential/providers.go`, `core/gate/rate_limit.go`, `core/registry/install.go` | Add benchmarked timeout budgets by command and fail-fast SLO assertions. |
| Port binding / API contract | CLI and artifact schemas are the primary contract (`--json`, schema validation, exit code stability) | `core/schema/`, `github.com/Clyra-AI/proof/schema`, `cmd/gait/error_output.go` | Add contract changelog file that records every CLI/schema surface change. |
| Concurrency and scale-out | Concurrency hardening tests for shared rate-limit state and deterministic lock behavior | `internal/integration/concurrency_test.go`, `core/gate/rate_limit.go` | Add stress profile for high-concurrency gate evaluations in nightly workflow. |
| Telemetry and tracing | Correlation IDs and operational events align command outputs with traces | `cmd/gait/correlation.go`, `core/schema/v1/scout/types.go`, `schemas/v1/scout/operational_event.schema.json` | Add optional OTEL mapping doc for teams exporting local events externally. |
| Security by default | Fail-closed high-risk behavior, signature verification, credential broker constraints | `github.com/Clyra-AI/proof/signing`, `core/credential/providers.go`, `cmd/gait/gate.go` | Add policy lint rule for unsafe flag usage in CI examples and docs. |

## Current posture

- Implemented strongly in v1: deterministic artifacts/contracts, fail-closed controls, operational command parity, and offline-first behavior.
- Priority gaps to close next: lockfile drift enforcement, stress-profile automation, and explicit configuration precedence documentation.
