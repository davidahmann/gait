# Project Defaults (Lightweight Config)

Use a local project config to avoid repeating the same `gait gate eval` flags.

Default lookup path:

- `.gait/config.yaml`

Command flags:

- `--config <path>` to load a custom config file
- `--no-config` to disable config lookup for one command

## Minimal Example

```yaml
gate:
  policy: examples/policy/base_high_risk.yaml
  profile: oss-prod
  key_mode: prod
  private_key: examples/scenarios/keys/approval_private.key
  credential_broker: stub
```

With this file present, the repeated command becomes:

```bash
gait gate eval --intent examples/policy/intents/intent_delete.json --json
```

## Supported `gate` Defaults

- `policy`
- `profile`
- `key_mode`
- `private_key`
- `private_key_env`
- `approval_public_key`
- `approval_public_key_env`
- `approval_private_key`
- `approval_private_key_env`
- `rate_limit_state`
- `credential_broker`
- `credential_env_prefix`
- `credential_ref`
- `credential_scopes`
- `credential_command`
- `credential_command_args`
- `credential_evidence_path`
- `trace_path`

## Guardrails

- CLI flags always override config values.
- Missing default `.gait/config.yaml` is ignored.
- Missing explicit `--config <path>` returns an input error.
- Keep sensitive key files out of git and prefer `*_env` options in shared repos.
