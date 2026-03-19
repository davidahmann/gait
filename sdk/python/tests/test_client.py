from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path
from typing import Sequence

import pytest

from gait import client as client_module
from gait import (
    IntentContext,
    IntentRequest,
    IntentScript,
    IntentScriptStep,
    IntentTarget,
    capture_demo_runpack,
    capture_intent,
    create_regress_fixture,
    evaluate_gate,
    record_runpack,
    write_trace,
)

from helpers import create_fake_gait_script


def test_capture_intent_and_evaluate_gate_allow(tmp_path: Path) -> None:
    fake_gait = tmp_path / "fake_gait.py"
    create_fake_gait_script(fake_gait)

    intent = capture_intent(
        tool_name="tool.allow",
        args={"path": "/tmp/out.txt"},
        targets=[IntentTarget(kind="path", value="/tmp/out.txt")],
        context=IntentContext(identity="alice", workspace="/repo/gait", risk_class="high"),
    )
    result = evaluate_gate(
        policy_path=tmp_path / "policy.yaml",
        intent=intent,
        gait_bin=[sys.executable, str(fake_gait)],
        cwd=tmp_path,
    )

    assert isinstance(intent, IntentRequest)
    assert result.exit_code == 0
    assert result.ok
    assert result.verdict == "allow"
    assert result.reason_codes == ["default_allow"]
    assert result.trace_path is not None
    assert result.trace_path.endswith("trace_fake.json")


def test_evaluate_gate_require_approval_exit_code(tmp_path: Path) -> None:
    fake_gait = tmp_path / "fake_gait.py"
    create_fake_gait_script(fake_gait)

    intent = capture_intent(
        tool_name="tool.approval",
        args={"path": "/tmp/out.txt"},
        context=IntentContext(identity="alice", workspace="/repo/gait", risk_class="high"),
    )
    result = evaluate_gate(
        policy_path=tmp_path / "policy.yaml",
        intent=intent,
        gait_bin=[sys.executable, str(fake_gait)],
        cwd=tmp_path,
    )

    assert result.exit_code == 4
    assert result.verdict == "require_approval"
    assert result.reason_codes == ["approval_required"]


def test_evaluate_gate_block_exit_code(tmp_path: Path) -> None:
    fake_gait = tmp_path / "fake_gait.py"
    create_fake_gait_script(fake_gait)

    intent = capture_intent(
        tool_name="tool.block",
        args={"path": "/tmp/out.txt"},
        context=IntentContext(identity="alice", workspace="/repo/gait", risk_class="high"),
    )
    result = evaluate_gate(
        policy_path=tmp_path / "policy.yaml",
        intent=intent,
        gait_bin=[sys.executable, str(fake_gait)],
        cwd=tmp_path,
    )

    assert result.exit_code == 3
    assert result.verdict == "block"
    assert result.reason_codes == ["blocked_tool"]


def test_capture_intent_script_payload() -> None:
    intent = capture_intent(
        tool_name="script",
        args={},
        context=IntentContext(identity="alice", workspace="/repo/gait", risk_class="high"),
        script=IntentScript(
            steps=[
                IntentScriptStep(tool_name="tool.read", args={"path": "/tmp/input.txt"}),
                IntentScriptStep(tool_name="tool.write", args={"path": "/tmp/output.txt"}),
            ]
        ),
    )

    payload = intent.to_dict()
    assert intent.script is not None
    assert payload["tool_name"] == "script"
    assert payload["script"]["steps"][0]["tool_name"] == "tool.read"
    assert payload["script"]["steps"][1]["tool_name"] == "tool.write"


def test_write_trace_copies_source_record(tmp_path: Path) -> None:
    source = tmp_path / "trace.json"
    trace_payload = {
        "schema_id": "gait.gate.trace",
        "schema_version": "1.0.0",
        "created_at": "2026-02-05T00:00:00Z",
        "producer_version": "0.0.0-dev",
        "trace_id": "trace_123",
        "tool_name": "tool.write",
        "args_digest": "1" * 64,
        "intent_digest": "2" * 64,
        "policy_digest": "3" * 64,
        "verdict": "allow",
    }
    source.write_text(json.dumps(trace_payload, indent=2) + "\n", encoding="utf-8")

    destination = tmp_path / "out" / "trace.json"
    written = write_trace(trace_path=source, destination_path=destination)

    assert written == destination
    assert destination.read_text(encoding="utf-8") == source.read_text(encoding="utf-8")


