# Troubleshooting Atlas

## `gait gate eval` fails closed

Symptoms:
- Tool never executes.
- Verdict is `error`, `block`, or `require_approval`.

Checks:
1. Validate policy intent pair:
   - `gait policy test <policy> <intent> --json`
2. Verify trace write path permissions.
3. Confirm high-risk mode is expected to fail-closed.

## `govulncheck` network failures

Symptoms:
- Local lint fails in restricted environment.

Fix:
- Use reachable `GOVULNDB` mirror or run in CI with network access.

Reference: `CONTRIBUTING.md`

## Python wrapper issues

Checks:
```bash
(cd sdk/python && PYTHONPATH=. uv run --python 3.13 --extra dev pytest tests/test_adapter.py tests/test_client.py -q)
```

## Adapter parity drift

Run:
```bash
make test-adapter-parity
```

If one adapter diverges, align behavior to common contract in `examples/integrations/README.md`.

## Live connector probe `401`

Cause:
- Invalid or expired provider API key.

Fix:
```bash
GAIT_ENABLE_LIVE_CONNECTOR_TESTS=1 OPENAI_API_KEY='<valid>' ANTHROPIC_API_KEY= bash scripts/test_live_connectors.sh
```
