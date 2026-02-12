from __future__ import annotations

from gait import IntentContext, ToolAdapter, gate_tool

adapter = ToolAdapter(policy_path="gait.policy.yaml", gait_bin="gait")


@gate_tool(
    adapter=adapter,
    context=IntentContext(identity="agent-openai", workspace="/srv/agent", risk_class="high"),
    tool_name="tool.write",
    trace_out="./gait-out/trace_openai_write.json",
)
def write_customer_note(customer_id: str, note: str) -> dict[str, str]:
    # Side effects run only on explicit allow; non-allow raises GateEnforcementError.
    return {"customer_id": customer_id, "status": "written"}


if __name__ == "__main__":
    result = write_customer_note("cust_123", "follow up in 24h")
    print(result)
