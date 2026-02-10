# Launch Kit

This folder is the repeatable distribution package for OSS launch cycles.

Use it when shipping a major release, announcing a new wedge milestone, or running a re-launch.

Contents:

- `narrative_one_liner.md`: positioning statements by audience
- `hn_post.md`: Hacker News launch post + first comment template
- `github_release_template.md`: release page structure and copy
- `faq_objections.md`: short objection handling for common buyer/developer concerns
- `kpi_scorecard.md`: launch metrics and thresholds
- `rfc_openclaw.md`: framework-specific RFC draft for OpenClaw integration
- `rfc_gastown.md`: framework-specific RFC draft for Gas Town integration
- `rfc_agent_framework_template.md`: reusable framework RFC skeleton
- `secure_deployment_openclaw.md`: safe deployment runbook for OpenClaw users
- `secure_deployment_gastown.md`: safe deployment runbook for Gas Town users

Demo asset:

- `scripts/demo_90s.sh` runs the deterministic 90-second terminal flow:
  - install check (`doctor`)
  - first win (`demo`, `verify`)
  - safety boundary (`policy test` block example)
  - incident-to-regress (`regress init` + `regress run`)
  - isolated workspace under `gait-out/demo_90s/workspace` to avoid repo-root artifact residue

Suggested launch sequence:

1. Run `scripts/demo_90s.sh` and capture terminal recording.
2. Generate ecosystem release summary with `python3 scripts/render_ecosystem_release_notes.py`.
3. Publish release with `github_release_template.md` (include integrity assets + `gait.rb`).
4. Post HN thread using `hn_post.md`.
5. Monitor metrics in `kpi_scorecard.md`.
