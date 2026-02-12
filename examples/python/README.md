# Python Reference Adapter

This path is the minimal Python wrapper integration contract.

## Canonical Flow

1. capture typed `IntentRequest`
2. evaluate with `ToolAdapter.execute(...)`
3. execute side effects only on `allow`
4. capture runpack and initialize regress fixture

Run from repo root:

```bash
uv run --python 3.13 --directory sdk/python python ../../examples/python/reference_adapter_demo.py
```

## 15-Minute Checklist

Stop if any expected field is missing:

- `gate verdict=allow executed=True`
- `runpack run_id=... bundle=...`
- `regress fixture=... config=...`

Decorator examples:

- `sdk/python/examples/openai_style_tool_decorator.py`
- `sdk/python/examples/langchain_style_tool_decorator.py`
