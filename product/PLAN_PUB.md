# PLAN_PUB

## Objective
Turn current partial distribution assets into a productized adoption engine:
- GitHub Marketplace action path
- one-PR CI template path with evidence artifacts
- canonical MCP distribution path
- PackSpec standard-setting package (stability + producer kit + compatibility matrix)

## Current State (from repo audit)
- Composite action exists: `.github/actions/gait-regress/action.yml`.
- Reusable CI template exists: `.github/workflows/adoption-regress-template.yml`.
- Integrations and MCP docs exist (`docs/mcp_capability_matrix.md`, OpenClaw skill path).
- PackSpec contract + TCK docs exist (`docs/contracts/packspec_v1.md`, `docs/contracts/packspec_tck.md`).
- Missing as first-class deliverables:
  - Marketplace-publishable action repository shape
  - explicit `pack verify` step in the adoption template lane
  - canonical MCP demo that clearly shows allow/block/require_approval + emitted packs
  - producer kit for third parties
  - public compatibility matrix (Gait version <-> PackSpec version <-> verifier version)

## Workstream 1: GitHub Marketplace Action

### Why publish is not visible now
Based on GitHub docs for Marketplace publishing, current repo layout does not satisfy listing prerequisites:
- Marketplace requires a single `action.yml` at repository root for listing.
- Current action metadata is in `.github/actions/gait-regress/action.yml`.
- Marketplace docs also state action repos must not contain workflow files.
- Current repo has many workflow files in `.github/workflows/`.

Likely outcome: the "Publish to Marketplace" path will not auto-enable for this monorepo layout.

### Plan
1. Create a dedicated public repo for marketplace action, for example `gait-regress-action`.
2. Move/copy action to root as `/action.yml` with action runtime files and README only.
3. Keep this monorepo action path for internal template usage, but treat it as non-marketplace distribution.
4. Ensure metadata quality:
- unique `name`
- clear `description`
- add `branding` (icon/color) for marketplace badge clarity
5. Confirm owner/org accepted GitHub Marketplace Developer Agreement.
6. Cut signed release tag (`v1.0.0`) and publish via release flow.
7. Add immutable major tag management (`v1` -> latest v1.x).
8. Back-link from this repo README/docs to marketplace action repo.

### Acceptance Criteria
- Marketplace listing exists and is searchable.
- External repo can use `owner/repo@v1` and receive uploaded evidence artifacts.
- Listing docs show regress + policy-test modes and artifact outputs.

## Workstream 2: Productized One-PR Adoption Path

### Goal
Fork + green CI + evidence artifacts in under 3 minutes.

### Plan
1. Keep reusable workflow as primary: `.github/workflows/adoption-regress-template.yml`.
2. Add explicit `pack verify` step to template lane before final artifact upload.
3. Add minimal template repo (separate from monorepo examples) containing:
- one policy block fixture
- one pack verification step
- one incident-to-regression gate
4. Provide single copy/paste quickstart in template README with exact expected outputs.
5. Standardize artifact bundle names and paths (deterministic across local/CI).
6. Add a dedicated 60-second PR evidence attachment demo that mirrors template output names.

### Acceptance Criteria
- Clean fork reaches first green check with artifacts in under 3 minutes on hosted runner.
- Uploaded artifacts include regress result, JUnit, trace, pack verify output, and ticket footer text.
- Demo video exactly matches template commands and artifact names.

## Workstream 3: MCP as Distribution Engine #2

### Goal
Canonical MCP boundary demo with enforcement + evidence, not just reference docs.

### Plan
1. Publish one canonical MCP walkthrough that demonstrates in one flow:
- `allow`
- `block`
- `require_approval`
2. Ensure each path emits trace + runpack/pack artifacts and shows where they are stored.
3. Add one script entrypoint for reproducible demo run (single command).
4. Wire this flow into acceptance lane so it cannot regress silently.
5. Add a short docs page linking runtime contract and artifact verification commands.

### Acceptance Criteria
- Single command run yields all three verdict classes.
- Non-allow paths are clearly non-executing.
- Evidence artifacts are emitted and verified with documented commands.

## Workstream 4: Lock Standard-Setting Advantage

### 4.1 PackSpec v1 stability contract
Status: mostly present. Strengthen with explicit non-breaking policy and deprecation policy section.

### 4.2 Producer kit
Plan:
1. Publish a tiny producer kit (language-agnostic doc + minimal reference code) that emits valid packs.
2. Include schema validation and deterministic zip rules.
3. Include compatibility test against `gait pack verify`.

Acceptance:
- Third-party can emit a valid pack without adopting full Gait runtime.

### 4.3 Public compatibility matrix
Plan:
1. Add docs page: `Compatibility Matrix`.
2. Track rows for:
- Gait CLI version
- PackSpec version(s)
- verifier behavior/version constraints
3. Include compatibility guarantees and migration rules.

Acceptance:
- Maintainers can decide upgrade path without reading code or changelog diff-by-diff.

## Execution Order
1. Marketplace repo split and publishability fix.
2. One-PR adoption template hardening (`pack verify` + fast path timings).
3. 60-second evidence demo aligned with template.
4. Canonical MCP enforcement + evidence lane.
5. Producer kit + compatibility matrix.

## Risks
- Marketplace constraints conflict with monorepo structure.
- CI time budget may exceed 3-minute target unless template is aggressively minimal.
- MCP demo can drift from runtime reality if not covered by acceptance tests.

## Tracking Metrics
- Time to first green CI from fork.
- Artifact completeness rate per run.
- Action adoption count (marketplace + template usage).
- MCP demo pass rate in CI.
- External producer-kit conformance pass rate.
