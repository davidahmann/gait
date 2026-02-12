# PLAN v1.8: Ecosystem Interception and Distribution Lift

Date: 2026-02-10
Source of truth: `product/PRD.md`, `product/ROADMAP.md`, `product/PLAN_v1.md`, `product/PLAN_v1.7.md`, current codebase (`main`)
Scope: OSS v1 execution-layer expansion only. ENT platform remains out of runtime scope.

This plan closes the remaining OSS wedge gaps identified after v1.7:

- official installable OpenClaw skill/package with one-command path
- long-running network interception mode for MCP proxy
- first-class Gas Town integration artifact
- deterministic Beads bridge for block/violation traces
- stronger framework-specific distribution kit

---

## v1.8 Objective

Turn existing adapter references into production-adjacent integration assets that teams can install, run continuously, and operationalize in workflow tooling.

Result target:
- Gait is no longer only "examples + CLI", but also a drop-in execution boundary package for fast-moving agent OSS ecosystems.

---

## Locked Product Decisions (OSS v1.8)

- Keep Gait category boundary stable: execution-boundary control + proof, not orchestration.
- Keep OSS offline-first and fail-closed for non-`allow` outcomes.
- Keep new integration assets deterministic and CI-verified.
- Keep MCP network interception mode local-first (loopback by default); no hosted dependency.
- Keep Beads bridge optional and local-tooling-aware (no hard runtime dependency on `bd`).
- Do not add multi-tenant control-plane features to OSS path.

---

## Current Baseline (Observed in Codebase)

Already implemented:
- Adapter parity examples and tests for OpenAI Agents, LangChain, AutoGen, OpenClaw, AutoGPT.
- `gait mcp proxy` and `gait mcp bridge` one-shot CLI evaluation flow.
- Skill provenance model and policy hooks in Gate.
- Registry install/verify primitives for signed packs.
- Ecosystem docs (`docs/ecosystem/*`) and launch docs (`docs/launch/*`).

Current gaps to close in v1.8:
- OpenClaw integration exists as quickstart but not an official installable skill/plugin path.
- MCP path is one-shot CLI, not long-running runtime interception service.
- Gas Town does not yet have an explicit first-class artifact/hook/guide.
- No deterministic bridge from blocked traces into Beads issue workflow.
- Distribution package lacks framework-specific RFC and secure deployment playbooks.

---

## v1.8 Exit Criteria

v1.8 is complete only when all are true:

- OpenClaw has an official one-command install path that materializes a runnable Gait boundary skill package.
- `gait mcp serve` (or equivalent) runs as a long-lived local network service and returns deterministic gate outputs.
- Gas Town has a first-class integration artifact and docs, aligned with existing adapter contract.
- Beads bridge can create deterministic issue payloads from blocked/approval-required traces.
- Launch/distribution docs include framework-specific RFC and secure deployment kits.
- CI enforces v1.8 acceptance without reducing existing coverage and safety gates.

---

## Epic V18.0: OpenClaw Official Skill Package

Objective: convert OpenClaw integration from reference script into an installable package path.

### Story V18.0.1: Skill package layout and manifest

Tasks:
- Add an official OpenClaw skill package directory containing:
  - executable wrapper that routes tool actions through Gait gate flow
  - package metadata/manifest
  - default policy stub and config template
- Include deterministic package version and digest metadata.

Repo paths:
- `examples/integrations/openclaw/skill/`
- `examples/integrations/openclaw/README.md`
- `docs/contracts/skill_provenance.md` (reference updates if needed)

Acceptance criteria:
- Package directory is self-contained and documented.
- Package metadata can be validated deterministically in tests.

### Story V18.0.2: One-command installer

Tasks:
- Add a one-command installer script for OpenClaw skill package.
- Support target directory override for local testability.
- Emit deterministic JSON output in machine-readable mode.

Repo paths:
- `scripts/install_openclaw_skill.sh`
- `Makefile` (new helper target)
- `docs/integration_checklist.md`

