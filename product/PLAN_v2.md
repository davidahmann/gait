# PLAN v2: Policy Authoring Ergonomics + Deterministic Validation

Date: 2026-02-11
Source of truth: `product/PLAN_v1.md`, `product/PLAN_v1.9.md`, `README.md`, `docs/policy_rollout.md`, current `main` codebase
Scope: OSS CLI and docs improvements only (no hosted control plane)

This plan is written to be executed top-to-bottom with minimal interpretation. Every story includes concrete repo paths, commands, and acceptance criteria.

---

## Global Decisions (Locked for v2)

- Keep Go authoritative for all policy parsing, normalization, evaluation, and output contracts.
- Keep policy behavior deterministic and offline-first.
- Treat policy authoring UX as a product surface (CLI + fixtures + docs), not just internal plumbing.
- Preserve stable exit code semantics and backward-compatible JSON output shape; new fields are additive only.
- Keep fail-closed posture for invalid or non-evaluable policy/input paths in production workflows.
- Keep CI merge path fast: Windows lint stays in nightly hardening lane, not per-push CI lane.

---

## v2 Objective

Reduce policy-adoption friction and authoring error rate without adding a hosted UI by shipping:

1. First-class policy authoring commands (`validate`, `fmt`) with strict checks.
2. Strict unknown-field and syntax diagnostics for policy YAML.
3. Better policy-test explainability (`matched_rule`) while preserving deterministic outputs.
4. A machine-readable policy schema artifact and IDE wiring guidance.
5. CI/UAT coverage updates so policy authoring behavior is enforced pre-merge.

---

## Current Baseline (Observed)

Already in place:

- `gait policy init` with baseline templates.
- `gait policy test` deterministic verdict and stable exit code mapping.
- Policy normalization and evaluation logic in `core/gate/policy.go`.
- Rich policy examples under `examples/policy/*` and fixtures under `examples/policy-test/*`.
- Dedicated CI policy compliance workflow step (`scripts/policy_compliance_ci.sh`).
- Main CI no longer runs Windows lint; nightly workflow exists (`.github/workflows/windows-lint-nightly.yml`).

Current gaps:

- No standalone `gait policy validate` command for syntax/schema-only checks.
- No deterministic formatter command for policy files.
- Policy parser can be improved with strict unknown-field rejection ergonomics.
- No published policy JSON schema artifact for editor validation.
- Policy test output lacks explicit matched-rule context.

---

## v2 Exit Criteria

v2 is complete only when all are true:

- `gait policy validate <policy.yaml>` exists, with stable JSON and text output.
- `gait policy fmt <policy.yaml>` exists, supports deterministic formatting and optional write-back.
- Policy YAML parsing rejects unknown fields and invalid structures with actionable errors.
- `gait policy test --json` includes additive explainability field(s) (`matched_rule`) without breaking existing consumers.
- A policy JSON schema artifact is added under `schemas/v1/gate/` and covered in schema tests.
- Documentation covers:
  - authoring flow (`init -> validate -> fmt -> test`)
  - IDE schema wiring
  - fixture-based CI expectations
- Main CI remains fast path (no Windows lint job on push/PR).
- Local validation passes for changed code paths and UAT plan remains green.

---

## Epic V2.0: Policy Command Surface

Objective: make policy authoring a first-class workflow in the CLI.

### Story V2.0.1: Add `gait policy validate`

Tasks:

- Add new policy subcommand `validate`.
- Inputs: `gait policy validate <policy.yaml> [--json]`.
- Behavior:
  - load and strictly parse policy YAML
  - run full normalization/semantic validation
  - return deterministic metadata (schema id/version, digest, rule count, default verdict)
- Exit behavior:
  - `0` for valid policy
  - `6` for parse/validation/input failures

Repo paths:

- `cmd/gait/policy.go`
- `cmd/gait/verify.go` (global usage)
- `cmd/gait/main_test.go`

Acceptance criteria:

- Valid baseline policies return `ok=true` and exit `0`.
- Invalid/unknown-field policies return `ok=false` and exit `6`.

### Story V2.0.2: Add `gait policy fmt`

Tasks:

- Add new policy subcommand `fmt`.
- Inputs: `gait policy fmt <policy.yaml> [--write] [--json]`.
- Behavior:
  - parse policy strictly
  - normalize policy via canonical Go policy model
  - emit deterministic YAML formatting
  - with `--write`, atomically overwrite file only when content changes
- Output includes changed/no-op status and path.

Repo paths:

- `cmd/gait/policy.go`
- `cmd/gait/main_test.go`

Acceptance criteria:

- Running fmt twice is idempotent (second run reports no change).
- Non-write mode emits formatted YAML to stdout deterministically.

---

## Epic V2.1: Strict Parsing + Explainability

Objective: reduce configuration mistakes and improve debugging speed.

### Story V2.1.1: Strict YAML parsing in Gate policy parser

Tasks:

- Update `ParsePolicyYAML` to use strict decoding options.
- Reject unknown fields and malformed types early.
- Keep normalized defaults/evaluation behavior unchanged for valid policies.

Repo paths:

- `core/gate/policy.go`
- `core/gate/policy_test.go`

Acceptance criteria:

- Unknown field typo (for example `default_verdit`) fails parse.
- Existing valid policy fixtures continue to parse and evaluate unchanged.

### Story V2.1.2: Add matched-rule explainability to policy test output

