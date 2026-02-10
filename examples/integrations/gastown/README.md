# Gas Town Quickstart (Wrapped Tool Boundary)

This guide shows the minimal wrapped execution path for Gas Town worker actions:

1. Gas Town worker payload -> normalized intent payload.
2. `gait gate eval` against policy.
3. Execute only on `allow`.
4. Persist trace artifact path for audit/debug.

The adapter includes:

- endpoint mapping (`endpoint_class=fs.write`)
- skill provenance (`skill_name`, `publisher`, `source`, digest metadata)
- deterministic output paths in `gait-out/integrations/gastown/`

## Run

From repo root:

```bash
go build -o ./gait ./cmd/gait
python3 examples/integrations/gastown/quickstart.py --scenario allow
python3 examples/integrations/gastown/quickstart.py --scenario block
```

Expected allow output:

```text
framework=gastown
scenario=allow
verdict=allow
executed=true
trace_path=/.../gait-out/integrations/gastown/trace_allow.json
executor_output=/.../gait-out/integrations/gastown/executor_allow.json
```

Expected block output:

```text
framework=gastown
scenario=block
verdict=block
executed=false
trace_path=/.../gait-out/integrations/gastown/trace_block.json
```