Acceptance criteria:
- Installer creates expected file tree with executable permissions.
- Re-run is idempotent (safe overwrite/update behavior).

### Story V18.0.3: Install and runtime smoke checks

Tasks:
- Add tests for installer correctness and minimal runtime invocation contract.

Repo paths:
- `scripts/test_openclaw_skill_install.sh`
- `.github/workflows/ci.yml`

Acceptance criteria:
- CI fails if installer output layout drifts.

---

## Epic V18.1: Long-Running MCP Network Interception

Objective: provide a drop-in runtime service for continuous tool-call gating.

### Story V18.1.1: MCP serve command

Tasks:
- Add long-running command (for example `gait mcp serve`) with local HTTP endpoint.
- Accept adapter + tool-call payload requests.
- Reuse existing Gate evaluation, trace signing, and optional export behavior.
- Provide health endpoint and deterministic JSON responses.

Repo paths:
- `cmd/gait/mcp.go` and/or `cmd/gait/mcp_server.go`
- `core/mcp/` (request/response helpers if needed)
- `docs/architecture.md`
- `docs/flows.md`

Acceptance criteria:
- Service starts, handles multiple requests, and shuts down cleanly.
- Responses preserve deterministic verdict/reason/trace semantics.

### Story V18.1.2: Service hardening and fail-closed behavior

Tasks:
- Enforce loopback-default bind (`127.0.0.1`) unless explicit override.
- Return deterministic errors for malformed payload, unsupported adapter, and policy failures.
- Keep non-`allow` outcomes non-executing.

Repo paths:
- `cmd/gait/mcp.go` and/or `cmd/gait/mcp_server.go`
- `cmd/gait/mcp_test.go`
- `internal/e2e/`

Acceptance criteria:
- Fail-closed semantics match one-shot `mcp proxy`.
- Service-mode tests validate repeated calls and error paths.

---

## Epic V18.2: Gas Town First-Class Integration Artifact

Objective: provide explicit Gas Town integration path instead of implied compatibility.

### Story V18.2.1: Gas Town adapter artifact

Tasks:
- Add `examples/integrations/gastown/` with:
  - quickstart wrapper
  - allow/block policy fixtures
  - deterministic output paths in `gait-out/integrations/gastown/`
- Align output contract to existing adapter parity fields.

Repo paths:
- `examples/integrations/gastown/`
- `examples/integrations/README.md`
- `scripts/test_adapter_parity.sh`

Acceptance criteria:
- Gas Town adapter passes allow/block parity checks.

### Story V18.2.2: Gas Town operational guide

Tasks:
- Add focused guide for integrating Gait into Gas Town worker execution path.
- Include "where to hook", fail-closed behavior, and incident-to-regress workflow.

Repo paths:
- `docs/integration_checklist.md`
- `docs/ecosystem/awesome.md`
- `docs/wiki/Integration-Cookbook.md`

Acceptance criteria:
- Guide is copy-paste usable and references runnable example paths.

---

## Epic V18.3: Beads Bridge for Blocked/Violation Traces

Objective: turn blocked traces into deterministic work items.

### Story V18.3.1: Trace-to-beads bridge tool

Tasks:
- Add bridge command/script that:
  - reads signed trace JSON
  - extracts deterministic summary (`verdict`, `reason_codes`, `violations`, digests, trace path)
  - creates a Beads issue when `bd` is available
  - supports dry-run output when `bd` is unavailable
- Add title/body template with stable format.

Repo paths:
- `scripts/bridge_trace_to_beads.sh` (or equivalent)
- `docs/wiki/Community-Patterns.md`
- `docs/approval_runbook.md`

Acceptance criteria:
- Dry-run mode always works without `bd`.
- Live mode creates issue with deterministic content fields.

### Story V18.3.2: Bridge tests and fixtures

Tasks:
- Add fixture traces and test harness for both dry-run and live-command simulation.

Repo paths:
- `scripts/test_beads_bridge.sh`
- `core/schema/testdata/` (fixture reuse where possible)
- `Makefile`

