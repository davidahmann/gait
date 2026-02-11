#!/usr/bin/env python3
"""Validate docs-site sidebar links resolve to repository docs slugs."""

from __future__ import annotations

import re
import sys
from pathlib import Path


def to_slug(value: str) -> str:
    return value.replace("\\", "/").removesuffix(".md").lower()


def collect_doc_slugs(repo_root: Path) -> set[str]:
    docs_root = repo_root / "docs"
    slugs: set[str] = set()
    for path in docs_root.rglob("*.md"):
        rel = path.relative_to(docs_root)
        slug = to_slug(str(rel))
        if slug == "readme":
            continue
        slugs.add(slug)

    # Root-doc aliases exposed in docs-site/src/lib/docs.ts
    if (repo_root / "SECURITY.md").exists():
        slugs.add("security")
    if (repo_root / "CONTRIBUTING.md").exists():
        slugs.add("contributing")
    if (repo_root / "README.md").exists():
        slugs.add("start-here")
    return slugs


def extract_nav_slugs(navigation_ts: Path) -> list[str]:
    text = navigation_ts.read_text(encoding="utf-8")
    hrefs = re.findall(r"href:\s*'(/docs[^']*)'", text)
    result: list[str] = []
    for href in hrefs:
        if href == "/docs":
            continue
        slug = href.removeprefix("/docs/").rstrip("/").lower()
        if slug:
            result.append(slug)
    return result


def main() -> int:
    repo_root = Path(__file__).resolve().parents[1]
    navigation_path = repo_root / "docs-site" / "src" / "lib" / "navigation.ts"
    if not navigation_path.exists():
        print(f"missing navigation file: {navigation_path}", file=sys.stderr)
        return 2

    known = collect_doc_slugs(repo_root)
    nav_slugs = extract_nav_slugs(navigation_path)
    missing = sorted({slug for slug in nav_slugs if slug not in known})

    if missing:
        print("docs navigation contains unresolved slugs:", file=sys.stderr)
        for slug in missing:
            print(f"  - {slug}", file=sys.stderr)
        return 1

    print(f"docs navigation validated: {len(nav_slugs)} links")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
