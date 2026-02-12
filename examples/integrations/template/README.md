# Local-Agent Integration Template (Canonical Wrapper Contract)

This template is the canonical minimal wrapper path for OSS v2.3.

Primary sequence (copy/paste):

1. tool payload -> `IntentRequest`
2. `gait gate eval`
3. execute real tool only when `verdict=allow`
4. persist trace path

## 15-Minute Commands

```bash
go build -o ./gait ./cmd/gait
python3 examples/integrations/template/quickstart.py --scenario allow
python3 examples/integrations/template/quickstart.py --scenario block
python3 examples/integrations/template/quickstart.py --scenario require_approval
```

Expected stop/go outputs:

- allow: `verdict=allow`, `executed=true`
- block: `verdict=block`, `executed=false`
- require approval: `verdict=require_approval`, `executed=false`

## CI Mapping

After local wrapper validation:

```bash
gait demo
gait regress init --from run_demo --json
gait regress run --json --junit ./gait-out/junit.xml
```

This path maps directly to `.github/workflows/adoption-regress-template.yml`.

## Secondary Paths

- sidecar mode: `examples/sidecar/gate_sidecar.py`
- MCP proxy mode: `gait mcp proxy`

Use secondary paths only after wrapper lane is green.
