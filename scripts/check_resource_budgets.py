#!/usr/bin/env python3
from __future__ import annotations

import json
import re
import statistics
import sys
from pathlib import Path
from typing import Any

BENCH_RE = re.compile(
    r"^(Benchmark[^\s]+)\s+\d+\s+([0-9]+(?:\.[0-9]+)?)\s+ns/op\s+[0-9]+\s+B/op\s+([0-9]+)\s+allocs/op$"
)


def parse_bench_output(path: Path) -> dict[str, dict[str, list[float]]]:
    values: dict[str, dict[str, list[float]]] = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        match = BENCH_RE.match(line.strip())
        if not match:
            continue
        raw_name, ns_op, allocs_op = match.groups()
        name = re.sub(r"-\d+$", "", raw_name)
        entry = values.setdefault(name, {"ns_op": [], "allocs_op": []})
        entry["ns_op"].append(float(ns_op))
        entry["allocs_op"].append(float(allocs_op))
    return values


def main() -> int:
    if len(sys.argv) not in (3, 4):
        print(
            "usage: check_resource_budgets.py <bench_output.txt> <resource_budgets.json> [report.json]",
            file=sys.stderr,
        )
        return 2

    bench_output_path = Path(sys.argv[1])
    budgets_path = Path(sys.argv[2])
    report_path = Path(sys.argv[3]) if len(sys.argv) == 4 else None

    if not bench_output_path.exists():
        print(f"benchmark output file not found: {bench_output_path}", file=sys.stderr)
        return 2
    if not budgets_path.exists():
        print(f"resource budget file not found: {budgets_path}", file=sys.stderr)
        return 2

    budgets = json.loads(budgets_path.read_text(encoding="utf-8"))
    if not isinstance(budgets, dict):
        print("resource budget file must be an object", file=sys.stderr)
        return 2

    observed = parse_bench_output(bench_output_path)
    failures: list[str] = []
    report: dict[str, Any] = {"benchmarks": {}, "failures": failures}

    for name, budget in budgets.items():
        if not isinstance(budget, dict):
            failures.append(f"invalid budget entry for {name}")
            continue
        samples = observed.get(name)
        if not samples:
            failures.append(f"missing benchmark in output: {name}")
            continue

        median_ns = float(statistics.median(samples["ns_op"]))
        median_allocs = float(statistics.median(samples["allocs_op"]))
        max_ns = float(budget.get("max_ns_op", 0.0))
        max_allocs = float(budget.get("max_allocs_op", 0.0))

        status = "ok"
        if median_ns > max_ns:
            status = "over_budget"
            failures.append(
                f"{name} ns/op over budget: median={median_ns:.0f} budget={max_ns:.0f}"
            )
        if median_allocs > max_allocs:
            status = "over_budget"
            failures.append(
                f"{name} allocs/op over budget: median={median_allocs:.0f} budget={max_allocs:.0f}"
            )

        report["benchmarks"][name] = {
            "samples_ns_op": samples["ns_op"],
            "samples_allocs_op": samples["allocs_op"],
            "median_ns_op": median_ns,
            "median_allocs_op": median_allocs,
            "budget_max_ns_op": max_ns,
            "budget_max_allocs_op": max_allocs,
            "status": status,
        }

    if report_path is not None:
        report_path.parent.mkdir(parents=True, exist_ok=True)
        report_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")

    for name in sorted(report["benchmarks"]):
        entry = report["benchmarks"][name]
        print(
            f"{name}: median_ns={entry['median_ns_op']:.0f} budget_ns={entry['budget_max_ns_op']:.0f} "
            f"median_allocs={entry['median_allocs_op']:.0f} budget_allocs={entry['budget_max_allocs_op']:.0f} "
            f"status={entry['status']}"
        )

    if failures:
        print("resource budget check failed:", file=sys.stderr)
        for failure in failures:
            print(f"- {failure}", file=sys.stderr)
        return 1

    print("resource budget check passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
