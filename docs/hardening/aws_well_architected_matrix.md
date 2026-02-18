# AWS Well-Architected Mapping

This matrix maps Gait controls to the six AWS Well-Architected pillars and tracks concrete gaps.

| Pillar | Implemented controls | Evidence in repo | Open gap / next action |
| --- | --- | --- | --- |
| Operational Excellence | Deterministic CLI outputs, doctor diagnostics, hardening/nightly workflows, pre-push enforcement | `cmd/gait/doctor.go`, `.github/workflows/ci.yml`, `.github/workflows/hardening-nightly.yml`, `.githooks/pre-push` | Add runbook-level incident response SOP with owner and escalation timelines. |
| Security | Signed artifacts/traces, credential broker hardening, allowlist controls, vulnerability scanning | `github.com/Clyra-AI/proof/signing`, `core/credential/providers.go`, `Makefile` (`gosec`, `govulncheck`, `bandit`) | Add threat model doc for broker and policy bypass scenarios with mitigations. |
| Reliability | Atomic writes, contention handling, retry/backoff for remote registry, deterministic error classification | `core/fsx/fsx.go`, `core/gate/rate_limit.go`, `core/registry/install.go`, `cmd/gait/error_output.go` | Add failure-injection CI job that runs targeted chaos tests on every merge. |
| Performance Efficiency | Deterministic zip/canonicalization routines and benchmark tooling | `core/zipx/zipx.go`, `github.com/Clyra-AI/proof/canon`, `perf/`, `Makefile` (`bench`) | Add automated benchmark regression gate on scheduled CI with alert threshold. |
| Cost Optimization | Offline-first defaults, no hosted dependency requirement for correctness, local artifacts for audits | `README.md`, `docs/hardening/contracts.md`, `core/*` | Add guidance to cap artifact retention by default to control local storage growth. |
| Sustainability | Minimal runtime footprint (single static binary core), optional telemetry/export layers | `cmd/gait`, `core/mcp/exporters.go`, `README.md` | Add baseline energy/runtime profile for heavy commands to guide optimization priorities. |

## Pillar summary

- Strongest current pillars: Security, Reliability, and Operational Excellence.
- Next minimum actions: automate perf regression checks and add explicit threat model plus incident SOP documentation.
