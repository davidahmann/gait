# Gait Skills (Thin Wrappers)

The official OSS skill set includes exactly three skills:

- `gait-capture-runpack`
- `gait-incident-to-regression`
- `gait-policy-test-rollout`

These skills are thin wrappers around the Gait CLI and contracts.
They do not reimplement policy logic.

Quick validation:

```bash
python3 scripts/validate_repo_skills.py
```
