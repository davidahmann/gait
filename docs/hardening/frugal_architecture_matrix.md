# Frugal Architecture Mapping (Vogels)

This matrix maps Gait design choices to frugal architecture principles and captures near-term decisions.

| Frugal principle | Current Gait alignment | Evidence in repo | Keep / do next / avoid |
| --- | --- | --- | --- |
| Make cost a non-functional requirement | Offline-first CLI avoids always-on service costs for core correctness | `README.md`, `core/*` | Keep: local-first core. Do next: define artifact retention defaults by profile. Avoid: mandatory hosted control plane in v1.x. |
| Use managed simplicity where possible | Leans on standard crypto/libs and GitHub Actions instead of bespoke infra | `github.com/Clyra-AI/proof/signing`, `.github/workflows/*`, `Makefile` | Keep: boring primitives. Do next: pin additional workflow dependencies for reproducibility. Avoid: custom crypto or custom scheduler layers. |
| Design for variable demand and constraints | Deterministic local artifacts support disconnected, constrained environments | `core/runpack/*`, `core/gate/*`, `core/registry/*` | Keep: deterministic/offline guarantees. Do next: stress-test large runpacks and concurrent operations in nightly suites. Avoid: network-required execution paths. |
| Architect for resilience before scale | Atomic writes, lock contention handling, retry classification, fail-closed policy decisions | `core/fsx/fsx.go`, `core/gate/rate_limit.go`, `core/registry/install.go` | Keep: fail-closed safety. Do next: add hardening acceptance gate in CI. Avoid: permissive fallback paths that bypass verification. |
| Optimize globally, not locally | Unified contracts across CLI, schemas, and CI quality gates | `core/schema/`, `cmd/gait/error_output.go`, `.github/workflows/ci.yml` | Keep: contract-first development. Do next: publish contract change log and release checklist sign-offs. Avoid: one-off command behaviors that diverge from global envelope rules. |

## Two-milestone operating decisions

### Keep

- Keep the Go core as the authoritative execution and contract surface.
- Keep Python wrapper thin and policy-logic-free.
- Keep offline determinism and fail-closed defaults as non-negotiable.

### Do next

- Add hardening acceptance workflow and script as release-blocking gate (H12).
- Add benchmark/stress automation for large inputs and concurrency budgets (H11).
- Add explicit threat-model and incident SOP docs to close operational governance gaps.

### Not doing (for now)

- No hosted dashboard/control-plane dependency in v1.x core flow.
- No expansion of network-bound features that weaken offline operation.
- No broad adapter proliferation without deterministic test coverage and policy conformance.
