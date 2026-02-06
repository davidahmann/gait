#!/usr/bin/env python3
"""Validate repo-shipped Codex skills under .agents/skills."""

from __future__ import annotations

import re
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
SKILLS_ROOT = REPO_ROOT / ".agents" / "skills"

PROVIDERS = {"codex", "claude", "both"}
REQUIRED_FRONTMATTER_KEYS = {"name", "description"}
CLAUDE_OPTIONAL_FRONTMATTER_KEYS = {
    "disable-model-invocation",
    "argument-hint",
    "user-invocable",
    "allowed-tools",
    "model",
    "agent",
    "context",
    "hooks",
}
FRONTMATTER_ALLOWED_KEYS = REQUIRED_FRONTMATTER_KEYS | CLAUDE_OPTIONAL_FRONTMATTER_KEYS
INTERFACE_REQUIRED_KEYS = {"display_name", "short_description", "default_prompt"}
FRONTMATTER_KEY_VALUE = re.compile(r"^([a-z][a-z0-9_-]*)\s*:\s*(.+?)\s*$")
YAML_KEY_VALUE = re.compile(r"^([a-z][a-z0-9_-]*)\s*:\s*(.+?)\s*$")
SKILL_NAME_PATTERN = re.compile(r"^[a-z0-9][a-z0-9-]{0,63}$")
BOOL_STRINGS = {"true", "false"}


def parse_frontmatter(skill_file: Path) -> tuple[dict[str, str], str, int]:
    raw = skill_file.read_text(encoding="utf-8")
    lines = raw.splitlines()
    if len(lines) < 3 or lines[0].strip() != "---":
        raise ValueError("SKILL.md must start with frontmatter delimiter ---")

    end_index = None
    for index in range(1, len(lines)):
        if lines[index].strip() == "---":
            end_index = index
            break
    if end_index is None:
        raise ValueError("SKILL.md frontmatter closing delimiter --- is missing")

    frontmatter_lines = lines[1:end_index]
    body = "\n".join(lines[end_index + 1 :]).strip()
    if not body:
        raise ValueError("SKILL.md body is empty")

    frontmatter: dict[str, str] = {}
    for line in frontmatter_lines:
        stripped = line.strip()
        if not stripped:
            continue
        match = FRONTMATTER_KEY_VALUE.match(stripped)
        if not match:
            raise ValueError(f"invalid frontmatter line: {line}")
        key = match.group(1)
        value = strip_yaml_scalar(match.group(2))
        frontmatter[key] = value

    return frontmatter, body, len(lines)


def parse_openai_yaml(path: Path) -> dict[str, str]:
    if not path.exists():
        raise ValueError("agents/openai.yaml is missing")

    raw = path.read_text(encoding="utf-8")
    lines = raw.splitlines()
    interface: dict[str, str] = {}
    in_interface = False
    for line in lines:
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue
        if not line.startswith(" "):
            key = stripped.rstrip(":")
            in_interface = key == "interface"
            continue
        if not in_interface:
            continue
        if not line.startswith("  "):
            continue
        nested = line.strip()
        match = YAML_KEY_VALUE.match(nested)
        if not match:
            continue
        key = match.group(1)
        value = strip_yaml_scalar(match.group(2))
        interface[key] = value
    return interface


def strip_yaml_scalar(raw: str) -> str:
    value = raw.strip()
    if len(value) >= 2 and (
        (value.startswith('"') and value.endswith('"'))
        or (value.startswith("'") and value.endswith("'"))
    ):
        return value[1:-1]
    return value


def parse_bool_string(value: str) -> str | None:
    lowered = value.strip().lower()
    if lowered in BOOL_STRINGS:
        return lowered
    return None


def validate_shared_skill_constraints(
    skill_dir: Path, frontmatter: dict[str, str], body: str
) -> list[str]:
    errors: list[str] = []

    unknown_keys = set(frontmatter.keys()) - FRONTMATTER_ALLOWED_KEYS
    if unknown_keys:
        errors.append(
            f"{skill_dir.name}: frontmatter has unsupported keys: {sorted(unknown_keys)}"
        )

    missing_keys = REQUIRED_FRONTMATTER_KEYS - set(frontmatter.keys())
    if missing_keys:
        errors.append(
            f"{skill_dir.name}: frontmatter missing keys: {sorted(missing_keys)}"
        )

    name = frontmatter.get("name", "")
    description = frontmatter.get("description", "")
    if name != skill_dir.name:
        errors.append(f"{skill_dir.name}: frontmatter name must match directory name")
    if name and not SKILL_NAME_PATTERN.fullmatch(name):
        errors.append(f"{skill_dir.name}: name must match ^[a-z0-9][a-z0-9-]{{0,63}}$")
    if not description.strip():
        errors.append(f"{skill_dir.name}: description must be non-empty")

    if "gait " not in body:
        errors.append(f"{skill_dir.name}: body must include at least one gait command")
    if "--json" not in body:
        errors.append(f"{skill_dir.name}: body must require --json output usage")

    return errors


