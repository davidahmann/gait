# External Tool Registry To Policy Workflow

Use this workflow when your organization already maintains an approved tool registry and you want Gait policy to enforce it at execution time.

## Goal

Convert external allowlist data into deterministic `gait.gate.policy` YAML and validate it before rollout.

Gait does not own the tool registry. It consumes registry outputs as policy input.

## Supported Allowlist Input Shapes

`scripts/render_tool_allowlist_policy.py` accepts:

1. Object with tools list:
```json
{ "tools": ["tool.read", "tool.write"] }
```

2. Array of tool names:
```json
["tool.read", "tool.write"]
```

3. Array of objects:
```json
[{ "tool_name": "tool.read" }, { "tool_name": "tool.write" }]
```

## Recipe

1. Export allowlist JSON from your registry system.
2. Render policy YAML from the allowlist.
3. Validate with `gait policy test`.
4. Promote using your normal rollout stages (`simulate` -> `dry_run` -> `enforce`).

```bash
python3 scripts/render_tool_allowlist_policy.py \
  --input ./examples/policy/external_tool_allowlist.json \
  --output ./gait-out/policy_external_allowlist.yaml

gait policy test ./gait-out/policy_external_allowlist.yaml examples/policy/intents/intent_read.json --json
gait policy test ./gait-out/policy_external_allowlist.yaml examples/policy/intents/intent_delete.json --json
```

Equivalent Make target:

```bash
make tool-allowlist-policy \
  INPUT=examples/policy/external_tool_allowlist.json \
  OUTPUT=gait-out/policy_external_allowlist.yaml
```

## Example Source File

Use this local fixture as a template:

- `examples/policy/external_tool_allowlist.json`

## Why This Is First-Class

- Keeps registry ownership with your platform/security tooling.
- Keeps enforcement deterministic at the action boundary.
- Produces explicit policy artifacts that can be reviewed, versioned, and tested in CI.
