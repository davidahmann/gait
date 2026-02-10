# Community Patterns

## What to Contribute

- Adapter examples
- Skills and policy packs
- Deterministic fixtures and scenario scripts

## Contribution Funnel

- Open-ended discussion: GitHub Discussions
- Execution proposals:
  - Adapter proposal issue form
  - Skill proposal issue form

Reference: `docs/ecosystem/contribute.md`

## Required Quality Bar

- Deterministic outputs and artifact paths
- `--json` support in command paths
- No bypass around `gait gate eval`
- CI conformance checks pass

## Release Notes Surface

Community entries are rendered into deterministic release notes from:
- `docs/ecosystem/community_index.json`

Validation:
```bash
python3 scripts/validate_community_index.py
make test-ecosystem-automation
```
