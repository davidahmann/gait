# Gait Integration Checklist (Epic A3.3)

This checklist is for application teams integrating Gait at the tool-call boundary.

Target: complete in under 30 minutes.

## Scope

This checklist verifies that a repository has the minimum integration needed for deterministic control:

- tool registration boundary control
- fail-closed wrapper or sidecar enforcement
- trace persistence
- runpack recording and verification
- CI regression enforcement

## Prerequisites

- `gait` is available in `PATH`
- Python example dependencies are available (`uv`)
- Repo contains the `examples/` fixtures from this project

## Step 1: Tool Boundary Registration

Required outcome:

- Agents can access only wrapped tools, not raw side-effecting executors.

Validation:

- Confirm your tool registry/factory exports wrapped callables only.
- Confirm every side-effecting tool call path flows through `ToolAdapter.execute(...)`.

Evidence to capture:

- Link to adapter or tool registry file in your repo.

## Step 2: Wrapper Enforcement Semantics

Required outcome:

- Execution occurs only on explicit `allow`.
- `block`, `require_approval`, invalid decision, and evaluation failure do not execute side effects.
- `dry_run` does not execute side effects.

Canonical wrapper path:

- Use `sdk/python/gait/adapter.py` (`ToolAdapter.execute`) as the reference behavior.
- Treat integration-specific examples (`examples/integrations/*`) as adapter-specific usage only.

Validation command:

```bash
uv run --python 3.13 --directory sdk/python python ../../examples/python/reference_adapter_demo.py
```

Evidence to capture:

- Command output showing `gate verdict=allow executed=True` for allow flow.
- Local test output showing fail-closed cases are covered:

```bash
(cd sdk/python && PYTHONPATH=. uv run --python 3.13 --extra dev pytest tests/test_adapter.py tests/test_client.py -q)
```

## Step 2A: Sidecar Enforcement (Non-Python Runtimes)

Required outcome:

- Non-Python runtimes can route tool intents through one sidecar command with fail-closed behavior.

Canonical sidecar path:

- `examples/sidecar/gate_sidecar.py`

Validation commands:

```bash
python3 examples/sidecar/gate_sidecar.py --policy examples/policy-test/allow.yaml --intent-file core/schema/testdata/gate_intent_request_valid.json --trace-out ./gait-out/trace_sidecar_allow.json
python3 examples/sidecar/gate_sidecar.py --policy examples/policy-test/block.yaml --intent-file core/schema/testdata/gate_intent_request_valid.json --trace-out ./gait-out/trace_sidecar_block.json
```

Evidence to capture:

- JSON output includes `gate_result`, `trace_path`, and `exit_code`.
- Blocked or approval-required decisions are treated as non-executable paths.

## Step 3: Gate Trace Persistence

Required outcome:

- Gate decisions produce persisted trace artifacts for audit and replay linkage.

Validation command:

```bash
gait gate eval --policy examples/policy-test/block.yaml --intent examples/policy-test/intent.json --trace-out ./gait-out/trace_check.json --json
```

Evidence to capture:

- `./gait-out/trace_check.json` exists and is non-empty.

## Step 4: Runpack Recording And Verification

Required outcome:

- Team can create and verify deterministic run artifacts locally.

Validation commands:

```bash
gait demo
gait verify run_demo --json
```

Evidence to capture:

- `./gait-out/runpack_run_demo.zip`
- Successful `verify` result.

## Step 5: CI Regression Path

Required outcome:

- At least one deterministic regression fixture runs in CI and can emit JUnit.

Canonical CI path:

- `.github/workflows/adoption-regress-template.yml`
- `docs/ci_regress_kit.md`

Validation commands:

```bash
gait regress init --from run_demo --json
gait regress run --json --junit=./gait-out/junit.xml
```

Evidence to capture:

- `gait.yaml` and `fixtures/` created.
- `./gait-out/junit.xml` generated.

## Step 6: Policy Regression Guard

Required outcome:

- Policy behavior changes are detectable before merge.

Validation commands:

```bash
gait policy test examples/policy-test/allow.yaml examples/policy-test/intent.json --json
gait policy test examples/policy-test/block.yaml examples/policy-test/intent.json --json
gait policy test examples/policy-test/require_approval.yaml examples/policy-test/intent.json --json
```

Evidence to capture:

- Exit codes map to expected decisions (`0`, `3`, `4`).

## Step 7: Integration Sign-Off

Mark integration ready only when all are true:

- Wrapped-tools-only registration is in place.
- Wrapper enforces fail-closed execution semantics.
- Trace persistence is configured.
- Runpack record/verify is reproducible.
- CI runs deterministic regress fixtures.
- Policy tests are part of pre-merge checks.

Recommended release gate:

- Block production rollout if any step above fails.
