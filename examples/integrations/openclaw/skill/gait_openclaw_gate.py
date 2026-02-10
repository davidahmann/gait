#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess  # nosec B404
import sys
from pathlib import Path
from typing import Any


TOOL_MAP: dict[str, str] = {
    "write_file": "tool.write",
    "read_file": "tool.read",
    "delete_file": "tool.delete",
    "shell_command": "tool.exec",
}


def resolve_gait_bin() -> str:
    configured = os.environ.get("GAIT_BIN", "").strip()
    if configured:
        return configured
    discovered = shutil.which("gait")
    if discovered:
        return discovered
    raise RuntimeError("unable to find gait binary; set GAIT_BIN or add gait to PATH")


def read_json(path: str) -> dict[str, Any]:
    if path == "-":
        raw = sys.stdin.read()
    else:
        raw = Path(path).read_text(encoding="utf-8")
    payload = json.loads(raw)
    if not isinstance(payload, dict):
        raise RuntimeError("tool call payload must be a JSON object")
    return payload


def map_openclaw_payload_to_mcp_call(payload: dict[str, Any], identity: str, risk_class: str) -> dict[str, Any]:
    envelope = payload.get("tool_call")
    if not isinstance(envelope, dict):
        raise RuntimeError("expected payload.tool_call object")

    raw_tool = str(envelope.get("tool", "")).strip()
    if not raw_tool:
        raise RuntimeError("expected payload.tool_call.tool")
    mapped_tool = TOOL_MAP.get(raw_tool, f"tool.{raw_tool}")

    params = envelope.get("params")
    if not isinstance(params, dict):
        raise RuntimeError("expected payload.tool_call.params object")

    targets: list[dict[str, str]] = []
    path_value = str(params.get("path", "")).strip()
    if path_value:
        operation = "write"
        if raw_tool == "read_file":
            operation = "read"
        elif raw_tool == "delete_file":
            operation = "delete"
        targets.append(
            {
                "kind": "path",
                "value": path_value,
                "operation": operation,
                "sensitivity": "internal",
            }
        )
    url_value = str(params.get("url", "")).strip()
    if url_value:
        targets.append({"kind": "url", "value": url_value, "operation": "request"})
    host_value = str(params.get("host", "")).strip()
    if host_value:
        targets.append({"kind": "host", "value": host_value, "operation": "request"})

    arg_provenance = [{"arg_path": "$", "source": "user"}]
    call: dict[str, Any] = {
        "name": mapped_tool,
        "args": params,
        "targets": targets,
        "arg_provenance": arg_provenance,
        "context": {
            "identity": identity,
            "workspace": str(Path.cwd()),
            "risk_class": risk_class,
        },
    }
    return call


def run_proxy(
    gait_bin: str,
    policy_path: str,
    mcp_call: dict[str, Any],
    trace_out: str | None,
    runpack_out: str | None,
    key_mode: str,
) -> tuple[int, dict[str, Any]]:
    command = [
        gait_bin,
        "mcp",
        "proxy",
        "--policy",
        policy_path,
        "--call",
        "-",
        "--adapter",
        "mcp",
        "--key-mode",
        key_mode,
        "--json",
    ]
    if trace_out:
        command.extend(["--trace-out", trace_out])
    if runpack_out:
        command.extend(["--runpack-out", runpack_out])

    completed = subprocess.run(  # nosec B603
        command,
        input=json.dumps(mcp_call),
        text=True,
        capture_output=True,
        check=False,
    )
    if not completed.stdout.strip():
        stderr = completed.stderr.strip()
        raise RuntimeError(f"gait mcp proxy returned no JSON output (stderr={stderr})")
    payload = json.loads(completed.stdout)
    if not isinstance(payload, dict):
        raise RuntimeError("gait mcp proxy response must be a JSON object")
    return completed.returncode, payload


def main() -> int:
    parser = argparse.ArgumentParser(description="OpenClaw skill boundary wrapper for gait mcp proxy")
    parser.add_argument("--policy", required=True, help="path to policy YAML")
    parser.add_argument("--call", required=True, help="path to OpenClaw tool-call JSON or '-' for stdin")
    parser.add_argument("--identity", default="openclaw-user", help="identity for intent context")
    parser.add_argument("--risk-class", default="high", help="risk class for intent context")
    parser.add_argument("--trace-out", default="", help="optional trace output path")
    parser.add_argument("--runpack-out", default="", help="optional runpack output path")
    parser.add_argument("--key-mode", default="dev", choices=["dev", "prod"], help="trace signing key mode")
    parser.add_argument("--json", action="store_true", help="emit JSON output")
    args = parser.parse_args()

    gait_bin = resolve_gait_bin()
    payload = read_json(args.call)
    mcp_call = map_openclaw_payload_to_mcp_call(payload, args.identity, args.risk_class)
    exit_code, proxy_payload = run_proxy(
        gait_bin=gait_bin,
        policy_path=args.policy,
        mcp_call=mcp_call,
        trace_out=args.trace_out or None,
        runpack_out=args.runpack_out or None,
        key_mode=args.key_mode,
    )

    verdict = str(proxy_payload.get("verdict", ""))
    result = {
        "ok": bool(proxy_payload.get("ok", False)),
        "framework": "openclaw",
        "verdict": verdict,
        "executed": verdict == "allow",
        "reason_codes": proxy_payload.get("reason_codes", []),
        "violations": proxy_payload.get("violations", []),
        "trace_path": proxy_payload.get("trace_path", ""),
        "runpack_path": proxy_payload.get("runpack_path", ""),
        "intent_digest": proxy_payload.get("intent_digest", ""),
        "policy_digest": proxy_payload.get("policy_digest", ""),
    }

    if args.json:
        print(json.dumps(result, separators=(",", ":"), sort_keys=True))
    else:
        print("framework=openclaw")
        print(f"verdict={result['verdict']}")
        print(f"executed={'true' if result['executed'] else 'false'}")
        print(f"trace_path={result['trace_path']}")

    return exit_code


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except RuntimeError as exc:
        print(json.dumps({"ok": False, "error": str(exc)}, separators=(",", ":"), sort_keys=True))
        raise SystemExit(1)
