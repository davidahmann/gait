# PLAN v2.3: Gait Adoption Proof and Distribution Hardening (Zero-Ambiguity Execution Plan)

Date: 2026-02-11
Source of truth: `product/PLAN_v1.md`, `product/PLAN_v2.2.md`, `gait-out/gtm_1/CHAOS.md`, `docs/integration_checklist.md`, `docs/ci_regress_kit.md`, `docs/ecosystem/awesome.md`
Scope: v2.3 only (convert reliability into adoption proof, lock highest-fit integration lane, publish Intent+Receipt conformance path, ship skill workflows as thin wrappers, harden GitHub Actions regression distribution)

This plan is written to be executed top-to-bottom with minimal interpretation. Each story includes concrete repo paths, commands, and acceptance criteria.

---

## Global Decisions (Locked for v2.3)

- v2.3 is an adoption-conversion release, not a new core-product-surface release.
- The product wedge remains unchanged:
  - Runpack
  - Regress
  - Gate
- v2.3 does not introduce hosted services or control-plane dependencies.
- Lane strategy is locked to highest-fit execution surface:
  - coding agents on developer machines
  - CI workflows (GitHub Actions first)
- Data-driven lane selection is required before adding any new official adapter lane.
- Skill workflows remain thin wrappers over CLI commands; no policy logic or signing logic in skills.
- Existing schema and artifact contracts remain additive-only within `v1.x`.
- Adoption proof must be measurable with local/offline-safe artifacts and deterministic test gates.

---

## Scope Boundaries (In / Out)

### In Scope

- Lock one blessed integration lane and prove it with measurable activation outcomes.
- Tighten wrapper docs and examples to true 15-minute setup from clean checkout.
- Publish conformance-style "Intent + Receipt" verification workflow using existing schemas and tests.
- Ship and harden the three official skill workflows as wrapper-only distribution assets.
- Strengthen GitHub Actions regression path to be a reusable, deterministic adoption baseline.

### Out of Scope

- New policy language features.
- New runpack schema versions.
- New hosted dashboard/control-plane functionality.
- Large adapter expansion beyond the blessed lane.
- Non-GitHub CI feature parity beyond compatibility snippets.

---

## Success Metrics and Stage Gates (Locked)

### Activation Metrics (blessed lane)

- `M1`: median install-to-`gait demo` success <= 5 minutes.
- `M2`: median install-to-`gait regress run` success <= 15 minutes.
- `M3`: wrapper quickstart completion rate >= 90% on clean local environment.
- `M4`: deterministic CI regression template pass rate >= 95% on representative repos.

### Conformance Metrics

- `C1`: Intent schema conformance checks pass 100% on required fixtures.
- `C2`: Receipt/ticket-footer verification path passes 100% in contract CI job.
- `C3`: additive-field compatibility checks remain green in enterprise consumer contract.

### Distribution Metrics

- `D1`: official skill install success >= 95% (`codex`, `claude`) in supply-chain test lanes.
- `D2`: skill workflow end-to-end tests pass on CI with deterministic outputs.
- `D3`: adoption-regress template reuse path validates in workflow-call mode without edits.

### Release Gate

v2.3 is releasable only when all metrics `M1..M4`, `C1..C3`, and `D1..D3` are green in CI/nightly checks and documented in release notes.

---

## Repository Touch Map (v2.3 planned)

```
product/
  PLAN_v2.3.md

docs/
  integration_checklist.md
  ci_regress_kit.md
  contracts/intent_receipt_conformance.md        # new
  ecosystem/awesome.md
  launch/kpi_scorecard.md

examples/
  integrations/template/README.md
  integrations/template/quickstart.py
  integrations/openai_agents/README.md           # blessed lane docs focus
  integrations/openai_agents/quickstart.py

.agents/skills/
  gait-capture-runpack/SKILL.md
  gait-incident-to-regression/SKILL.md
  gait-policy-test-rollout/SKILL.md

scripts/
  quickstart.sh
  test_adoption_smoke.sh
  test_skill_supply_chain.sh
  test_intent_receipt_conformance.sh             # new
  check_integration_lane_scorecard.py            # new

.github/workflows/
  adoption-regress-template.yml
  adoption-nightly.yml
  ci.yml

cmd/gait/
  run_receipt.go
  gate.go
  demo.go
  main.go                                        # telemetry milestones where needed

core/
  scout/                                         # adoption event scoring inputs
  schema/validate/                               # conformance fixture checks
```

