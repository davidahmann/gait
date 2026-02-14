# PLAN v2.7 Distribution: Integrate Into Existing Gravity Wells

Date: 2026-02-14
Scope: Distribution and adoption packaging only (no core policy/runtime rewrite)

## Goal

Make Gait adoptable with one pull request by meeting teams where they already run work:

1. GitHub Actions first.
2. CI portability second (GitLab CI, Jenkins, CircleCI).
3. Agent tool-boundary integrations third (MCP + SDK middleware).
4. Minimal OpenTelemetry export fourth.

North-star outcome: "one PR to adopt," not "migrate to a new platform."

## Why This Order

- GitHub Actions has the highest OSS distribution leverage and shortest trust path.
- CI portability is mostly packaging and templates once a canonical lane exists.
- Tool-boundary integrations are the enforcement value path, but require more integration detail than CI.
- OTEL is valuable for observability fit, but should stay additive and optional after core adoption paths are frictionless.

## Non-Goals

- No Kubernetes admission-control work in v2.7.
- No hosted control-plane or dashboard-first scope.
- No provider-specific policy logic in core Go paths.
- No breaking schema or exit-code changes.

## Success Metrics (Locked)

- `G1`: A repo can add Gait CI regression gating in one PR using the default GitHub Actions path.
- `G2`: GitHub Actions quickstart path succeeds with deterministic artifacts and stable exit semantics (`0`, `5`, passthrough).
- `C1`: Provider templates for GitLab/Jenkins/Circle are copy-paste usable and map to the same CLI contract.
- `C2`: CI portability docs specify local parity checks before remote CI rollout.
- `B1`: Tool-boundary integration kits cover MCP and at least the blessed SDK middleware path with fail-closed examples.
- `B2`: Adapter parity and adoption smoke suites remain green for official integration references.
- `O1`: Minimal OTEL decision export contract is documented and stable for low-risk fields (`run_id`, `verdict`, `policy_digest`, timing, trace reference).
- `O2`: OTEL export remains optional and does not alter enforcement correctness.

## Workstream 1: GitHub Actions First (Primary Gravity Well)

### W1.1 One-PR GitHub adoption path

Tasks:
- Promote one canonical "drop-in" workflow path in onboarding docs with minimal required edits.
- Keep reusable workflow + composite action path aligned and cross-linked.
- Ensure README path leads directly to CI runbook without ambiguity.

Acceptance:
- New repo can adopt by adding one workflow file and baseline fixture/config.
- Docs show exact expected artifacts and exit-code semantics.

### W1.2 Harden action contract as distribution surface

Tasks:
- Keep action output contract explicit (`exit_code`, summary path, artifact path).
- Validate release-binary verification and deterministic artifact layout in docs/tests.
- Ensure failure triage starts from workflow summary, not raw logs.

Acceptance:
- Action contract is stable and tested by existing regress-template checks.
- Adoption docs contain a deterministic troubleshooting path.

### W1.3 Add provider-agnostic CI handoff from GitHub lane

Tasks:
- Publish "contract-first" mapping from GitHub workflow behavior to shell-based CI contract.
- Keep GitHub as default lane, with portability as extension.

Acceptance:
- Non-GitHub users can identify equivalent contract in under five minutes from docs.

## Workstream 2: CI Portability Kits (Second Gravity Well)

### W2.1 Ship copy-paste templates for major CI providers

Tasks:
- Add versioned templates/examples for:
  - GitLab CI
  - Jenkins
  - CircleCI
- Keep template logic thin: call existing CLI commands and honor stable exit codes.

Acceptance:
- Templates require only path/environment edits, not behavioral rewrites.
- Templates preserve deterministic artifact paths and regress semantics.

### W2.2 Create one compatibility script contract

Tasks:
- Define one shell contract script as source-of-truth behavior.
- Make provider templates wrappers around this contract.

Acceptance:
- Contract script can run locally and in CI runners with equivalent outputs.
- Provider templates differ only in CI syntax, not Gait behavior.

### W2.3 Add CI portability validation gate

Tasks:
- Add deterministic checks for template correctness (path expectations, exit handling, required outputs).
- Include docs checks so examples do not drift.

