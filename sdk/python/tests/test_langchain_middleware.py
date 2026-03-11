from __future__ import annotations

import importlib
import sys
from dataclasses import dataclass
from pathlib import Path

import pytest

from gait import ToolAdapter, run_session
from gait.adapter import GateEnforcementError

from helpers import create_fake_gait_script, install_fake_langchain_modules


@dataclass(slots=True)
class RuntimeContext:
    identity: str
    workspace: str
    risk_class: str
    run_id: str
    request_id: str
    auth_context: dict[str, str]


@dataclass(slots=True)
class Runtime:
    context: RuntimeContext


def load_langchain_test_module() -> tuple[type[object], object]:
    tool_call_request = install_fake_langchain_modules()
    import gait.langchain as gait_langchain

    return tool_call_request, importlib.reload(gait_langchain)


def make_request(tool_call_request: type[object], *, tool_name: str, scenario: str) -> object:
    return tool_call_request(
        tool_call={
            "name": tool_name,
            "args": {
                "path": f"/tmp/gait-langchain-{scenario}.json",
                "content": {"scenario": scenario},
            },
            "id": f"call_langchain_{scenario}",
        },
        tool=None,
        state={},
        runtime=Runtime(
            context=RuntimeContext(
                identity="agent-langchain",
                workspace="/repo/gait",
                risk_class="high",
                run_id=f"run_langchain_{scenario}",
                request_id=f"req_langchain_{scenario}",
                auth_context={"framework": "langchain"},
            )
        ),
    )


def make_adapter(tmp_path: Path) -> ToolAdapter:
    fake_gait = tmp_path / "fake_gait.py"
    create_fake_gait_script(fake_gait)
    return ToolAdapter(
        policy_path=tmp_path / "policy.yaml",
        gait_bin=[sys.executable, str(fake_gait)],
    )


def test_langchain_allow_executes_and_surfaces_metadata(tmp_path: Path) -> None:
    tool_call_request, gait_langchain = load_langchain_test_module()
    callback = gait_langchain.GaitLangChainCallbackHandler()
    middleware = gait_langchain.GaitLangChainMiddleware(
        make_adapter(tmp_path),
        trace_dir=tmp_path,
        callback_handler=callback,
    )
    request = make_request(tool_call_request, tool_name="tool.allow", scenario="allow")
    calls = {"count": 0}

    def _handler(_: object) -> dict[str, bool]:
        calls["count"] += 1
        return {"ok": True}

    result = middleware.wrap_tool_call(request, _handler)

    assert result == {"ok": True}
    assert calls["count"] == 1
    assert len(callback.events) == 1
    metadata = callback.events[0]
    assert metadata.executed is True
    assert metadata.verdict == "allow"
    assert metadata.run_id == "run_langchain_allow"
    assert metadata.request_id == "req_langchain_allow"
    assert metadata.auth_context == {"framework": "langchain"}
    assert metadata.trace_path is not None
    assert Path(metadata.trace_path).exists()
    assert metadata.policy_digest == "3" * 64
    assert metadata.intent_digest == "2" * 64


@pytest.mark.parametrize(
    ("tool_name", "scenario", "expected_verdict"),
    [
        ("tool.block", "block", "block"),
        ("tool.approval", "require_approval", "require_approval"),
    ],
)
def test_langchain_non_allow_verdicts_fail_closed(
    tmp_path: Path,
    tool_name: str,
    scenario: str,
    expected_verdict: str,
) -> None:
    tool_call_request, gait_langchain = load_langchain_test_module()
    callback = gait_langchain.GaitLangChainCallbackHandler()
    middleware = gait_langchain.GaitLangChainMiddleware(
        make_adapter(tmp_path),
        trace_dir=tmp_path,
        callback_handler=callback,
    )
    request = make_request(tool_call_request, tool_name=tool_name, scenario=scenario)
    calls = {"count": 0}

    def _handler(_: object) -> dict[str, bool]:
        calls["count"] += 1
        return {"ok": True}

    with pytest.raises(GateEnforcementError) as error:
        middleware.wrap_tool_call(request, _handler)

    assert error.value.decision.verdict == expected_verdict
    assert calls["count"] == 0
    assert len(callback.events) == 1
    metadata = callback.events[0]
    assert metadata.executed is False
    assert metadata.verdict == expected_verdict
    assert metadata.trace_path is not None
    assert Path(metadata.trace_path).exists()


def test_langchain_run_session_keeps_run_id_and_trace_metadata(tmp_path: Path) -> None:
    tool_call_request, gait_langchain = load_langchain_test_module()
    middleware = gait_langchain.GaitLangChainMiddleware(
        make_adapter(tmp_path),
        trace_dir=tmp_path,
    )
    request = make_request(tool_call_request, tool_name="tool.allow", scenario="session")

    with run_session(
        run_id="run_langchain_session",
        gait_bin=[sys.executable, str(tmp_path / "fake_gait.py")],
        cwd=tmp_path,
        out_dir=tmp_path / "gait-out",
    ) as session:
        result = middleware.wrap_tool_call(request, lambda _: {"ok": True})

    assert result == {"ok": True}
    assert session.capture is not None
    assert session.capture.run_id == "run_langchain_session"
    assert session.record_input is not None
    assert session.record_input["run"]["run_id"] == "run_langchain_session"
    assert session.attempts[0].verdict == "allow"
    receipts = session.record_input["refs"]["receipts"]
    assert receipts[0]["source_locator"].endswith("tool.allow_call_langchain_session.json")