---

## Epic 0: Data-Driven Lane Selection and Governance

Objective: choose and lock the highest-fit integration lane with explicit evidence, not opinion.

### Story 0.1: Define lane scorecard and decision rule

Tasks:
- Add a lane scorecard spec documenting candidates and weights:
  - coding-agent local wrapper lane
  - CI workflow lane
  - IT workflow lane (comparison only)
- Define weighted metrics:
  - setup time
  - failure rate
  - determinism pass rate
  - policy-enforcement correctness
  - distribution reach potential
- Add scoring script:
  - `scripts/check_integration_lane_scorecard.py`

Repo paths:
- `docs/integration_checklist.md`
- `docs/launch/kpi_scorecard.md`
- `scripts/check_integration_lane_scorecard.py`

Commands:
- `python3 scripts/check_integration_lane_scorecard.py --input gait-out/adoption_metrics.json --out gait-out/integration_lane_scorecard.json`

Acceptance criteria:
- Scorecard output is deterministic for identical inputs.
- Decision rule explicitly identifies the selected lane and confidence.
- Result artifact is committed in CI output for auditability.

### Story 0.2: Lock blessed lane and freeze expansion policy

Tasks:
- Update integration docs to mark one blessed lane:
  - coding-agent wrapper + GitHub Actions CI
- Add explicit policy:
  - no new official integration lane in v2.3 unless scorecard threshold is met.
- Add PR checklist item requiring lane-scorecard reference for any new adapter proposal.

Repo paths:
- `docs/integration_checklist.md`
- `docs/ecosystem/contribute.md`
- `.github/pull_request_template.md`

Acceptance criteria:
- Docs reference a single default lane.
- Adapter expansion policy is explicit and enforceable in review templates.

---

## Epic 1: Blessed Lane Execution Path (Coding-Agent + CI)

Objective: make one integration lane fast, deterministic, and hard to misuse.

### Story 1.1: Canonical wrapper path reduced to 15-minute flow

Tasks:
- Narrow integration checklist to one primary path first:
  1. emit `IntentRequest`
  2. evaluate Gate
  3. enforce non-`allow` fail-closed
  4. persist trace
  5. emit runpack
  6. convert to regress fixture
- Ensure docs emphasize copy/paste sequence with exact expected outputs.
- Keep advanced lanes below fold.

Repo paths:
- `docs/integration_checklist.md`
- `examples/integrations/template/README.md`
- `examples/integrations/template/quickstart.py`
- `sdk/python/gait/adapter.py`

Commands:
- `bash scripts/quickstart.sh`
- `bash scripts/test_adapter_parity.sh`
- `python3 examples/integrations/template/quickstart.py --scenario allow`
- `python3 examples/integrations/template/quickstart.py --scenario block`

Acceptance criteria:
- New user can complete canonical wrapper flow in <= 15 minutes using only listed commands.
- Block and approval paths show `executed=false` behavior in examples.

### Story 1.2: OpenAI Agents + GitHub Actions as documented blessed default

Tasks:
- Promote one official adapter as top-of-funnel reference in docs:
  - `examples/integrations/openai_agents/`
- Keep other adapters as parity references, not primary docs path.
- Add explicit mapping from local wrapper output to CI regress enforcement.

Repo paths:
- `examples/integrations/openai_agents/README.md`
- `README.md`
- `docs/ci_regress_kit.md`

Commands:
- `python3 examples/integrations/openai_agents/quickstart.py --scenario allow`
- `python3 examples/integrations/openai_agents/quickstart.py --scenario block`

Acceptance criteria:
- Blessed adapter docs and CI docs reference each other directly.
- Local run -> CI regress path is reproducible without undocumented steps.

### Story 1.3: Activation timing instrumentation for blessed lane

Tasks:
- Add milestone tagging for:
  - first demo
  - first verify
  - first regress init
  - first regress run
- Ensure adoption events can produce median timing outputs for `M1` and `M2`.
- Add report command usage examples to docs.

Repo paths:
- `cmd/gait/main.go`
- `core/scout/adoption.go`
- `docs/integration_checklist.md`