Acceptance:
- Template regressions are caught before release.

## Workstream 3: Tool-Boundary Integrations (Third Gravity Well)

### W3.1 Prioritize boundary kits with highest adoption pull

Tasks:
- Keep MCP boundary path first-class (`proxy`/`serve`) with fail-closed examples.
- Keep one blessed SDK middleware path as default (OpenAI Agents) and maintain parity references.
- Ensure docs emphasize interception requirement and non-`allow` non-execute behavior.

Acceptance:
- Integration checklist and examples guide users to a default boundary path without branching confusion.
- Adapter parity suite remains the conformance guardrail for official references.

### W3.2 Publish "one PR boundary integration" recipes

Tasks:
- Add minimal recipes for:
  - wrapper middleware insertion
  - sidecar/boundary service path
  - MCP-first path
- Include expected allow/block/require_approval output contracts.

Acceptance:
- Teams can prove fail-closed behavior from recipes with existing acceptance commands.

### W3.3 Keep Python and adapters thin

Tasks:
- Reassert boundary: Go is authoritative for policy/signing/verification.
- Keep SDK/adapters as serialization and subprocess ergonomics only.

Acceptance:
- No framework-local policy semantics added.
- Existing SDK and adapter tests remain contract-compatible.

## Workstream 4: Minimal OpenTelemetry Export (Fourth Gravity Well)

### W4.1 Lock minimal OTEL field contract

Tasks:
- Define a minimal, low-risk event mapping for decision telemetry:
  - run/session/trace references
  - verdict and reason codes
  - policy and intent digests
  - timing fields
- Keep mapping aligned with existing artifact identifiers for offline correlation.

Acceptance:
- OTEL mapping is documented as additive observability output, not source of truth.

### W4.2 Extend and unify OTEL export surfaces

Tasks:
- Keep existing MCP export path stable.
- Evaluate and add equivalent export options on additional decision surfaces only where deterministic and low-risk.
- Avoid mandatory collector/network dependencies.

Acceptance:
- OTEL export can be enabled or disabled without changing enforcement behavior or artifacts.

### W4.3 Publish observability quick recipes

Tasks:
- Expand ingestion docs with minimal collector/SIEM examples tied to the locked field contract.
- Document privacy defaults and recommended redaction posture.

Acceptance:
- Users can route events to existing stacks quickly without schema guesswork.

## Repository Touch Map (Planned)

- `README.md` (GitHub-first adoption call path)
- `docs/ci_regress_kit.md` (primary + portability contract)
- `docs/integration_checklist.md` (boundary default path and one-PR framing)
- `docs/siem_ingestion_recipes.md` (minimal OTEL contract and recipes)
- `examples/ci/*` (GitLab/Jenkins/Circle templates)
- `.github/actions/gait-regress/*` and `.github/workflows/adoption-regress-template.yml` (GitHub lane hardening)
- `scripts/test_ci_regress_template.sh` (existing gate)
- `scripts/test_ci_portability_templates.sh` (planned)

## Validation Gates (for v2.7 slices)

Minimum per-PR validation set for this scope:

- `make lint-fast`
- `make test-fast`
- `bash scripts/test_ci_regress_template.sh`
- `make test-adoption`
- `make test-adapter-parity`
- `make test-v2-6-acceptance`

Additional gates when portability templates land:

- `bash scripts/test_ci_portability_templates.sh` (planned)

## Delivery Sequence

- Phase 1: Workstream 1 (GitHub Actions first)
- Phase 2: Workstream 2 (CI portability templates)
- Phase 3: Workstream 3 (tool-boundary integration kits)
- Phase 4: Workstream 4 (minimal OTEL expansion + docs)

## Definition of Done (v2.7 Distribution)

- GitHub Actions is the clearest and fastest documented adoption path.
- Non-GitHub CI users get copy-paste templates backed by the same CLI contract.
- Boundary integration recipes exist for MCP and blessed middleware path with fail-closed proof.
- Minimal OTEL export is documented and optional, using stable low-risk fields.
- Kubernetes admission-control work is explicitly deferred from v2.7.
