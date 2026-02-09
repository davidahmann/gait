from __future__ import annotations

import json
from pathlib import Path

from gait import GateEvalResult, IntentRequest, TraceRecord


def _repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


def _fixture(name: str) -> dict[str, object]:
    path = _repo_root() / "core" / "schema" / "testdata" / name
    return json.loads(path.read_text(encoding="utf-8"))


def test_intent_request_fixture_parses_with_sdk_model() -> None:
    payload = _fixture("gate_intent_request_valid.json")
    intent = IntentRequest.from_dict(payload)

    assert intent.schema_id == "gait.gate.intent_request"
    assert intent.tool_name == "tool.demo"
    assert intent.context.identity == "user"
    assert len(intent.targets) == 1
    assert intent.to_dict()["schema_version"] == "1.0.0"


def test_gate_result_fixture_parses_with_sdk_model() -> None:
    payload = _fixture("gate_result_valid.json")
    gate_result = GateEvalResult.from_dict(payload, exit_code=0)

    assert gate_result.ok is False
    assert gate_result.verdict == "allow"
    assert gate_result.reason_codes == []
    assert gate_result.violations == []
    assert gate_result.exit_code == 0


def test_trace_record_fixture_parses_with_sdk_model() -> None:
    payload = _fixture("gate_trace_record_valid.json")
    trace_record = TraceRecord.from_dict(payload)

    assert trace_record.schema_id == "gait.gate.trace"
    assert trace_record.trace_id == "trace_1"
    assert trace_record.verdict == "allow"
    assert trace_record.policy_digest == "3" * 64
