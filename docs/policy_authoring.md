# Policy Authoring Workflow

Use this workflow to reduce policy rollout mistakes before runtime enforcement.

## Objective

Make policy changes deterministic, reviewable, and low-friction for developers:

- scaffold from a baseline template
- validate syntax and semantics early
- normalize formatting deterministically
- test verdict behavior against intent fixtures

## Recommended Authoring Loop

```bash
gait policy init baseline-mediumrisk --out gait.policy.yaml --json
gait policy validate gait.policy.yaml --json
gait policy fmt gait.policy.yaml --write --json
gait policy test gait.policy.yaml examples/policy/intents/intent_write.json --json
gait policy simulate --baseline examples/policy/base_medium_risk.yaml --policy gait.policy.yaml --fixtures examples/policy/intents --json
```

Interpretation:

- `policy validate` checks strict YAML parsing + policy semantics only.
- `policy fmt` rewrites normalized YAML deterministically.
- `policy test` evaluates one intent fixture and returns verdict, reason codes, and `matched_rule`.
- `policy simulate` compares baseline vs candidate verdicts over fixture corpora and recommends rollout stage (`observe`, `require_approval`, `enforce`).

## Failure Semantics

- exit `0`: valid / allow path
- exit `3`: policy test verdict `block`
- exit `4`: policy test verdict `require_approval`
- exit `6`: invalid input, parse error, unknown field, or invalid schema/value

Treat exit `6` as fail-closed in CI and production rollout lanes.

## IDE Schema Wiring (YAML Language Server)

For editors that support YAML schema mapping, point policy files at:

- `schemas/v1/gate/policy.schema.json`

Example VS Code workspace settings:

```json
{
  "yaml.schemas": {
    "./schemas/v1/gate/policy.schema.json": [
      "gait.policy.yaml",
      "examples/policy/**/*.yaml"
    ]
  }
}
```

This gives fast feedback for enum values and unknown keys before runtime.

## Team Workflow Recommendation

- Require `policy validate` + fixture `policy test` in pre-merge CI.
- Run `policy simulate` against representative fixture sets before changing rollout stage.
- Keep policy files formatted by `policy fmt --write` before review.
- Review policy changes with fixture deltas and matched-rule evidence, not raw YAML diff alone.

## v2.1 Additive Readiness (Current Best Practice)

Even before delegation/session-chain features fully land, include representative context in fixtures:

- `context.session_id`
- `context.request_id`
- `context.auth_context`
- `context.credential_scopes`
- `context.environment_fingerprint`

This keeps policy tests stable while integration payloads evolve.

## Signing Key Lifecycle (Local)

For production verification profiles and trace signing workflows, manage keys with CLI primitives:

```bash
gait keys init --out-dir ./gait-out/keys --prefix prod --json
gait keys rotate --out-dir ./gait-out/keys --prefix prod --json
gait keys verify --private-key ./gait-out/keys/prod_private.key --public-key ./gait-out/keys/prod_public.key --json
```

## Example Policy Packs

- baseline templates: `examples/policy/base_low_risk.yaml`, `examples/policy/base_medium_risk.yaml`, `examples/policy/base_high_risk.yaml`
- endpoint controls: `examples/policy/endpoint/*`
- skill trust controls: `examples/policy/skills/*`
- simple guard fixtures: `examples/policy-test/*`
