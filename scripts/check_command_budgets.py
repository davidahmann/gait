#!/usr/bin/env python3
from __future__ import annotations

import json
import statistics
import subprocess
import sys
import tempfile
import time
from pathlib import Path
from typing import Any

DEFAULT_BUDGETS_MS = {
    "verify": 1200.0,
    "gate_eval": 1200.0,
    "regress_run": 3000.0,
    "guard_pack": 3000.0,
}


def run_checked(command: list[str], cwd: Path) -> None:
    result = subprocess.run(
        command,
        cwd=cwd,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise RuntimeError(
            f"command failed ({result.returncode}): {' '.join(command)}\n"
            f"stdout:\n{result.stdout}\n"
            f"stderr:\n{result.stderr}"
        )


def measure_median_ms(command: list[str], cwd: Path, repeats: int = 5) -> tuple[float, list[float]]:
    timings_ms: list[float] = []
    for _ in range(repeats):
        start = time.perf_counter()
        run_checked(command, cwd)
        elapsed_ms = (time.perf_counter() - start) * 1000.0
        timings_ms.append(elapsed_ms)
    return float(statistics.median(timings_ms)), timings_ms


def main() -> int:
    if len(sys.argv) != 3:
        print(
            "usage: check_command_budgets.py <gait_binary_path> <report.json>",
            file=sys.stderr,
        )
        return 2

    gait_path = Path(sys.argv[1]).resolve()
    report_path = Path(sys.argv[2])
    if not gait_path.exists():
        print(f"gait binary not found: {gait_path}", file=sys.stderr)
        return 2

    failures: list[str] = []
    report: dict[str, Any] = {"commands": {}, "failures": failures}

    with tempfile.TemporaryDirectory(prefix="gait-budget-") as temp_dir:
        work_dir = Path(temp_dir)
        policy_path = work_dir / "policy.yaml"
        intent_path = work_dir / "intent.json"
        policy_path.write_text(
            "default_verdict: allow\n"
            "rules:\n"
            "  - name: allow-write\n"
            "    effect: allow\n"
            "    match:\n"
            "      tool_names: [tool.write]\n",
            encoding="utf-8",
        )
        intent_path.write_text(
            json.dumps(
                {
                    "schema_id": "gait.gate.intent_request",
                    "schema_version": "1.0.0",
                    "created_at": "2026-02-06T00:00:00Z",
                    "producer_version": "0.0.0-bench",
                    "tool_name": "tool.write",
                    "args": {"path": "/tmp/demo.txt"},
                    "context": {
                        "identity": "alice",
                        "workspace": str(work_dir),
                        "risk_class": "high",
                    },
                }
            )
            + "\n",
            encoding="utf-8",
        )

        run_checked([str(gait_path), "demo", "--json"], work_dir)
        run_checked([str(gait_path), "regress", "init", "--from", "run_demo", "--json"], work_dir)

        command_map = {
            "verify": [str(gait_path), "verify", "run_demo", "--json"],
            "gate_eval": [
                str(gait_path),
                "gate",
                "eval",
                "--policy",
                str(policy_path),
                "--intent",
                str(intent_path),
                "--json",
            ],
            "regress_run": [str(gait_path), "regress", "run", "--json"],
            "guard_pack": [str(gait_path), "guard", "pack", "--run", "run_demo", "--json"],
        }

        for name, command in command_map.items():
            median_ms, samples = measure_median_ms(command, work_dir)
            budget_ms = DEFAULT_BUDGETS_MS[name]
            status = "ok" if median_ms <= budget_ms else "over_budget"
            if status != "ok":
                failures.append(
                    f"{name} exceeded budget: median={median_ms:.1f}ms budget={budget_ms:.1f}ms"
                )
            report["commands"][name] = {
                "command": command,
                "samples_ms": samples,
                "median_ms": median_ms,
                "budget_median_ms": budget_ms,
                "status": status,
            }

    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")

    for name in sorted(report["commands"]):
        entry = report["commands"][name]
        print(
            f"{name}: median={entry['median_ms']:.1f}ms "
            f"budget={entry['budget_median_ms']:.1f}ms "
            f"status={entry['status']}"
        )

    if failures:
        print("command budget check failed:", file=sys.stderr)
        for failure in failures:
            print(f"- {failure}", file=sys.stderr)
        return 1

    print("command budget check passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