Commands:
- `GAIT_ADOPTION_LOG=./gait-out/adoption.jsonl ./gait demo`
- `GAIT_ADOPTION_LOG=./gait-out/adoption.jsonl ./gait verify run_demo --json`
- `GAIT_ADOPTION_LOG=./gait-out/adoption.jsonl ./gait regress init --from run_demo --json`
- `GAIT_ADOPTION_LOG=./gait-out/adoption.jsonl ./gait regress run --json`
- `./gait doctor adoption --from ./gait-out/adoption.jsonl --json`

Acceptance criteria:
- Adoption report returns activation milestone timing and blockers deterministically.
- Timing fields are sufficient to calculate `M1` and `M2`.

---

## Epic 2: Intent + Receipt Conformance Path (Contract-First)

Objective: publish and enforce a clear conformance workflow that proves boundary intent and receipt continuity using existing schema/tests.

### Story 2.1: Publish Intent + Receipt conformance contract doc

Tasks:
- Add `docs/contracts/intent_receipt_conformance.md` describing:
  - IntentRequest schema conformance
  - Gate evaluation digest continuity (`intent_digest`, `policy_digest`)
  - Runpack receipt continuity (`refs.receipts`)
  - Ticket footer extraction and verification contract (`gait run receipt`)
- Include minimal normative command sequence and expected fields.

Repo paths:
- `docs/contracts/intent_receipt_conformance.md`
- `docs/contracts/primitive_contract.md` (cross-link)
- `README.md` (link in integration section)

Acceptance criteria:
- Contract is discoverable from README and integration checklist.
- Document uses existing schema IDs and command contracts only.

### Story 2.2: Add deterministic conformance script and CI gate

Tasks:
- Add script:
  - `scripts/test_intent_receipt_conformance.sh`
- Script validates:
  - `schemas/v1/gate/intent_request.schema.json` required fields unchanged
  - Gate output includes required digest fields
  - `gait run receipt --from <run>` returns contract-valid ticket footer
  - runpack `refs.json` contains deterministic receipt structure
- Wire script into contract CI lane.

Repo paths:
- `scripts/test_intent_receipt_conformance.sh`
- `scripts/test_contracts.sh`
- `.github/workflows/ci.yml`

Commands:
- `bash scripts/test_intent_receipt_conformance.sh ./gait`
- `make test-contracts`

Acceptance criteria:
- Conformance script fails on schema/field drift.
- Contract job remains offline-safe and deterministic.

### Story 2.3: Enterprise additive compatibility coverage for conformance fields

Tasks:
- Extend enterprise consumer contract checks to include:
  - intent/receipt projection digest
  - additive-field tolerance for conformance-relevant fields
- Ensure no breaking change risk for downstream parsers.

Repo paths:
- `scripts/test_ent_consumer_contract.sh`
- `docs/contracts/artifact_graph.md`

Acceptance criteria:
- Enterprise contract test passes with additive unknown fields.
- No conformance field dependency on non-stable output ordering.

---

## Epic 3: 15-Minute Onboarding Proof (Docs + Scripts)

Objective: make "first real runpack from real workflow" achievable in a strict 15-minute envelope.

### Story 3.1: Tighten quickstart to activation envelope

Tasks:
- Update `scripts/quickstart.sh` to produce timing checkpoints and explicit next-step outputs.
- Ensure quickstart covers:
  - demo
  - verify
  - regress init
  - regress run (optional fast lane with clear flag)
- Keep failure messages action-oriented and deterministic.

Repo paths:
- `scripts/quickstart.sh`
- `README.md`
- `docs/integration_checklist.md`

Commands:
- `bash scripts/quickstart.sh`
- `GAIT_OUT_DIR=./gait-out/quickstart bash scripts/quickstart.sh`

Acceptance criteria:
- Quickstart path produces runpack, verify proof, and regress guidance in one run.
- Median quickstart duration on supported laptop environment <= 15 minutes end-to-end.

### Story 3.2: Wrapper docs rewritten for strict minimal path

Tasks:
- Update wrapper docs to remove optional branches from primary sequence.
- Ensure one minimal code snippet is canonical and reused across docs.
- Add a "15-minute checklist" section with explicit stop/go outputs.

Repo paths:
- `docs/integration_checklist.md`
- `examples/python/README.md`
- `sdk/python/examples/openai_style_tool_decorator.py`
- `sdk/python/examples/langchain_style_tool_decorator.py`

