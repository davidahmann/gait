# Local-Agent Integration Template

This is the canonical OSS integration template for local/offline agent runtimes.

It shows the three supported paths:

- wrapper mode (this folder `quickstart.py`)
- sidecar mode (`examples/sidecar/gate_sidecar.py`)
- MCP proxy mode (`gait mcp proxy`)

The wrapper template demonstrates:

1. framework payload -> `IntentRequest`
2. endpoint mapping (`endpoint_class`)
3. skill provenance attachment (`skill_provenance`)
4. `gait gate eval`
5. execute only on `allow`

## Wrapper quickstart

```bash
go build -o ./gait ./cmd/gait
python3 examples/integrations/template/quickstart.py --scenario allow
python3 examples/integrations/template/quickstart.py --scenario block
```

## Sidecar quickstart

```bash
python3 examples/sidecar/gate_sidecar.py \
  --policy examples/policy-test/allow.yaml \
  --intent-file core/schema/testdata/gate_intent_request_valid.json \
  --trace-out ./gait-out/trace_sidecar_allow.json
```

## MCP proxy quickstart

```bash
gait mcp proxy \
  --policy examples/policy-test/allow.yaml \
  --call examples/mcp/openai_function_call.json \
  --adapter openai \
  --json
```
