#!/usr/bin/env python3
from __future__ import annotations

import argparse
import os
import sys
from pathlib import Path

FRAMEWORK = "langchain"


def resolve_repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


def ensure_sdk_path(repo_root: Path) -> None:
    sdk_root = repo_root / "sdk" / "python"
    if str(sdk_root) not in sys.path:
        sys.path.insert(0, str(sdk_root))


def resolve_gait_bin(repo_root: Path) -> str:
    configured = os.environ.get("GAIT_BIN", "")
    if configured:
        return configured
    local_binary = repo_root / "gait"
    if local_binary.exists():
        return str(local_binary)
    discovered = shutil.which("gait")
    if discovered:
        return discovered
    raise RuntimeError("unable to find gait binary; set GAIT_BIN or build ./gait")


def main() -> int:
    parser = argparse.ArgumentParser(description="LangChain middleware quickstart")
    parser.add_argument(
        "--scenario",
        choices=["allow", "block", "require_approval"],
        required=True,
    )
    args = parser.parse_args()

    repo_root = resolve_repo_root()
    ensure_sdk_path(repo_root)
    try:
        from agent_middleware import run_langchain_scenario
    except ImportError as error:
        if error.name and error.name.startswith("langchain"):
            raise RuntimeError(
                "langchain quickstart requires optional LangChain dependencies; "
                "run `cd sdk/python && uv sync --extra langchain --extra dev` first."
            ) from error
        raise

    gait_bin = resolve_gait_bin(repo_root)
    result = run_langchain_scenario(
        repo_root=repo_root,
        gait_bin=gait_bin,
        scenario=args.scenario,
    )

    for key, value in result.items():
        print(f"{key}={value}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except RuntimeError as error:
        print(f"error={error}")
        raise SystemExit(1)
