#!/usr/bin/env python3
from __future__ import annotations

import json
import re
import statistics
import sys
from pathlib import Path
from typing import Any

BENCH_RE = re.compile(r"^(Benchmark[^\s]+)\s+\d+\s+([0-9]+(?:\.[0-9]+)?)\s+ns/op")


def load_baseline(path: Path) -> dict[str, Any]:
    data = json.loads(path.read_text(encoding="utf-8"))
    benchmarks = data.get("benchmarks")
    if not isinstance(benchmarks, dict):
        raise ValueError("baseline file missing 'benchmarks' object")
    return data


def parse_bench_output(path: Path) -> dict[str, list[float]]:
    values: dict[str, list[float]] = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        match = BENCH_RE.match(line.strip())
        if not match:
            continue
        raw_name, ns = match.groups()
        name = re.sub(r"-\d+$", "", raw_name)
        values.setdefault(name, []).append(float(ns))
    return values


def median_ns(values: list[float]) -> float:
    return float(statistics.median(values))


def main() -> int:
    if len(sys.argv) not in (3, 4):
        print(
            "usage: check_bench_regression.py <bench_output.txt> <baseline.json> [report.json]",
            file=sys.stderr,
        )
        return 2

    bench_output_path = Path(sys.argv[1])
    baseline_path = Path(sys.argv[2])
    report_path = Path(sys.argv[3]) if len(sys.argv) == 4 else None

    if not bench_output_path.exists():
        print(f"benchmark output file not found: {bench_output_path}", file=sys.stderr)
        return 2
    if not baseline_path.exists():
        print(f"benchmark baseline file not found: {baseline_path}", file=sys.stderr)
        return 2

    baseline = load_baseline(baseline_path)
    observed = parse_bench_output(bench_output_path)

    benchmark_config: dict[str, dict[str, Any]] = baseline["benchmarks"]
    failures: list[str] = []
    report: dict[str, Any] = {"benchmarks": {}, "failures": failures}

    for name, config in benchmark_config.items():
        baseline_ns = float(config.get("baseline_ns_op", 0.0))
        max_factor = float(config.get("max_regression_factor", 1.0))
        samples = observed.get(name)
        if not samples:
            failures.append(f"missing benchmark in output: {name}")
            continue

        current_ns = median_ns(samples)
        allowed_ns = baseline_ns * max_factor
        status = "ok" if current_ns <= allowed_ns else "regression"
        if status != "ok":
            failures.append(
                f"{name} regression: current={current_ns:.0f}ns/op baseline={baseline_ns:.0f}ns/op factor={max_factor:.2f}"
            )

        report["benchmarks"][name] = {
            "samples_ns_op": samples,
            "current_ns_op_median": current_ns,
            "baseline_ns_op": baseline_ns,
            "max_regression_factor": max_factor,
            "allowed_ns_op": allowed_ns,
            "status": status,
        }

    if report_path is not None:
        report_path.parent.mkdir(parents=True, exist_ok=True)
        report_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")

    for name in sorted(report["benchmarks"]):
        entry = report["benchmarks"][name]
        print(
            f"{name}: median={entry['current_ns_op_median']:.0f}ns/op "
            f"baseline={entry['baseline_ns_op']:.0f}ns/op "
            f"limit={entry['allowed_ns_op']:.0f}ns/op status={entry['status']}"
        )

    if failures:
        print("benchmark regression check failed:", file=sys.stderr)
        for failure in failures:
            print(f"- {failure}", file=sys.stderr)
        return 1

    print("benchmark regression check passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
