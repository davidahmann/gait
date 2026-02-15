#!/usr/bin/env python3
"""Emit a deterministic minimal PackSpec v1 run pack without using gait pack build.

The output is intended for producer-kit interoperability checks.
"""

from __future__ import annotations

import argparse
import copy
import hashlib
import io
import json
import zipfile
from datetime import UTC, datetime
from pathlib import Path

EPOCH_DT = (1980, 1, 1, 0, 0, 0)


def canonical_json(data: object) -> bytes:
    return json.dumps(data, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode("utf-8")


def sha256_hex(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def deterministic_zip_bytes(entries: list[tuple[str, bytes]]) -> bytes:
    buffer = io.BytesIO()
    with zipfile.ZipFile(buffer, mode="w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
        for path, payload in sorted(entries, key=lambda item: item[0]):
            info = zipfile.ZipInfo(path)
            info.date_time = EPOCH_DT
            info.external_attr = (0o644 & 0xFFFF) << 16
            info.create_system = 3
            archive.writestr(info, payload)
    return buffer.getvalue()


def build_source_runpack_stub(run_id: str) -> bytes:
    manifest = {
        "schema_id": "gait.runpack.manifest",
        "schema_version": "1.0.0",
        "created_at": "2026-01-01T00:00:00Z",
        "producer_version": "producer-kit-1.0.0",
        "run_id": run_id,
        "capture_mode": "reference",
        "files": [
            {"path": "run.json", "sha256": sha256_hex(canonical_json({"run_id": run_id}))},
            {"path": "intents.jsonl", "sha256": sha256_hex(b"")},
            {"path": "results.jsonl", "sha256": sha256_hex(b"")},
            {"path": "refs.json", "sha256": sha256_hex(canonical_json({"receipts": []}))},
        ],
        "manifest_digest": "",
    }
    manifest_for_digest = copy.deepcopy(manifest)
    manifest_for_digest["manifest_digest"] = ""
    manifest["manifest_digest"] = sha256_hex(canonical_json(manifest_for_digest))

    run_json = canonical_json({"run_id": run_id})
    refs_json = canonical_json({"receipts": []})
    manifest_json = canonical_json(manifest)
    return deterministic_zip_bytes(
        [
            ("manifest.json", manifest_json),
            ("run.json", run_json),
            ("intents.jsonl", b""),
            ("results.jsonl", b""),
            ("refs.json", refs_json),
        ]
    )


def build_pack(run_id: str, producer_version: str, created_at: str) -> bytes:
    source_runpack = build_source_runpack_stub(run_id)
    run_payload = {
        "schema_id": "gait.pack.run",
        "schema_version": "1.0.0",
        "created_at": created_at,
        "run_id": run_id,
        "capture_mode": "reference",
        "manifest_digest": sha256_hex(source_runpack),
        "intents_count": 0,
        "results_count": 0,
        "refs_count": 0,
    }
    run_payload_bytes = canonical_json(run_payload)

    contents = [
        {"path": "run_payload.json", "sha256": sha256_hex(run_payload_bytes), "type": "json"},
        {"path": "source/runpack.zip", "sha256": sha256_hex(source_runpack), "type": "zip"},
    ]
    manifest = {
        "schema_id": "gait.pack.manifest",
        "schema_version": "1.0.0",
        "created_at": created_at,
        "producer_version": producer_version,
        "pack_id": "",
        "pack_type": "run",
        "source_ref": run_id,
        "contents": contents,
    }

    manifest_for_id = copy.deepcopy(manifest)
    manifest_for_id["pack_id"] = ""
    manifest_for_id.pop("signatures", None)
    manifest["pack_id"] = sha256_hex(canonical_json(manifest_for_id))
    manifest_bytes = canonical_json(manifest)

    return deterministic_zip_bytes(
        [
            ("pack_manifest.json", manifest_bytes),
            ("run_payload.json", run_payload_bytes),
            ("source/runpack.zip", source_runpack),
        ]
    )


def main() -> int:
    parser = argparse.ArgumentParser(description="Emit deterministic minimal PackSpec v1 run pack")
    parser.add_argument("--out", required=True, help="output pack zip path")
    parser.add_argument("--run-id", default="run_producer_kit", help="source run id")
    parser.add_argument("--producer-version", default="producer-kit-1.0.0", help="producer version string")
    parser.add_argument(
        "--created-at",
        default="2026-01-01T00:00:00Z",
        help="RFC3339 timestamp for deterministic pack metadata",
    )
    args = parser.parse_args()

    output = Path(args.out)
    output.parent.mkdir(parents=True, exist_ok=True)
    try:
        datetime.fromisoformat(args.created_at.replace("Z", "+00:00")).astimezone(UTC)
    except ValueError as err:
        raise SystemExit(f"--created-at must be RFC3339: {err}")

    payload = build_pack(
        run_id=args.run_id,
        producer_version=args.producer_version,
        created_at=args.created_at,
    )
    output.write_bytes(payload)

    result = {
        "ok": True,
        "path": str(output),
        "run_id": args.run_id,
        "producer_version": args.producer_version,
        "created_at": args.created_at,
        "sha256": sha256_hex(payload),
    }
    print(json.dumps(result, separators=(",", ":"), sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
