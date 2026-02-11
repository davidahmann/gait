#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess  # nosec B404
from pathlib import Path
from typing import Any

FRAMEWORK = "autogen"
CREATED_AT = "2026-02-06T00:00:00Z"


def resolve_repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


def resolve_gait_bin(repo_root: Path) -> str:
    configured = os.environ.get("GAIT_BIN", "")
    if configured:
        return configured
    local_binary = repo_root / "gait"
    if local_binary.exists():
        return str(local_binary)
    discovered = shutil.which("gait")
    if discovered:
        return discovered
    raise RuntimeError("unable to find gait binary; set GAIT_BIN or build ./gait")


def build_autogen_tool_call(scenario: str) -> dict[str, Any]:
    arguments = {
        "path": f"/tmp/gait-{FRAMEWORK}-{scenario}.json",
        "content": {"framework": FRAMEWORK, "scenario": scenario},
        "skill": {
            "name": "safe_writer",
            "version": "1.0.0",
            "source": "registry",
            "publisher": "acme",
            "digest": "a" * 64,
            "signature_key_id": "acme-dev-key",
        },
    }
    return {
        "function_call": {
            "name": "write_file",
            "arguments": json.dumps(arguments, separators=(",", ":")),
        }
    }


def to_intent_payload(tool_call: dict[str, Any], scenario: str) -> dict[str, Any]:
    function_call = dict(tool_call["function_call"])
    arguments = json.loads(str(function_call["arguments"]))
    skill = dict(arguments.pop("skill"))
    path = str(arguments["path"])
    return {
        "schema_id": "gait.gate.intent_request",
        "schema_version": "1.0.0",
        "created_at": CREATED_AT,
        "producer_version": "0.0.0-example",
        "tool_name": "tool.write",
        "args": arguments,
        "targets": [
            {
                "kind": "path",
                "value": path,
                "operation": "write",
                "endpoint_class": "fs.write",
            }
        ],
        "arg_provenance": [{"arg_path": "$.path", "source": "user"}],
        "skill_provenance": {
            "skill_name": str(skill["name"]),
            "skill_version": str(skill.get("version", "")),
            "source": str(skill["source"]),
            "publisher": str(skill["publisher"]),
            "digest": str(skill.get("digest", "")),
            "signature_key_id": str(skill.get("signature_key_id", "")),
        },
        "delegation": {
            "requester_identity": "autogen-user",
            "scope_class": "write",
            "token_refs": [f"delegation-{FRAMEWORK}-{scenario}"],
            "chain": [
                {
                    "delegator_identity": "agent.lead",
                    "delegate_identity": "autogen-user",
                    "scope_class": "write",
                    "token_ref": f"delegation-{FRAMEWORK}-{scenario}",
                }
            ],
        },
        "context": {
            "identity": "autogen-user",
            "workspace": "/tmp/gait-autogen",
            "risk_class": "high",
            "session_id": f"sess-{FRAMEWORK}-{scenario}",
            "auth_context": {
                "framework": FRAMEWORK,
                "operator": "example-user",
            },
            "credential_scopes": ["files.write", "network.egress"],
            "environment_fingerprint": f"{FRAMEWORK}:local:{scenario}",
        },
    }


def run_gate_eval(
    gait_bin: str, policy_path: Path, intent_path: Path, trace_path: Path
) -> tuple[int, dict[str, Any]]:
    command = [
        gait_bin,
        "gate",
        "eval",
        "--policy",
        str(policy_path),
        "--intent",
        str(intent_path),
        "--trace-out",
        str(trace_path),
        "--key-mode",
        "dev",
        "--json",
    ]
    completed = subprocess.run(  # nosec B603
        command,
        capture_output=True,
        text=True,
        check=False,
    )
    if not completed.stdout.strip():
        raise RuntimeError(
            f"gait gate eval produced no JSON output (stderr={completed.stderr.strip()})"
        )
    payload = json.loads(completed.stdout)
    if not isinstance(payload, dict):
        raise RuntimeError("gait gate eval returned non-object JSON")
    return completed.returncode, payload


def execute_wrapped_tool(intent_payload: dict[str, Any], output_path: Path) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(
        json.dumps(intent_payload["args"], indent=2) + "\n", encoding="utf-8"
    )


def main() -> int:
    parser = argparse.ArgumentParser(description="AutoGen wrapped tool-call quickstart")
    parser.add_argument("--scenario", choices=["allow", "block"], required=True)
    args = parser.parse_args()

    repo_root = resolve_repo_root()
    gait_bin = resolve_gait_bin(repo_root)
    scenario = args.scenario
    policy_path = Path(__file__).with_name(f"policy_{scenario}.yaml")
    run_dir = repo_root / "gait-out" / "integrations" / FRAMEWORK
    run_dir.mkdir(parents=True, exist_ok=True)
    intent_path = run_dir / f"intent_{scenario}.json"
    trace_path = run_dir / f"trace_{scenario}.json"
    executor_path = run_dir / f"executor_{scenario}.json"

    autogen_call = build_autogen_tool_call(scenario)
    intent_payload = to_intent_payload(autogen_call, scenario)
    intent_path.write_text(
        json.dumps(intent_payload, indent=2) + "\n", encoding="utf-8"
    )

    exit_code, gate_payload = run_gate_eval(
        gait_bin, policy_path, intent_path, trace_path
    )
    verdict = str(gate_payload.get("verdict", "unknown"))
    traced = str(gate_payload.get("trace_path", trace_path))

    if scenario == "allow":
        if exit_code != 0 or verdict != "allow":
            raise RuntimeError(
                f"expected allow flow, got exit_code={exit_code} verdict={verdict}"
            )
        execute_wrapped_tool(intent_payload, executor_path)
        print(f"framework={FRAMEWORK}")
        print("scenario=allow")
        print(f"verdict={verdict}")
        print("executed=true")
        print(f"trace_path={traced}")
        print(f"executor_output={executor_path}")
        return 0

    if verdict != "block" or exit_code not in (0, 3):
        raise RuntimeError(
            f"expected block flow, got exit_code={exit_code} verdict={verdict}"
        )
    print(f"framework={FRAMEWORK}")
    print("scenario=block")
    print(f"verdict={verdict}")
    print("executed=false")
    print(f"trace_path={traced}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except RuntimeError as error:
        print(f"error={error}")
        raise SystemExit(1)