def validate_codex_constraints(
    skill_dir: Path, frontmatter: dict[str, str]
) -> list[str]:
    errors: list[str] = []
    name = frontmatter.get("name", "")
    if "disable-model-invocation" in frontmatter:
        parsed = parse_bool_string(frontmatter["disable-model-invocation"])
        if parsed is None:
            errors.append(
                f"{skill_dir.name}: disable-model-invocation must be true/false when set"
            )

    if "user-invocable" in frontmatter:
        parsed = parse_bool_string(frontmatter["user-invocable"])
        if parsed is None:
            errors.append(
                f"{skill_dir.name}: user-invocable must be true/false when set"
            )

    try:
        interface = parse_openai_yaml(skill_dir / "agents" / "openai.yaml")
    except Exception as exc:  # noqa: BLE001
        errors.append(f"{skill_dir.name}: {exc}")
        return errors

    missing_interface = INTERFACE_REQUIRED_KEYS - set(interface.keys())
    if missing_interface:
        errors.append(
            f"{skill_dir.name}: agents/openai.yaml missing interface keys: {sorted(missing_interface)}"
        )
    short_description = interface.get("short_description", "")
    if short_description and not 25 <= len(short_description) <= 64:
        errors.append(
            f"{skill_dir.name}: short_description length must be 25-64 chars (got {len(short_description)})"
        )
    default_prompt = interface.get("default_prompt", "")
    expected_skill_ref = f"${name}"
    if default_prompt and expected_skill_ref not in default_prompt:
        errors.append(
            f"{skill_dir.name}: default_prompt must include {expected_skill_ref}"
        )

    return errors


def validate_claude_constraints(
    skill_dir: Path, frontmatter: dict[str, str], line_count: int
) -> list[str]:
    errors: list[str] = []

    if line_count > 500:
        errors.append(
            f"{skill_dir.name}: SKILL.md should stay under 500 lines for Claude compatibility"
        )

    disable_value = frontmatter.get("disable-model-invocation")
    disable_bool = None
    if disable_value is not None:
        disable_bool = parse_bool_string(disable_value)
        if disable_bool is None:
            errors.append(
                f"{skill_dir.name}: disable-model-invocation must be true/false"
            )

    user_value = frontmatter.get("user-invocable")
    user_bool = None
    if user_value is not None:
        user_bool = parse_bool_string(user_value)
        if user_bool is None:
            errors.append(f"{skill_dir.name}: user-invocable must be true/false")

    if disable_bool == "true" and user_bool == "false":
        errors.append(
            f"{skill_dir.name}: disable-model-invocation=true and user-invocable=false makes the skill unusable"
        )

    if "allowed-tools" in frontmatter and not frontmatter["allowed-tools"].strip():
        errors.append(f"{skill_dir.name}: allowed-tools cannot be empty when provided")

    return errors


def validate_skill(skill_dir: Path, provider: str) -> list[str]:
    skill_md = skill_dir / "SKILL.md"
    try:
        frontmatter, body, line_count = parse_frontmatter(skill_md)
    except Exception as exc:  # noqa: BLE001
        return [f"{skill_dir.name}: {exc}"]

    errors = validate_shared_skill_constraints(skill_dir, frontmatter, body)
    if errors:
        return errors

    if provider in {"codex", "both"}:
        errors.extend(validate_codex_constraints(skill_dir, frontmatter))

    if provider in {"claude", "both"}:
        errors.extend(validate_claude_constraints(skill_dir, frontmatter, line_count))

    return errors


def parse_provider_from_args(args: list[str]) -> str:
    provider = "both"
    index = 0
    while index < len(args):
        token = args[index]
        if token == "--provider":
            if index + 1 >= len(args):
                raise ValueError("--provider requires a value: codex|claude|both")
            provider = args[index + 1].strip().lower()
            index += 2
            continue
        if token in {"-h", "--help"}:
            print("Usage: validate_repo_skills.py [--provider codex|claude|both]")
            raise SystemExit(0)
        raise ValueError(f"unknown argument: {token}")

    if provider not in PROVIDERS:
        raise ValueError(
            f"invalid provider '{provider}', expected one of: codex, claude, both"
        )

    return provider


def main() -> int:
    try:
        provider = parse_provider_from_args(sys.argv[1:])
    except ValueError as exc:
        print(str(exc), file=sys.stderr)
        return 1
    except SystemExit as exc:
        return int(exc.code)

    if not SKILLS_ROOT.exists():
        print(f"skills root not found: {SKILLS_ROOT}", file=sys.stderr)
        return 1

    skill_dirs = sorted(path for path in SKILLS_ROOT.iterdir() if path.is_dir())
    if not skill_dirs:
        print(f"no skills found under {SKILLS_ROOT}", file=sys.stderr)
        return 1

    all_errors: list[str] = []
    for skill_dir in skill_dirs:
        all_errors.extend(validate_skill(skill_dir, provider))

    if all_errors:
        for error in all_errors:
            print(error, file=sys.stderr)
        return 1

    print(
        f"validated {len(skill_dirs)} skills under {SKILLS_ROOT} for provider={provider}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
