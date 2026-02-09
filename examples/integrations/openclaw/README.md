# OpenClaw Quickstart (Wrapped Tool Boundary)

This guide shows the minimal wrapped execution path:

1. OpenClaw tool payload -> normalized intent payload.
2. `gait gate eval` against policy.
3. Execute only on `allow`.
4. Persist trace artifact path for audit/debug.

The adapter includes:

- endpoint mapping (`endpoint_class=fs.write`)
- skill provenance (`skill_name`, `publisher`, `source`, digest metadata)

## Run

From repo root:

```bash
go build -o ./gait ./cmd/gait
python3 examples/integrations/openclaw/quickstart.py --scenario allow
python3 examples/integrations/openclaw/quickstart.py --scenario block
```

Expected allow output:

```text
framework=openclaw
scenario=allow
verdict=allow
executed=true
trace_path=/.../gait-out/integrations/openclaw/trace_allow.json
executor_output=/.../gait-out/integrations/openclaw/executor_allow.json
```

Expected block output:

```text
framework=openclaw
scenario=block
verdict=block
executed=false
trace_path=/.../gait-out/integrations/openclaw/trace_block.json
```
