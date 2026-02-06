# AutoGen Quickstart (Wrapped Tool Boundary)

This guide shows the minimal wrapped execution path:

1. AutoGen function-call payload -> normalized intent payload.
2. `gait gate eval` against policy.
3. Execute only on `allow`.
4. Persist trace artifact path for audit/debug.

## Run

From repo root:

```bash
go build -o ./gait ./cmd/gait
python3 examples/integrations/autogen/quickstart.py --scenario allow
python3 examples/integrations/autogen/quickstart.py --scenario block
```

Expected allow output:

```text
framework=autogen
scenario=allow
verdict=allow
executed=true
trace_path=/.../gait-out/integrations/autogen/trace_allow.json
executor_output=/.../gait-out/integrations/autogen/executor_allow.json
```

Expected block output:

```text
framework=autogen
scenario=block
verdict=block
executed=false
trace_path=/.../gait-out/integrations/autogen/trace_block.json
```

Trace record location:

- `gait-out/integrations/autogen/trace_allow.json`
- `gait-out/integrations/autogen/trace_block.json`