Tasks:

- Wire policy testing to detailed evaluation path.
- Add additive `matched_rule` field in policy test result schema and CLI JSON output.
- Keep existing fields and verdict semantics unchanged.

Repo paths:

- `core/policytest/run.go`
- `core/schema/v1/policytest/types.go`
- `schemas/v1/policytest/policy_test_result.schema.json`
- `cmd/gait/policy.go`
- schema fixtures/tests as needed

Acceptance criteria:

- `gait policy test --json` includes `matched_rule` when a rule matched.
- Existing consumers reading old fields continue to work.

---

## Epic V2.2: Schema + Docs + IDE Guidance

Objective: remove adoption friction in editors and runbooks.

### Story V2.2.1: Add gate policy JSON schema artifact

Tasks:

- Add `schemas/v1/gate/policy.schema.json` representing policy file structure and enums.
- Add schema validation fixtures/tests.

Repo paths:

- `schemas/v1/gate/policy.schema.json`
- `core/schema/testdata/*` (policy schema fixtures)
- `core/schema/validate/validate_test.go`

Acceptance criteria:

- Schema validates valid fixture and rejects invalid fixture in tests.
- Schema versioning follows existing v1 schema conventions.

### Story V2.2.2: Surgical docs updates

Tasks:

- Update README policy section with recommended authoring loop.
- Update policy rollout guide with `validate` and `fmt` stages.
- Update integration checklist with pre-merge policy validation steps.
- Update docs map and docs-site navigation to expose policy authoring guidance.

Repo paths:

- `README.md`
- `docs/policy_rollout.md`
- `docs/integration_checklist.md`
- `docs/README.md`
- `docs-site/src/lib/navigation.ts`

Acceptance criteria:

- New commands are discoverable in README and docs ladder.
- No docs contradictions with CLI behavior.

---

## Epic V2.3: CI/UAT Coverage + Throughput Guard

Objective: keep confidence high without slowing merge velocity.

### Story V2.3.1: Expand policy compliance checks for authoring workflow

Tasks:

- Extend policy compliance script to include:
  - policy validate checks for baseline/example packs
  - fmt idempotence check for fixture policy
- Keep output artifact summary and deterministic pass/fail behavior.

Repo paths:

- `scripts/policy_compliance_ci.sh`
- `.github/workflows/ci.yml` (only if invocation changes)

Acceptance criteria:

- CI fails when policy authoring commands regress.
- Compliance summary reflects new checks.

### Story V2.3.2: Keep Windows lint off merge path

Tasks:

- Verify/retain nightly-only Windows lint workflow.
- Confirm main CI matrix excludes Windows for lint lane.
- Document this throughput decision in plan and leave implementation stable.

Repo paths:

- `.github/workflows/ci.yml`
- `.github/workflows/windows-lint-nightly.yml`

Acceptance criteria:

- Push/PR CI has no Windows lint job.
- Nightly Windows lint remains runnable manually and by schedule.

---

## Validation Plan (Local)

Run in order:

1. `gofmt -w .`
2. `go test ./core/gate ./core/policytest ./core/schema/validate ./cmd/gait`
3. `go test ./...`
4. `bash scripts/policy_compliance_ci.sh`
5. `make test`
6. `make test-acceptance`
7. `make test-v1-6-acceptance`
8. `make test-v1-7-acceptance`
9. `make test-v1-8-acceptance`
10. `bash scripts/test_uat_local.sh --skip-brew`

If any command fails, fix and rerun from the failing step.

---

## Release/Delivery Checklist

- [ ] `PLAN_v2.md` committed with implementation.
- [ ] Code + tests + docs merged on `main`.
- [ ] CI workflow and docs-site workflow both green.
- [ ] UAT local run green after CI green.
- [ ] Changelog update only if user requests version/tag release in this cycle.

---

## Execution Order (Strict)

1. Add plan file.
2. Implement CLI and parser changes.
3. Add schema and schema tests.
4. Update docs and navigation.
5. Run validation + UAT.
6. Commit and push.
7. Monitor CI/docs to green; fix forward if red.

---

## OSS Compounding Scope (Post-v2 Policy Ergonomics)

This section captures the next OSS compounding loop and explicitly separates it from enterprise control-plane features.

### OSS Now (Implement in CLI + local docs/tests)

1. Artifact specification hardening:
   - publish a formal artifact protocol document for runpack/trace/regress linkage and hash references
2. Historical policy simulation:
   - compare candidate policy vs baseline policy across deterministic intent fixtures
   - emit verdict deltas and rollout-stage recommendation (`observe`, `require_approval`, `enforce`)
3. Regress corpus accumulation:
   - optional append-only local history output for repeated regress runs
4. Key lifecycle basics:
   - local key init/rotate/verify commands for ed25519 signing workflow
5. Adapter conformance baseline:
   - keep adapter-parity suite and contract docs aligned as the de-facto conformance kit
6. Local corpus query primitives:
   - maintain CLI-first query surface (`run inspect`, `scout signal`) as local read path

### ENT Later (Tracked in `PLAN_ENT.md`)

- centralized artifact indexing and cross-team search
- org policy inheritance/composition
- fleet-scale simulation and rollout impact analysis
- enterprise KMS/HSM integrations and trust hierarchy
- compliance control mapping/reporting
- hosted workflow integrations (ticketing/SIEM/alerting)