def test_capture_demo_and_create_regress_fixture(tmp_path: Path) -> None:
    fake_gait = tmp_path / "fake_gait.py"
    create_fake_gait_script(fake_gait)

    demo = capture_demo_runpack(gait_bin=[sys.executable, str(fake_gait)], cwd=tmp_path)
    assert demo.run_id == "run_demo"
    assert demo.verified
    assert demo.bundle_path == "./gait-out/runpack_run_demo.zip"

    fixture = create_regress_fixture(
        from_run=demo.run_id,
        gait_bin=[sys.executable, str(fake_gait)],
        cwd=tmp_path,
    )
    assert fixture.run_id == "run_demo"
    assert fixture.fixture_name == "run_demo"
    assert fixture.runpack_path == "fixtures/run_demo/runpack.zip"


def test_capture_demo_runpack_uses_json_cli_contract(tmp_path: Path) -> None:
    observed_commands: list[list[str]] = []

    def fake_run(command: Sequence[str], cwd: object = None) -> client_module._CommandResult:
        observed_commands.append(list(command))
        return client_module._CommandResult(
            command=list(command),
            exit_code=0,
            stdout=json.dumps(
                {
                    "ok": True,
                    "run_id": "run_demo",
                    "bundle": "./gait-out/runpack_run_demo.zip",
                    "ticket_footer": "GAIT run_id=run_demo manifest=sha256:abc",
                    "verify": "ok",
                }
            ),
            stderr="",
        )

    with pytest.MonkeyPatch.context() as monkeypatch:
        monkeypatch.setattr(client_module, "_run_command", fake_run)
        demo = capture_demo_runpack(gait_bin="gait", cwd=tmp_path)

    assert demo.run_id == "run_demo"
    assert observed_commands == [["gait", "demo", "--json"]]


def test_record_runpack_round_trip(tmp_path: Path) -> None:
    fake_gait = tmp_path / "fake_gait.py"
    create_fake_gait_script(fake_gait)
    capture_path = tmp_path / "captured_record_input.json"

    result = record_runpack(
        record_input={
            "run": {
                "schema_id": "gait.runpack.run",
                "schema_version": "1.0.0",
                "created_at": "2026-02-12T00:00:00Z",
                "producer_version": "0.0.0-dev",
                "run_id": "run_sdk",
                "env": {"os": "darwin", "arch": "arm64", "runtime": "python3.11"},
                "timeline": [{"event": "run_started", "ts": "2026-02-12T00:00:00Z"}],
            },
            "intents": [],
            "results": [],
            "refs": {
                "schema_id": "gait.runpack.refs",
                "schema_version": "1.0.0",
                "created_at": "2026-02-12T00:00:00Z",
                "producer_version": "0.0.0-dev",
                "run_id": "run_sdk",
                "receipts": [],
            },
            "capture_mode": "reference",
        },
        gait_bin=[sys.executable, str(fake_gait)],
        cwd=tmp_path,
        out_dir=tmp_path / "gait-out",
    )

    assert result.run_id == "run_sdk"
    assert result.bundle_path.endswith("runpack_run_sdk.zip")
    assert result.manifest_digest == "4" * 64
    assert result.ticket_footer.startswith("GAIT run_id=run_sdk")

    with pytest.MonkeyPatch.context() as monkeypatch:
        monkeypatch.setenv("FAKE_GAIT_RECORD_CAPTURE", str(capture_path))
        record_runpack(
            record_input={
                "run": {
                    "schema_id": "gait.runpack.run",
                    "schema_version": "1.0.0",
                    "created_at": "2026-02-12T00:00:00Z",
                    "producer_version": "0.0.0-dev",
                    "run_id": "run_capture",
                    "env": {"os": "darwin", "arch": "arm64", "runtime": "python3.11"},
                    "timeline": [{"event": "run_started", "ts": "2026-02-12T00:00:00Z"}],
                },
                "intents": [],
                "results": [],
                "refs": {
                    "schema_id": "gait.runpack.refs",
                    "schema_version": "1.0.0",
                    "created_at": "2026-02-12T00:00:00Z",
                    "producer_version": "0.0.0-dev",
                    "run_id": "run_capture",
                    "receipts": [],
                },
                "capture_mode": "reference",
            },
            gait_bin=[sys.executable, str(fake_gait)],
            cwd=tmp_path,
            out_dir=tmp_path / "gait-out",
        )

    captured = json.loads(capture_path.read_text(encoding="utf-8"))
    assert captured["run"]["run_id"] == "run_capture"


