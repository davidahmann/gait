# OpenClaw Skill Package: `gait-gate`

This package is the official OpenClaw installable boundary for Gait OSS.

It routes OpenClaw-style tool call envelopes through `gait mcp proxy` before execution.

## Install (One Command)

From repo root:

```bash
bash scripts/install_openclaw_skill.sh
```

Default install path:

- `~/.openclaw/skills/gait-gate`

## Configure

Set the policy path in:

- `~/.openclaw/skills/gait-gate/skill_config.json`

## Invoke

Use the skill entrypoint:

```bash
python3 ~/.openclaw/skills/gait-gate/gait_openclaw_gate.py \
  --policy examples/policy/base_high_risk.yaml \
  --call /path/to/openclaw_tool_call.json \
  --json
```

The tool exits fail-closed for non-`allow` outcomes and emits deterministic JSON with:

- `verdict`
- `executed` (true only on `allow`)
- `trace_path`
- `reason_codes`
