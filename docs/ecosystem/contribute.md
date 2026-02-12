# Ecosystem Contribution Funnel

Use this path to propose a community adapter or skill and get it listed.

## 1) Open The Correct Proposal

- adapter proposal: `.github/ISSUE_TEMPLATE/adapter.yml`
- skill proposal: `.github/ISSUE_TEMPLATE/skill.yml`

Required in proposal:

- problem solved and target runtime/framework
- minimal runnable example path
- fail-closed behavior (`allow` vs non-`allow` execution)
- deterministic output paths + test plan

## 2) Respect v2.3 Lane Governance

Blessed lane in v2.3:

- coding-agent wrapper + GitHub Actions regress template

No new official lane is merged without scorecard evidence:

```bash
python3 scripts/check_integration_lane_scorecard.py \
  --input gait-out/adoption_metrics.json \
  --out gait-out/integration_lane_scorecard.json
```

Gate for official-lane expansion:

- selected score >= `0.75`
- confidence delta >= `0.03`

If threshold is not met, lane can still ship as community/experimental.

## 3) Add Or Update Community Index Entry

Edit `docs/ecosystem/community_index.json`:

- add one entry with unique `id`
- point `repo` to public GitHub URL
- set `source` to `community` unless maintained in this repo
- set initial `status` to `experimental` for new submissions

Validate locally:

```bash
python3 scripts/validate_community_index.py
```

## 4) Prove Contract Compatibility

Adapter submissions should include:

- intent evaluation via `gait gate eval` (or MCP boundary equivalent)
- deterministic trace/run artifact paths
- explicit non-`allow` non-execution behavior

Validation commands:

```bash
make lint
make test-adoption
make test-adapter-parity
```

Skill submissions should include:

- provenance metadata (`source`, `publisher`, digest/signature where applicable)
- wrapper-only orchestration (no policy parsing/evaluation logic outside Go)

Validation commands:

```bash
make lint
make test-skill-supply-chain
```

## 5) Review And Listing

A contribution is listed in `docs/ecosystem/awesome.md` once:

- index entry validates in CI
- reviewer confirms deterministic/no-bypass behavior
- install + quickstart commands are documented

## 6) Release Automation Artifact

Before tagged release, generate deterministic ecosystem release notes:

```bash
python3 scripts/render_ecosystem_release_notes.py
```

Output:

- `gait-out/ecosystem_release_notes.md`
