---
name: app-audit
description: Run an evidence-based audit for Gait OSS project across product, architecture, DX, security, and GTM readiness; output a strict go/no-go report with P0-P3 blockers and launch risk by technology, messaging, and expectations.
disable-model-invocation: true
---

# Gait OSS Audit

Execute this workflow when asked to perform app review, release readiness, architecture clarity, or public-launch audit for Gait.

## Scope

- Repository: `/Users/tr/gait`
- Analyze current code/docs only; do not invent features/markets.
- Default mode is read-only unless user explicitly asks for fixes.

## Workflow

1. Build whiteboard mental model from first-screen README/onboarding to sustained use.
2. Map personas, JTBD, MVP user-story coverage, and the first-10-minutes experience.
3. Audit architecture boundaries and code organization:
   - orchestration thickness vs focused packages
   - explicit side effects in API names/signatures
   - symmetry of public semantics (`plan` vs `apply`, `read` vs `read+validate`)
   - timeout/cancellation propagation in long-running flows
   - extension points vs likely enterprise forks
4. Audit public contract surfaces:
   - stable vs internal surface map
   - schema/versioning policy clarity and migration expectations
   - machine-readable error support for library or SDK consumers
   - install/version discoverability
5. List required inputs/config/secrets/dependencies and setup friction.
6. Audit docs for integration clarity:
   - README first screen answers what it is, who it is for, how it integrates, and how to get value quickly
   - integration before internals
   - problem -> solution framing before primitive lists
   - file/state lifecycle and canonical path model
   - single docs source of truth and how generated/public docs derive from repo docs
7. Audit OSS trust baseline:
   - `CONTRIBUTING.md`, `CHANGELOG.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`
   - issue/PR templates
   - maintainer/org context and support expectations
   - standalone vs ecosystem dependency clarity
   - public vs private planning artifact choices
8. Run technical validation on affected surfaces (build/lint/tests as needed).
9. Compare stated product intent vs implemented behavior.
10. Assess security posture and fail-closed guarantees.
11. Assess market wedge sharpness for existing personas only.
12. Produce final go/no-go verdict with minimum blocker set.

## Non-Negotiables

- Evidence-first: every claim must cite command output or file path.
- Contract-first: evaluate behavior, boundaries, and integration guarantees before polish.
- Boundary-first: explicitly separate ownership and integration points.
- Incident-first: lead with failure scenarios and operational impact.
- No cosmetics as blockers.
- Distinguish facts vs inference.

## Command Anchors

- `gait doctor --json` to capture environment diagnostics in machine-readable form.
- `gait pack inspect <artifact.zip> --json` to inspect artifact envelope integrity and payload shape.
- `gait gate eval --policy <policy.yaml> --intent <intent.json> --json` to validate fail-closed policy behavior.

## Severity

- P0: release blocker / high reputational risk
- P1: major launch risk
- P2: meaningful gap, non-blocking for project
- P3: polish

## Output Contract

- Section 1: End-to-end product model
- Section 2: Persona/story coverage map
- Section 3: Contract and boundary audit
- Section 4: Inputs/install/config friction table
- Section 5: Docs and OSS trust-baseline audit
- Section 6: Technical audit + release readiness
- Section 7: Business/market fit assessment
- Section 8: Final verdict (go/no-go) + top 3 launch risks

Each section must include:
- Findings
- Evidence references
- Risk color (Green/Yellow/Red)
- Blockers (if any)
- Minimum fix set (only release-critical)