Acceptance criteria:
- Bridge output contract is CI-enforced.

---

## Epic V18.4: Distribution and Launch Kit Upgrade

Objective: strengthen ecosystem adoption assets around real integration points.

### Story V18.4.1: Framework-specific RFC templates

Tasks:
- Add framework RFC templates for:
  - OpenClaw
  - Gas Town
  - AutoGPT/Open-source orchestrators pattern
- Include concrete problem framing, integration points, and acceptance checks.

Repo paths:
- `docs/launch/rfc_openclaw.md`
- `docs/launch/rfc_gastown.md`
- `docs/launch/rfc_agent_framework_template.md`

Acceptance criteria:
- Templates are specific enough for direct issue/discussion submission.

### Story V18.4.2: Community-facing secure deployment guides

Tasks:
- Add deploy-safe guides for OpenClaw and Gas Town:
  - minimum policy set
  - required provenance fields
  - fail-closed recommendations
  - SIEM/export hooks

Repo paths:
- `docs/launch/secure_deployment_openclaw.md`
- `docs/launch/secure_deployment_gastown.md`
- `docs/zero_trust_stack.md`

Acceptance criteria:
- Guides map directly to runnable repo artifacts and commands.

### Story V18.4.3: Discovery surface updates

Tasks:
- Update ecosystem index/docs to include new OpenClaw skill package and Gas Town artifact.
- Ensure docs-site navigation includes new high-value distribution docs if appropriate.

Repo paths:
- `docs/ecosystem/community_index.json`
- `docs/ecosystem/awesome.md`
- `docs/README.md`
- `README.md`

Acceptance criteria:
- New assets are discoverable from top-level docs paths.

---

## Epic V18.5: v1.8 Acceptance and CI Enforcement

Objective: lock the new wedge guarantees as non-regression checks.

### Story V18.5.1: v1.8 acceptance script

Tasks:
- Add `scripts/test_v1_8_acceptance.sh` covering:
  - OpenClaw skill install path
  - MCP serve request/response path
  - Gas Town adapter parity path
  - Beads bridge dry-run contract
  - distribution artifact existence checks

Repo paths:
- `scripts/test_v1_8_acceptance.sh`
- `Makefile`

Acceptance criteria:
- Script runs deterministically from clean checkout.

### Story V18.5.2: CI job integration

Tasks:
- Add dedicated CI job for v1.8 acceptance.
- Ensure no reduction in existing lint/test/coverage gates.

Repo paths:
- `.github/workflows/ci.yml`

Acceptance criteria:
- PRs fail when v1.8 integration/distribution guarantees drift.

---

## Measurement and Success Metrics (v1.8)

Adoption metrics:
- `A1`: count of successful one-command OpenClaw skill installs in operator feedback loops.
- `A2`: Gas Town artifact usage in integration pilots.
- `A3`: MCP serve usage for local runtime interception workflows.

Operational metrics:
- `O1`: MCP serve p95 request latency within documented budget.
- `O2`: non-`allow` request handling remains fail-closed with stable exit/reason mapping.
- `O3`: Beads bridge dry-run/live-mode success rate.

Distribution metrics:
- `D1`: framework RFC/template consumption in launch workflows.
- `D2`: ecosystem index updates per release without schema drift.

---

## Explicit Non-Goals (v1.8)

- building or hosting a centralized runtime dashboard
- adding tenant-aware control-plane orchestration
- replacing customer identity/vault/gateway systems
- introducing prompt/content scanning engines
- changing OSS artifact semantics required by v1.x contracts

---

## Definition of Done (v1.8)

- OpenClaw integration ships as an official installable package with one-command path.
- MCP interception supports long-running network service mode with deterministic responses.
- Gas Town has a first-class adapter artifact and guide aligned to parity contract.
- Beads bridge exists for block/violation trace triage with deterministic output.
- Distribution kit includes framework RFCs and secure deployment guides.
- CI enforces v1.8 acceptance while preserving existing safety and coverage gates.
