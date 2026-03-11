from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

from gait import (
    GaitLangChainCallbackHandler,
    GaitLangChainMiddleware,
    IntentArgProvenance,
    IntentContext,
    IntentTarget,
    ToolAdapter,
    capture_intent,
    run_session,
)
from gait.adapter import GateEnforcementError
from langchain.agents import create_agent
from langchain_core.language_models.chat_models import BaseChatModel
from langchain_core.messages import AIMessage
from langchain_core.outputs import ChatGeneration, ChatResult
from langchain_core.tools import tool

FRAMEWORK = "langchain"


@dataclass(slots=True)
class LangChainRuntimeContext:
    identity: str
    workspace: str
    risk_class: str
    run_id: str
    request_id: str
    auth_context: dict[str, str]


class DeterministicToolModel(BaseChatModel):
    responses: list[AIMessage]

    @property
    def _llm_type(self) -> str:
        return "deterministic-tool-model"

    def bind_tools(self, tools: Any, *, tool_choice: str | None = None, **kwargs: Any) -> Any:
        return self

    def _generate(
        self,
        messages: list[Any],
        stop: list[str] | None = None,
        run_manager: Any | None = None,
        **kwargs: Any,
    ) -> ChatResult:
        del messages, stop, run_manager, kwargs
        if not self.responses:
            raise RuntimeError("deterministic tool model ran out of responses")
        message = self.responses.pop(0)
        return ChatResult(generations=[ChatGeneration(message=message)])


def run_langchain_scenario(
    *,
    repo_root: Path,
    gait_bin: str,
    scenario: str,
) -> dict[str, str]:
    run_dir = repo_root / "gait-out" / "integrations" / FRAMEWORK
    runpack_dir = run_dir / "runpacks"
    run_dir.mkdir(parents=True, exist_ok=True)
    runpack_dir.mkdir(parents=True, exist_ok=True)

    executor_path = run_dir / f"executor_{scenario}.json"
    intent_path = run_dir / f"intent_{scenario}.json"
    trace_path = run_dir / f"trace_{scenario}.json"
    run_id = f"run_{FRAMEWORK}_{scenario}"
    request_id = f"req_{FRAMEWORK}_{scenario}"
    policy_path = Path(__file__).with_name(f"policy_{scenario}.yaml")

    executor_path.unlink(missing_ok=True)
    intent_path.unlink(missing_ok=True)
    trace_path.unlink(missing_ok=True)
    (runpack_dir / f"runpack_{run_id}.zip").unlink(missing_ok=True)

    runtime_context = LangChainRuntimeContext(
        identity="agent-langchain",
        workspace="/tmp/gait-langchain",
        risk_class="high",
        run_id=run_id,
        request_id=request_id,
        auth_context={"framework": FRAMEWORK, "operator": "example-user"},
    )
    callback = GaitLangChainCallbackHandler()
    adapter = ToolAdapter(policy_path=policy_path, gait_bin=gait_bin)
    middleware = GaitLangChainMiddleware(
        adapter,
        intent_factory=lambda request: build_langchain_intent(
            request=request,
            runtime_context=runtime_context,
            intent_path=intent_path,
        ),
        cwd=repo_root,
        trace_path_factory=lambda _request: trace_path,
        callback_handler=callback,
        producer_version="0.0.0-example",
    )
    model = DeterministicToolModel(
        responses=[
            AIMessage(content="", tool_calls=[build_tool_call(scenario)]),
            AIMessage(content=f"{scenario} complete"),
        ]
    )
    agent = create_agent(
        model=model,
        tools=[build_write_tool(executor_path)],
        context_schema=LangChainRuntimeContext,
        middleware=[middleware],
    )

    with run_session(
        run_id=run_id,
        gait_bin=gait_bin,
        cwd=repo_root,
        out_dir=runpack_dir,
    ) as session:
        try:
            result = agent.invoke(
                {"messages": [{"role": "user", "content": "write the file"}]},
                context=runtime_context,
            )
            final_message = str(result["messages"][-1].content)
        except GateEnforcementError as error:
            final_message = str(error)
            if error.decision.verdict is None:
                raise RuntimeError("langchain middleware returned a gate error without verdict") from error

    if session.capture is None:
        raise RuntimeError("langchain middleware example did not emit a runpack")
    if not callback.events:
        raise RuntimeError("langchain middleware example did not emit correlation metadata")

    metadata = callback.events[-1]
    output = {
        "framework": FRAMEWORK,
        "scenario": scenario,
        "verdict": metadata.verdict or "unknown",
        "executed": "true" if metadata.executed else "false",
        "trace_path": metadata.trace_path or str(trace_path),
        "intent_path": str(intent_path),
        "run_id": run_id,
        "request_id": request_id,
        "policy_digest": metadata.policy_digest or "",
        "intent_digest": metadata.intent_digest or "",
        "runpack_path": session.capture.bundle_path,
        "ticket_footer": session.capture.ticket_footer,
        "agent_response": final_message,
    }
    if metadata.executed:
        output["executor_output"] = str(executor_path)
    return output


def build_tool_call(scenario: str) -> dict[str, Any]:
    return {
        "name": "tool.write",
        "args": {
            "path": f"/tmp/gait-{FRAMEWORK}-{scenario}.json",
            "content": {"framework": FRAMEWORK, "scenario": scenario},
        },
        "id": f"call_{FRAMEWORK}_{scenario}",
        "type": "tool_call",
    }


def build_langchain_intent(
    *,
    request: Any,
    runtime_context: LangChainRuntimeContext,
    intent_path: Path | None = None,
) -> Any:
    args = dict(request.tool_call["args"])
    intent = capture_intent(
        tool_name=str(request.tool_call["name"]),
        args=args,
        context=IntentContext(
            identity=runtime_context.identity,
            workspace=runtime_context.workspace,
            risk_class=runtime_context.risk_class,
            session_id=runtime_context.run_id,
            request_id=runtime_context.request_id,
            auth_context=runtime_context.auth_context,
        ),
        targets=[
            IntentTarget(
                kind="path",
                value=str(args["path"]),
                operation="write",
            )
        ],
        arg_provenance=[IntentArgProvenance(arg_path="$.path", source="user")],
        producer_version="0.0.0-example",
    )
    if intent_path is not None:
        intent_path.write_text(json.dumps(intent.to_dict(), indent=2) + "\n", encoding="utf-8")
    return intent


def build_write_tool(executor_path: Path) -> Any:
    @tool("tool.write")
    def write_tool(path: str, content: dict[str, str]) -> str:
        """Write deterministic JSON content after middleware approval."""

        executor_path.parent.mkdir(parents=True, exist_ok=True)
        executor_path.write_text(
            json.dumps({"path": path, "content": content}, indent=2) + "\n",
            encoding="utf-8",
        )
        return str(executor_path)

    return write_tool
