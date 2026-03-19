#!/usr/bin/env python3

from __future__ import annotations

import os
import subprocess
import sys


def main() -> int:
    go_bin = os.environ.get("GO", "go")
    result = subprocess.run(
        [go_bin, "list", "./..."],
        check=True,
        capture_output=True,
        text=True,
    )

    packages = [
        line.strip()
        for line in result.stdout.splitlines()
        if line.strip() and "/node_modules/" not in line
    ]
    if not packages:
        print("no Go packages matched after filtering", file=sys.stderr)
        return 1

    sys.stdout.write("\n".join(packages))
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