Acceptance criteria:
- Primary wrapper docs can be executed without reading secondary pages.
- Copy/paste commands produce the documented output fields.

### Story 3.3: Adoption smoke upgraded to assert 15-minute-first-win artifacts

Tasks:
- Extend `scripts/test_adoption_smoke.sh` to assert:
  - runpack exists
  - verify passes
  - regress init/run outputs exist
  - expected trace paths exist for adapter/sidecar smoke
- Add clear fail reasons for any missing adoption proof artifact.

Repo paths:
- `scripts/test_adoption_smoke.sh`
- `Makefile` (`test-adoption` target remains canonical)

Acceptance criteria:
- Adoption smoke fails on missing first-win artifacts.
- Output clearly identifies which step broke conversion path.

---

## Epic 4: Skill Workflow Shipping and Hardening (Thin Wrappers Only)

Objective: ship the three skill workflows as reliable distribution adapters, with strict CLI-only logic.

### Story 4.1: Normalize all three official skills to thin-wrapper contract

Tasks:
- Review and tighten skill bodies:
  - `.agents/skills/gait-capture-runpack/SKILL.md`
  - `.agents/skills/gait-incident-to-regression/SKILL.md`
  - `.agents/skills/gait-policy-test-rollout/SKILL.md`
- Require `--json` command usage and explicit parse fields.
- Remove any implied business logic beyond CLI orchestration.

Acceptance criteria:
- Skill files contain only wrapper orchestration, safety rules, and deterministic reporting requirements.
- No policy parsing or evaluator logic described outside Go CLI invocation.

### Story 4.2: Skill packaging/install path hardening

Tasks:
- Ensure `scripts/install_repo_skills.sh` remains deterministic and provider-safe.
- Add idempotence and overwrite behavior tests where needed.
- Verify codex/claude install paths with deterministic folder outcomes.

Repo paths:
- `scripts/install_repo_skills.sh`
- `scripts/test_skill_supply_chain.sh`
- `docs/ecosystem/awesome.md`
- `docs/ecosystem/contribute.md`

Commands:
- `bash scripts/install_repo_skills.sh --provider codex`
- `bash scripts/install_repo_skills.sh --provider claude`
- `bash scripts/test_skill_supply_chain.sh`

Acceptance criteria:
- Re-running install script yields the same installed file set.
- Skill supply-chain checks remain green in CI.

### Story 4.3: Data-driven skill workflow fitness reporting

Tasks:
- Add skill workflow score section in adoption report:
  - success rate per skill workflow
  - median runtime
  - most common failure code
- Use local event/log artifacts only; no network dependency.

Repo paths:
- `core/scout/adoption.go`
- `cmd/gait/doctor.go`
- `docs/launch/kpi_scorecard.md`

Acceptance criteria:
- Skill workflow metrics are available in deterministic JSON output.
- Metrics can be consumed by lane scorecard script.

---

## Epic 5: GitHub Actions Regress Path Hardening

Objective: make GitHub Actions the strongest and easiest CI adoption path for v2.3.

### Story 5.1: Reusable workflow contract for regress gate

Tasks:
- Harden `.github/workflows/adoption-regress-template.yml` as reusable workflow:
  - `workflow_call` inputs/outputs documented
  - deterministic fixture restoration behavior
  - stable artifact upload paths
- Ensure workflow can be consumed by downstream repos without edits beyond policy/fixture paths.

Repo paths:
- `.github/workflows/adoption-regress-template.yml`
- `docs/ci_regress_kit.md`

Acceptance criteria:
- Template runs in both `workflow_dispatch` and `workflow_call` modes.
- Exit-code semantics remain stable (`0`, `5`, error passthrough).

### Story 5.2: Stronger regress summary and failure diagnostics

Tasks:
- Add workflow step that surfaces key JSON fields in job summary:
  - status
  - top failure reason
  - next command
  - artifact paths
- Keep machine-readable artifacts unchanged.

Repo paths:
- `.github/workflows/adoption-regress-template.yml`
- `docs/ci_regress_kit.md`

Acceptance criteria:
- Regression failure triage can start from workflow summary alone.
- No change to deterministic JSON artifact contract.

### Story 5.3: Path-filtered CI enforcement for adoption-critical changes