def test_evaluate_gate_with_all_optional_key_flags(tmp_path: Path) -> None:
    fake_gait = tmp_path / "fake_gait.py"
    create_fake_gait_script(fake_gait)

    intent = capture_intent(
        tool_name="tool.allow",
        args={"path": "/tmp/out.txt"},
        context=IntentContext(identity="alice", workspace="/repo/gait", risk_class="high"),
    )
    result = evaluate_gate(
        policy_path=tmp_path / "policy.yaml",
        intent=intent,
        gait_bin=[sys.executable, str(fake_gait)],
        cwd=tmp_path,
        trace_out=tmp_path / "trace.json",
        approval_token=tmp_path / "approval.json",
        private_key=tmp_path / "private.key",
        private_key_env="GAIT_PRIVATE_KEY",
        approval_public_key=tmp_path / "approval.pub",
        approval_public_key_env="GAIT_APPROVAL_PUBLIC_KEY",
        approval_private_key=tmp_path / "approval.key",
        approval_private_key_env="GAIT_APPROVAL_PRIVATE_KEY",
    )

    assert result.ok
    assert result.exit_code == 0


def test_evaluate_gate_with_delegation_and_delegation_key_flags(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    observed_commands: list[list[str]] = []

    def fake_run(command: Sequence[str], cwd: object = None) -> client_module._CommandResult:
        observed_commands.append(list(command))
        return client_module._CommandResult(
            command=list(command),
            exit_code=0,
            stdout=json.dumps(
                {
                    "ok": True,
                    "verdict": "allow",
                    "reason_codes": ["default_allow"],
                    "trace_id": "trace_1",
                    "trace_path": "trace.json",
                    "policy_digest": "p" * 64,
                    "intent_digest": "i" * 64,
                }
            ),
            stderr="",
        )

    intent = capture_intent(
        tool_name="tool.allow",
        args={"path": "/tmp/out.txt"},
        context=IntentContext(identity="alice", workspace="/repo/gait", risk_class="high"),
    )

    with pytest.MonkeyPatch.context() as patch:
        patch.setattr(client_module, "_run_command", fake_run)
        result = evaluate_gate(
            policy_path=tmp_path / "policy.yaml",
            intent=intent,
            gait_bin="gait",
            cwd=tmp_path,
            delegation_token=tmp_path / "delegation.json",
            delegation_token_chain=[tmp_path / "delegation-a.json", tmp_path / "delegation-b.json"],
            delegation_public_key=tmp_path / "delegation.pub",
            delegation_public_key_env="GAIT_DELEGATION_PUBLIC_KEY",
            delegation_private_key=tmp_path / "delegation.key",
            delegation_private_key_env="GAIT_DELEGATION_PRIVATE_KEY",
        )

    assert result.ok
    assert len(observed_commands) == 1
    command = observed_commands[0]
    assert command[:4] == ["gait", "gate", "eval", "--policy"]
    assert str(tmp_path / "policy.yaml") in command
    assert "--intent" in command
    assert "--delegation-token" in command
    assert str(tmp_path / "delegation.json") in command
    assert "--delegation-token-chain" in command
    assert f"{tmp_path / 'delegation-a.json'},{tmp_path / 'delegation-b.json'}" in command
    assert "--delegation-public-key" in command
    assert str(tmp_path / "delegation.pub") in command
    assert "--delegation-public-key-env" in command
    assert "GAIT_DELEGATION_PUBLIC_KEY" in command
    assert "--delegation-private-key" in command
    assert str(tmp_path / "delegation.key") in command
    assert "--delegation-private-key-env" in command
    assert "GAIT_DELEGATION_PRIVATE_KEY" in command


def test_evaluate_gate_non_verdict_error_raises_command_error(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    intent = capture_intent(
        tool_name="tool.allow",
        args={"path": "/tmp/out.txt"},
        context=IntentContext(identity="alice", workspace="/repo/gait", risk_class="high"),
    )

    monkeypatch.setattr(
        client_module,
        "_run_command",
        lambda command, cwd=None: client_module._CommandResult(
            command=list(command),
            exit_code=1,
            stdout=json.dumps({"ok": False, "error": "gate exploded"}),
            stderr="boom",
        ),
    )

    with pytest.raises(client_module.GaitCommandError) as raised:
        evaluate_gate(policy_path=tmp_path / "policy.yaml", intent=intent, gait_bin="gait")
    assert "gate exploded" in str(raised.value)


def test_internal_helpers_parse_json_and_prefix() -> None:
    assert client_module._parse_json_stdout("") is None
    assert client_module._parse_json_stdout("[]") is None
    assert client_module._parse_json_stdout('{"ok": true}') == {"ok": True}
    assert client_module._command_prefix("gait") == ["gait"]
    assert client_module._command_prefix(["python", "script.py"]) == ["python", "script.py"]


def test_run_command_timeout_raises_command_error(monkeypatch: pytest.MonkeyPatch) -> None:
    def timeout_run(*args: object, **kwargs: object) -> object:
        raise subprocess.TimeoutExpired(cmd=["gait", "demo"], timeout=0.01)

    monkeypatch.setattr(client_module.subprocess, "run", timeout_run)
    with pytest.raises(client_module.GaitCommandError) as raised:
        client_module._run_command(["gait", "demo"], cwd=None)
    assert "timed out" in str(raised.value)


def test_run_command_binary_not_found_raises_actionable_error(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    def missing_binary(*args: object, **kwargs: object) -> object:
        raise FileNotFoundError("No such file or directory: gait")

    monkeypatch.setattr(client_module.subprocess, "run", missing_binary)
    with pytest.raises(client_module.GaitCommandError) as raised:
        client_module._run_command(["gait", "demo"], cwd=None)
    assert "binary not found" in str(raised.value)
    assert "gait_bin" in str(raised.value)
    assert raised.value.exit_code == 127


def test_run_command_missing_cwd_raises_path_error(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    missing_cwd = tmp_path / "missing-cwd"

    def missing_cwd_run(*args: object, **kwargs: object) -> object:
        raise FileNotFoundError(2, "No such file or directory", str(missing_cwd))

    monkeypatch.setattr(client_module.subprocess, "run", missing_cwd_run)
    with pytest.raises(client_module.GaitCommandError) as raised:
        client_module._run_command(["gait", "demo"], cwd=missing_cwd)
    assert "cwd not found" in str(raised.value)
    assert "binary not found" not in str(raised.value)
    assert str(missing_cwd) in str(raised.value)
    assert raised.value.exit_code == -1


def test_evaluate_gate_malformed_json_raises_command_error(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    intent = capture_intent(
        tool_name="tool.allow",
        args={"path": "/tmp/out.txt"},
        context=IntentContext(identity="alice", workspace="/repo/gait", risk_class="high"),
    )

    monkeypatch.setattr(
        client_module,
        "_run_command",
        lambda command, cwd=None: client_module._CommandResult(
            command=list(command),
            exit_code=0,
            stdout="not-json",
            stderr="",
        ),
    )

    with pytest.raises(client_module.GaitCommandError) as raised:
        evaluate_gate(policy_path=tmp_path / "policy.yaml", intent=intent, gait_bin="gait")
    assert "failed to parse JSON" in str(raised.value)


def test_create_regress_fixture_malformed_json_raises_command_error(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    monkeypatch.setattr(
        client_module,
        "_run_command",
        lambda command, cwd=None: client_module._CommandResult(
            command=list(command),
            exit_code=0,
            stdout="{}[]",
            stderr="",
        ),
    )

    with pytest.raises(client_module.GaitCommandError) as raised:
        create_regress_fixture(from_run="run_demo", gait_bin="gait", cwd=tmp_path)
    assert "failed to parse JSON" in str(raised.value)


def test_capture_demo_runpack_malformed_json_raises_command_error(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    monkeypatch.setattr(
        client_module,
        "_run_command",
        lambda command, cwd=None: client_module._CommandResult(
            command=list(command),
            exit_code=0,
            stdout="run_id=run_demo\nbundle=./gait-out/runpack_run_demo.zip\n",
            stderr="",
        ),
    )

    with pytest.raises(client_module.GaitCommandError) as raised:
        capture_demo_runpack(gait_bin="gait", cwd=tmp_path)
    assert "failed to parse JSON" in str(raised.value)


def test_record_runpack_invalid_capture_mode_raises(tmp_path: Path) -> None:
    with pytest.raises(client_module.GaitError):
        record_runpack(
            record_input={},
            gait_bin="gait",
            cwd=tmp_path,
            capture_mode="invalid",
        )


def test_write_trace_rejects_missing_file_and_wrong_schema(tmp_path: Path) -> None:
    with pytest.raises(client_module.GaitError) as missing:
        write_trace(trace_path=tmp_path / "missing.json", destination_path=tmp_path / "out.json")
    assert "trace file not found" in str(missing.value)

    wrong_schema = tmp_path / "wrong-trace.json"
    wrong_schema.write_text(
        json.dumps(
            {
                "schema_id": "gait.not_trace",
                "schema_version": "1.0.0",
                "created_at": "2026-02-05T00:00:00Z",
                "producer_version": "0.0.0-dev",
                "trace_id": "trace_1",
                "tool_name": "tool.write",
                "args_digest": "1" * 64,
                "intent_digest": "2" * 64,
                "policy_digest": "3" * 64,
                "verdict": "allow",
            }
        )
        + "\n",
        encoding="utf-8",
    )

    with pytest.raises(client_module.GaitError) as wrong:
        write_trace(trace_path=wrong_schema, destination_path=tmp_path / "out.json")
    assert "unexpected trace schema_id" in str(wrong.value)


def test_capture_demo_runpack_error_paths(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr(
        client_module,
        "_run_command",
        lambda command, cwd=None: client_module._CommandResult(
            command=list(command),
            exit_code=1,
            stdout=json.dumps({"ok": False, "error": "demo failed"}),
            stderr="",
        ),
    )
    with pytest.raises(client_module.GaitCommandError) as raised:
        capture_demo_runpack(gait_bin="gait", cwd=tmp_path)
    assert "demo failed" in str(raised.value)

    monkeypatch.setattr(
        client_module,
        "_run_command",
        lambda command, cwd=None: client_module._CommandResult(
            command=list(command),
            exit_code=0,
            stdout=json.dumps({"ok": False}),
            stderr="",
        ),
    )
    with pytest.raises(client_module.GaitError) as ok_false:
        capture_demo_runpack(gait_bin="gait", cwd=tmp_path)
    assert "ok=false" in str(ok_false.value)

    monkeypatch.setattr(
        client_module,
        "_run_command",
        lambda command, cwd=None: client_module._CommandResult(
            command=list(command),
            exit_code=0,
            stdout=json.dumps({"ok": True, "run_id": "", "bundle": ""}),
            stderr="",
        ),
    )
    with pytest.raises(client_module.GaitError) as missing_fields:
        capture_demo_runpack(gait_bin="gait", cwd=tmp_path)
    assert "missing run_id or bundle" in str(missing_fields.value)


def test_create_regress_fixture_error_paths(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr(
        client_module,
        "_run_command",
        lambda command, cwd=None: client_module._CommandResult(
            command=list(command),
            exit_code=1,
            stdout=json.dumps({"ok": False, "error": "regress init failed"}),
            stderr="",
        ),
    )
    with pytest.raises(client_module.GaitCommandError) as raised:
        create_regress_fixture(from_run="run_demo", gait_bin="gait", cwd=tmp_path)
    assert "regress init failed" in str(raised.value)

    monkeypatch.setattr(
        client_module,
        "_run_command",
        lambda command, cwd=None: client_module._CommandResult(
            command=list(command),
            exit_code=0,
            stdout=json.dumps({"ok": False}),
            stderr="",
        ),
    )
    with pytest.raises(client_module.GaitError) as ok_false:
        create_regress_fixture(from_run="run_demo", gait_bin="gait", cwd=tmp_path)
    assert "ok=false" in str(ok_false.value)
