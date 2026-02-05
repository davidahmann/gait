#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
from pathlib import Path


def main() -> int:
    if len(sys.argv) != 3:
        print(
            "usage: check_go_coverage.py <coverprofile> <min_percent>", file=sys.stderr
        )
        return 2

    coverprofile = Path(sys.argv[1])
    if not coverprofile.exists():
        print(f"coverage profile not found: {coverprofile}", file=sys.stderr)
        return 2

    try:
        minimum = float(sys.argv[2])
    except ValueError:
        print(f"invalid minimum coverage value: {sys.argv[2]}", file=sys.stderr)
        return 2

    output = subprocess.check_output(
        ["go", "tool", "cover", "-func", str(coverprofile)],
        text=True,
    )
    total_line = ""
    for line in output.splitlines():
        if line.strip().startswith("total:"):
            total_line = line.strip()

    if not total_line:
        print("go coverage total line not found", file=sys.stderr)
        return 1

    percent_text = total_line.split()[-1].rstrip("%")
    try:
        coverage = float(percent_text)
    except ValueError:
        print(f"unable to parse coverage from: {total_line}", file=sys.stderr)
        return 1

    if coverage < minimum:
        print(
            f"Go coverage too low: {coverage:.1f}% (required {minimum:.1f}%)",
            file=sys.stderr,
        )
        return 1

    print(f"Go coverage OK: {coverage:.1f}% (required {minimum:.1f}%)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
