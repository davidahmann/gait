---
name: code-review
description: Perform Codex-style full-repository review for Gait (not PR-limited), with severity-ranked findings focused on regressions, fail-closed safety, determinism, portability, and docs/CLI contract correctness.
disable-model-invocation: true
---

# Full-Repo Code Review (Gait)

Execute this workflow for: "review the codebase", "audit repo health", "run a full code review", or "find risks in Gait."

## Reviewer Personality

- Contract-first: behavior and guarantees over style.
- Boundary-first: controls belong at the execution boundary, not prompt/UI layers.
- Regression-first: look for latent breakage paths.
- Fail-closed safety bias: block safety/control weakening.
- Determinism required: same input/state should produce the same decisions and artifacts.
- Scenario-driven: each finding includes concrete break path and impact.
- Portability-aware: Linux/macOS/CI/toolchain/path behavior.
- Signal over noise: findings-first, severity-ranked output.

## Scope

- Repository root: `/Users/tr/gait`
- Review entire repo, not only current diffs.
- Prioritize high-risk surfaces first, then remaining components.

## High-Risk Surfaces (Priority Order)

1. `core/gate`, `core/contextproof`, `core/pack`, `core/runpack`, `core/regress`, `core/jobruntime`
2. `cmd/gait` CLI behavior, flags, exit codes, JSON outputs
3. `core/mcp` and adapter boundaries
4. `sdk/python` wrapper behavior and error mapping
5. `schemas/v1` and compatibility-sensitive artifacts
6. `docs`, `README.md`, `docs-site`, OSS baseline files, and `.github` workflow/template surfaces

## Workflow

1. Build repository map and contract map from code/tests/help text, including stable/internal/deprecated surfaces, schema/version/CLI/exit-code compatibility rules, authoritative-core boundaries, and docs source-of-truth flow.
2. Run baseline validation where feasible (lint/build/tests) and record gaps if not run.
3. Review each subsystem for:
   - Enforceable boundary violations, including controls living in prompt/UI paths instead of execution/runtime boundaries
   - Contract drift before logic drift: schema/version/CLI/exit-code breaks, non-additive evolution, or missing dual-reader compatibility
   - Public APIs that hide side effects, blur `plan` vs `apply`, or weaken scoped approvals, out-of-band stop, and destructive-budget safety patterns
   - Missing timeout/cancellation propagation, crash-safe state handling, lock/permission discipline, or contention coverage in long-running flows
   - Fail-closed gaps where ambiguity in high-risk paths can silently allow instead of block
   - Determinism or reproducibility breaks across decisions, hashes, traces, diagnostics, and artifacts for the same input/state
   - Authoritative-core leaks where wrappers/SDKs duplicate policy, signing, verification, or other decision logic
   - Missing machine-readable failures: stable JSON envelopes, error taxonomy, retryability hints, or consistent error mapping on public boundaries
   - Missing operator observability: deterministic `doctor` output, correlation IDs, or local structured logs
   - False-green CI/release paths, including risk-tiered lane gaps (fast/core/acceptance/cross-platform/chaos/perf), weak merge policy wiring, supply-chain verification gaps, or missing post-merge regression monitoring
   - Weak adoption architecture: unclear quickstart, missing expected outputs/integration diagrams, or troubleshooting-only-afterthought docs
   - Governance and launch drift: missing ADRs, risk register, explicit non-goals, definition-of-done, or OSS trust-baseline artifacts
4. Verify findings with concrete evidence (file refs, commands, test output).
5. Rank findings by severity and confidence.
6. Report minimum blocker set for safe release posture.

## Severity Model

- P0: release blocker, severe safety/integrity break, high reputational risk.
- P1: major behavioral regression or control bypass with real user impact.
- P2: meaningful correctness/portability/docs-contract issue.
- P3: minor maintainability concern.

## Finding Format

- `Severity`: P0/P1/P2/P3
- `Title`: short and action-oriented
- `Location`: file + line
- `Problem`: what is wrong
- `Break Scenario`: concrete failure path
- `Impact`: user/safety/CI/compliance effect
- `Fix Direction`: minimal safe correction

## Review Rules

- Findings are primary output; summaries stay brief.
- Do not report style nits unless they cause runtime/contract risk.
- Do not claim tests/commands were run if they were not.
- Separate facts from inference.
- Prefer the smallest finding that exposes the broken contract or unsafe boundary.
- If no findings, explicitly state `No material findings` and list residual risks/testing gaps.

## Command Anchors

- `gait doctor --json` to verify baseline runtime diagnostics and dependency posture.
- `gait gate eval --policy <policy.yaml> --intent <intent.json> --json` to validate policy verdict/exit behavior.
- `gait pack verify <artifact.zip> --json` to check artifact integrity and signature status.

## Output Contract

1. `Findings` (required, ordered by severity)
2. `Subsystem Coverage` (Green/Yellow/Red per major area)
3. `Open Questions / Assumptions` (if any)
4. `Residual Risk / Testing Gaps`
5. `Final Judgment`:
   - technical health today
   - minimum blockers (if any)
   - top 3 risk concentrations
   - merge/release confidence given current CI, observability, and release-integrity posture
