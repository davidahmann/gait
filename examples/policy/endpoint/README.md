# Endpoint Policy Examples

These files demonstrate endpoint constraint controls in Gate:

- `allow_safe_endpoints.yaml`: allow path + domain allowlisted endpoint actions.
- `block_denied_endpoints.yaml`: block denylisted filesystem paths and domains.
- `require_approval_destructive.yaml`: require approval for destructive endpoint actions.

Fixture intents:

- `fixtures/intent_allow.json`
- `fixtures/intent_block.json`
- `fixtures/intent_destructive.json`

Validation commands:

```bash
gait policy test examples/policy/endpoint/allow_safe_endpoints.yaml examples/policy/endpoint/fixtures/intent_allow.json --json
gait policy test examples/policy/endpoint/block_denied_endpoints.yaml examples/policy/endpoint/fixtures/intent_block.json --json
gait policy test examples/policy/endpoint/require_approval_destructive.yaml examples/policy/endpoint/fixtures/intent_destructive.json --json
```

Expected exit codes:

- allow fixture -> `0`
- block fixture -> `3`
- require approval fixture -> `4`
