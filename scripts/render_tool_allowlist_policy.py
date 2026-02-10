#!/usr/bin/env python3
"""Render a deterministic Gait policy from an external tool allowlist JSON."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Generate a gait policy YAML from external tool allowlist JSON. "
            "Supported inputs: {\"tools\": [...]}, [\"tool.a\", ...], or "
            "[{\"tool_name\":\"tool.a\"}, ...]."
        )
    )
    parser.add_argument("--input", required=True, help="path to allowlist JSON")
    parser.add_argument("--output", required=True, help="path to output YAML policy")
    parser.add_argument(
        "--rule-name",
        default="allow_registry_tools",
        help="policy rule name (default: allow_registry_tools)",
    )
    parser.add_argument(
        "--allow-reason-code",
        default="allowed_by_external_registry",
        help="reason code for allow verdict (default: allowed_by_external_registry)",
    )
    return parser.parse_args()


def extract_tools(payload: object) -> list[str]:
    if isinstance(payload, dict):
        tools = payload.get("tools")
        if not isinstance(tools, list):
            raise ValueError("object payload must include a list field 'tools'")
        return normalize_tools(tools)
    if isinstance(payload, list):
        return normalize_tools(payload)
    raise ValueError("payload must be an object or array")


def normalize_tools(values: list[object]) -> list[str]:
    tools: set[str] = set()
    for item in values:
        if isinstance(item, str):
            tool_name = item.strip()
        elif isinstance(item, dict):
            raw_tool_name = item.get("tool_name")
            if not isinstance(raw_tool_name, str):
                raise ValueError("object entries must include string field 'tool_name'")
            tool_name = raw_tool_name.strip()
        else:
            raise ValueError("tool entries must be string or object with tool_name")
        if not tool_name:
            continue
        tools.add(tool_name.lower())
    if not tools:
        raise ValueError("no tool names found in allowlist input")
    return sorted(tools)


def render_policy_yaml(rule_name: str, reason_code: str, tools: list[str]) -> str:
    tool_yaml = ", ".join(tools)
    return (
        "schema_id: gait.gate.policy\n"
        "schema_version: 1.0.0\n"
        "default_verdict: block\n"
        "rules:\n"
        f"  - name: {rule_name}\n"
        "    priority: 10\n"
        "    effect: allow\n"
        "    match:\n"
        f"      tool_names: [{tool_yaml}]\n"
        f"    reason_codes: [{reason_code}]\n"
        "  - name: block_unknown_tools\n"
        "    priority: 100\n"
        "    effect: block\n"
        "    reason_codes: [blocked_tool_not_in_external_registry]\n"
        "    violations: [tool_not_allowlisted]\n"
    )


def main() -> int:
    args = parse_args()
    input_path = Path(args.input).expanduser()
    output_path = Path(args.output).expanduser()
    try:
        payload = json.loads(input_path.read_text(encoding="utf-8"))
        tools = extract_tools(payload)
        rendered = render_policy_yaml(
            rule_name=args.rule_name.strip() or "allow_registry_tools",
            reason_code=args.allow_reason_code.strip() or "allowed_by_external_registry",
            tools=tools,
        )
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_text(rendered, encoding="utf-8")
    except Exception as exc:  # noqa: BLE001
        print(f"render tool allowlist policy failed: {exc}", file=sys.stderr)
        return 2
    print(
        f"rendered policy: tools={len(tools)} input={input_path} output={output_path}",
        file=sys.stdout,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
