# GitHub Release Template

Use this structure for tagged releases.

## Title

`vX.Y.Z - <short release theme>`

## Header

Gait is the offline-first policy-as-code runtime for AI agent tool calls.

Supporting promise:

Prove the install fast, enforce at the tool boundary when you own the seam, and graduate to hardened `oss-prod` readiness explicitly.

## What Shipped

- `<feature 1>` with path references
- `<feature 2>` with path references
- `<feature 3>` with path references
- `gait mcp verify` or `gait trace` only if the corresponding local example/docs/tests landed

## First 5 Minutes

```bash
curl -fsSL https://raw.githubusercontent.com/Clyra-AI/gait/main/scripts/install.sh | bash
gait init --json
gait check --json
gait demo
gait verify run_demo --json
gait regress bootstrap --from run_demo --json --junit ./gait-out/junit.xml
```

This is the fast-proof path. It is not the same thing as hardened `oss-prod` readiness, and it does not imply strict inline enforcement without a real interception seam.

## Security Boundary Example

```bash
gait policy test examples/prompt-injection/policy.yaml examples/prompt-injection/intent_injected.json --json
```

Expected:

- `verdict: block`
- `reason_codes: ["blocked_prompt_injection"]`

## v2.3 Release Gate Snapshot

Attach and summarize:

- `gait-out/v2_3_metrics_snapshot.json`
- `gait-out/integration_lane_scorecard.json`

Required release check:

- `release_gate_passed: true`
- `M1..M4`, `C1..C3`, `D1..D3` all pass

## Integrity Artifacts

- `checksums.txt`
- `checksums.txt.sig`
- `checksums.txt.pem`
- `checksums.txt.intoto.jsonl`
- `sbom.spdx.json`
- `provenance.json`
- `gait.rb` (Homebrew formula asset)

## Upgrade Notes

- breaking changes: `<none | list>`
- schema or exit-code compatibility notes: `<notes>`

## Go / No-Go

- decision: `<GO | NO-GO>`
- doctor truthfulness evidence: `gait doctor --json`
- production-readiness evidence: `gait doctor --production-readiness --json`
- workflow runtime guard evidence: `python3 scripts/check_github_action_runtime_versions.py .github/workflows docs/adopt_in_one_pr.md`
- staged-boundary note: fast proof != hardened `oss-prod` readiness; strict inline enforcement requires interception before tool execution
- residual non-blocking risks: `<none | list>`

## Docs

- `README.md`
- `docs/install.md`
- `docs/integration_checklist.md`
- `docs/contracts/intent_receipt_conformance.md`
- `docs/ci_regress_kit.md`
- `docs/ecosystem/awesome.md`
- `docs/launch/README.md`