Tasks:
- Add/adjust CI path filters so adoption regress workflow runs when touching:
  - `cmd/gait/**`
  - `core/runpack/**`
  - `core/regress/**`
  - `core/gate/**`
  - `schemas/**`
  - `docs/integration_checklist.md`
  - `docs/ci_regress_kit.md`
  - `.agents/skills/**`
- Avoid unnecessary runs for unrelated docs-only changes.

Repo paths:
- `.github/workflows/ci.yml`
- `.github/workflows/adoption-nightly.yml`

Acceptance criteria:
- Adoption-critical changes always exercise regress template path.
- CI runtime remains bounded for unrelated changes.

### Story 5.4: Golden downstream-consumer test for template reuse

Tasks:
- Add a lightweight fixture repo simulation in test scripts that:
  - copies template
  - runs regress path with fixture restore
  - asserts stable outputs/artifacts
- Keep simulation offline-safe.

Repo paths:
- `scripts/test_ecosystem_release_automation.sh` (extend) or `scripts/test_ci_regress_template.sh` (new)
- `docs/ci_regress_kit.md`

Acceptance criteria:
- Template reuse is tested automatically; breakage is caught pre-release.

---

## Epic 6: Adoption Proof Packaging (Reliability -> Pull)

Objective: convert technical reliability into credible external proof artifacts.

### Story 6.1: Standardized adoption proof bundle format

Tasks:
- Define a deterministic adoption proof bundle containing:
  - sample runpack
  - verify output
  - regress result
  - CI JUnit
  - conformance report
  - lane scorecard output
- Document bundle generation command sequence.

Repo paths:
- `docs/launch/README.md`
- `docs/launch/kpi_scorecard.md`
- `scripts/demo_90s.sh`

Acceptance criteria:
- Bundle can be generated in one local run with deterministic file paths.
- Bundle contents map directly to v2.3 success metrics.

### Story 6.2: Design-partner incident conversion runbook (OSS-safe)

Tasks:
- Add runbook for converting one real incident into:
  - runpack
  - regression fixture
  - gated policy check
  - receipt/ticket-footer artifact
- Keep runbook free of hosted assumptions.

Repo paths:
- `docs/evidence_templates.md`
- `docs/ci_regress_kit.md`
- `examples/scenarios/README.md`

Acceptance criteria:
- Runbook is executable with current CLI and example assets.
- Outputs are compatible with existing schema and verify commands.

---

## Epic 7: Test and Acceptance Gates for v2.3

Objective: prevent adoption-path regressions from shipping.

### Story 7.1: Add dedicated v2.3 acceptance script

Tasks:
- Add `scripts/test_v2_3_acceptance.sh` covering:
  - blessed wrapper flow
  - Intent+Receipt conformance script
  - skill workflow smoke
  - reusable GitHub regress template assumptions
- Add make target `test-v2-3-acceptance`.

Repo paths:
- `scripts/test_v2_3_acceptance.sh`
- `Makefile`

Acceptance criteria:
- One script validates all v2.3 scope claims offline.
- Script failure messages map directly to story-level gaps.

### Story 7.2: CI wiring for v2.3 release gate

Tasks:
- Wire v2.3 acceptance lane into CI/release gating.
- Ensure lane runs on PRs touching adoption/distribution surface.

Repo paths:
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`

Acceptance criteria:
- v2.3 acceptance lane blocks merge/release when failing.
- Gate is stable across Linux/macOS/Windows where applicable.

### Story 7.3: Metrics snapshot export for release notes

Tasks:
- Emit machine-readable snapshot for `M*`, `C*`, `D*` metrics.
- Add release-note checklist section requiring this snapshot.

Repo paths:
- `docs/launch/github_release_template.md`
- `scripts/render_ecosystem_release_notes.py` (extend if needed)

Acceptance criteria:
- Release artifacts include v2.3 score snapshot.
- Snapshot fields are deterministic and reviewable.

---

## Definition of Done (applies to every story)

- Code is formatted and linted (`make fmt`, `make lint`).
- Tests added/updated and passing (`make test`, plus adoption/contract/acceptance lanes where relevant).
- Any contract/documentation update includes concrete commands and expected output fields.
- Skill workflow updates preserve thin-wrapper boundary (CLI orchestration only).
- `--json` output dependencies are explicit and covered by tests/scripts.
- No new network requirement is introduced in core/baseline adoption workflows.
- GitHub Actions changes preserve deterministic artifacts and stable exit-code behavior.

