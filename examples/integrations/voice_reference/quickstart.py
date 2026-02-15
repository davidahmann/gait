#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess  # nosec B404
from pathlib import Path
from typing import Any

FRAMEWORK = "voice_reference"
CREATED_AT = "2026-02-15T00:00:00Z"


def resolve_repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


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


def run_json_command(
    command: list[str], *, cwd: Path, expected_exit_codes: tuple[int, ...]
) -> tuple[int, dict[str, Any]]:
    completed = subprocess.run(  # nosec B603
        command,
        cwd=str(cwd),
        capture_output=True,
        text=True,
        check=False,
    )
    if completed.returncode not in expected_exit_codes:
        raise RuntimeError(
            f"command failed: exit={completed.returncode} stderr={completed.stderr.strip()}"
        )
    stdout = completed.stdout.strip()
    if not stdout:
        raise RuntimeError(f"command produced no JSON output: {' '.join(command)}")
    payload = json.loads(stdout)
    if not isinstance(payload, dict):
        raise RuntimeError(f"command returned non-object JSON: {' '.join(command)}")
    return completed.returncode, payload


def resolve_artifact_path(value: str, cwd: Path) -> Path:
    candidate = Path(value)
    if candidate.is_absolute():
        return candidate
    return cwd / candidate


def write_commitment_intent(path: Path, call_id: str) -> dict[str, Any]:
    payload = {
        "schema_id": "gait.voice.commitment_intent",
        "schema_version": "1.0.0",
        "created_at": CREATED_AT,
        "producer_version": "0.0.0-example",
        "call_id": call_id,
        "turn_index": 2,
        "call_seq": 2,
        "commitment_class": "quote",
        "currency": "USD",
        "quote_min_cents": 1200,
        "quote_max_cents": 1800,
        "context": {
            "identity": "voice.agent",
            "workspace": "/srv/voice",
            "risk_class": "high",
            "session_id": f"sess-{call_id}",
            "request_id": f"req-{call_id}",
            "environment_fingerprint": f"{FRAMEWORK}:local",
        },
    }
    path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    return payload


def build_call_record(
    *,
    scenario: str,
    runpack_path: Path,
    call_id: str,
    verdict: str,
    reason_codes: list[str],
    intent_digest: str,
    policy_digest: str,
    say_token_id: str,
) -> dict[str, Any]:
    gated_tts_event: dict[str, Any] = {
        "schema_id": "gait.voice.call_event",
        "schema_version": "1.0.0",
        "created_at": "2026-02-15T00:00:04Z",
        "call_id": call_id,
        "call_seq": 5,
        "turn_index": 2,
        "event_type": "tts.emitted",
    }
    speak_receipts: list[dict[str, Any]] = []
    if scenario == "allow":
        gated_tts_event["commitment_class"] = "quote"
        gated_tts_event["say_token_id"] = say_token_id
        speak_receipts = [
            {
                "call_id": call_id,
                "call_seq": 5,
                "turn_index": 2,
                "commitment_class": "quote",
                "say_token_id": say_token_id,
                "spoken_digest": "2" * 64,
                "emitted_at": "2026-02-15T00:00:04Z",
            }
        ]

    return {
        "schema_id": "gait.voice.call_record",
        "schema_version": "1.0.0",
        "created_at": CREATED_AT,
        "producer_version": "0.0.0-example",
        "call_id": call_id,
        "runpack_path": str(runpack_path),
        "privacy_mode": "hash_only",
        "environment_fingerprint": f"{FRAMEWORK}:local:{scenario}",
        "events": [
            {
                "schema_id": "gait.voice.call_event",
                "schema_version": "1.0.0",
                "created_at": CREATED_AT,
                "call_id": call_id,
                "call_seq": 1,
                "turn_index": 1,
                "event_type": "asr.final",
                "payload_digest": "1" * 64,
            },
            {
                "schema_id": "gait.voice.call_event",
                "schema_version": "1.0.0",
                "created_at": "2026-02-15T00:00:01Z",
                "call_id": call_id,
                "call_seq": 2,
                "turn_index": 2,
                "event_type": "commitment.declared",
                "commitment_class": "quote",
            },
            {
                "schema_id": "gait.voice.call_event",
                "schema_version": "1.0.0",
                "created_at": "2026-02-15T00:00:02Z",
                "call_id": call_id,
                "call_seq": 3,
                "turn_index": 2,
                "event_type": "gate.decision",
                "commitment_class": "quote",
                "intent_digest": intent_digest,
                "policy_digest": policy_digest,
            },
            {
                "schema_id": "gait.voice.call_event",
                "schema_version": "1.0.0",
                "created_at": "2026-02-15T00:00:03Z",
                "call_id": call_id,
                "call_seq": 4,
                "turn_index": 2,
                "event_type": "tts.request",
                "commitment_class": "quote",
            },
            gated_tts_event,
            {
                "schema_id": "gait.voice.call_event",
                "schema_version": "1.0.0",
                "created_at": "2026-02-15T00:00:05Z",
                "call_id": call_id,
                "call_seq": 6,
                "turn_index": 2,
                "event_type": "tool.intent",
            },
            {
                "schema_id": "gait.voice.call_event",
                "schema_version": "1.0.0",
                "created_at": "2026-02-15T00:00:06Z",
                "call_id": call_id,
                "call_seq": 7,
                "turn_index": 2,
                "event_type": "tool.result",
            },
        ],
        "commitments": [
            {
                "schema_id": "gait.voice.commitment_intent",
                "schema_version": "1.0.0",
                "created_at": "2026-02-15T00:00:01Z",
                "producer_version": "0.0.0-example",
                "call_id": call_id,
                "turn_index": 2,
                "call_seq": 2,
                "commitment_class": "quote",
                "currency": "USD",
                "quote_min_cents": 1200,
                "quote_max_cents": 1800,
                "context": {
                    "identity": "voice.agent",
                    "workspace": "/srv/voice",
                    "risk_class": "high",
                    "session_id": f"sess-{call_id}",
                    "request_id": f"req-{call_id}",
                    "environment_fingerprint": f"{FRAMEWORK}:local",
                },
            }
        ],
        "gate_decisions": [
            {
                "call_id": call_id,
                "call_seq": 3,
                "turn_index": 2,
                "commitment_class": "quote",
                "verdict": verdict,
                "reason_codes": reason_codes,
                "intent_digest": intent_digest,
                "policy_digest": policy_digest,
            }
        ],
        "speak_receipts": speak_receipts,
        "reference_digests": [{"ref_id": "kb.quote", "sha256": "3" * 64}],
    }


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Voice reference adapter quickstart with say-token enforcement"
    )
    parser.add_argument(
        "--scenario",
        choices=["allow", "block", "require_approval"],
        required=True,
    )
    args = parser.parse_args()

    scenario = args.scenario
    repo_root = resolve_repo_root()
    gait_bin = resolve_gait_bin(repo_root)
    run_dir = repo_root / "gait-out" / "integrations" / FRAMEWORK
    run_dir.mkdir(parents=True, exist_ok=True)

    _, keys_payload = run_json_command(
        [
            gait_bin,
            "keys",
            "init",
            "--out-dir",
            str(run_dir / "keys"),
            "--prefix",
            "voice",
            "--force",
            "--json",
        ],
        cwd=run_dir,
        expected_exit_codes=(0,),
    )
    private_key_path = str(keys_payload.get("private_key_path", ""))
    if not private_key_path:
        raise RuntimeError("keys init did not return private key path")
    private_key = Path(private_key_path)
    if not private_key.exists():
        raise RuntimeError(f"private key missing: {private_key}")

    _, demo_payload = run_json_command(
        [gait_bin, "demo", "--json"],
        cwd=run_dir,
        expected_exit_codes=(0,),
    )
    bundle = str(demo_payload.get("bundle", ""))
    runpack_path = resolve_artifact_path(bundle, run_dir)
    if not runpack_path.exists():
        raise RuntimeError(f"runpack bundle missing: {runpack_path}")

    call_id = f"call_{FRAMEWORK}_{scenario}"
    intent_path = run_dir / f"intent_{scenario}.json"
    trace_path = run_dir / f"trace_{scenario}.json"
    token_path = run_dir / f"say_token_{scenario}.json"
    call_record_path = run_dir / f"call_record_{scenario}.json"
    callpack_path = run_dir / f"callpack_{scenario}.zip"
    write_commitment_intent(intent_path, call_id)

    policy_path = Path(__file__).with_name(f"policy_{scenario}.yaml")
    token_exit, token_payload = run_json_command(
        [
            gait_bin,
            "voice",
            "token",
            "mint",
            "--intent",
            str(intent_path),
            "--policy",
            str(policy_path),
            "--trace-out",
            str(trace_path),
            "--out",
            str(token_path),
            "--key-mode",
            "prod",
            "--private-key",
            str(private_key),
            "--json",
        ],
        cwd=run_dir,
        expected_exit_codes=(0, 3, 4),
    )
    verdict = str(token_payload.get("verdict", "unknown"))

    if scenario == "allow":
        if token_exit != 0 or verdict != "allow":
            raise RuntimeError(
                f"expected allow scenario, got exit={token_exit} verdict={verdict}"
            )
    elif scenario == "block":
        if token_exit not in (0, 3) or verdict != "block":
            raise RuntimeError(
                f"expected block scenario, got exit={token_exit} verdict={verdict}"
            )
    else:
        if token_exit not in (0, 4) or verdict != "require_approval":
            raise RuntimeError(
                f"expected require_approval scenario, got exit={token_exit} verdict={verdict}"
            )

    intent_digest = str(token_payload.get("intent_digest", ""))
    policy_digest = str(token_payload.get("policy_digest", ""))
    if len(intent_digest) != 64 or len(policy_digest) != 64:
        raise RuntimeError("token mint did not return intent/policy digest")

    reason_codes = token_payload.get("reason_codes", [])
    if not isinstance(reason_codes, list):
        reason_codes = []
    normalized_reason_codes = [str(item) for item in reason_codes]
    say_token_id = str(token_payload.get("token_id", ""))
    speak_emitted = scenario == "allow"
    if speak_emitted:
        if not token_path.exists():
            raise RuntimeError(f"missing say token file: {token_path}")
        _, verify_payload = run_json_command(
            [
                gait_bin,
                "voice",
                "token",
                "verify",
                "--token",
                str(token_path),
                "--private-key",
                str(private_key),
                "--intent-digest",
                intent_digest,
                "--policy-digest",
                policy_digest,
                "--call-id",
                call_id,
                "--turn-index",
                "2",
                "--call-seq",
                "2",
                "--commitment-class",
                "quote",
                "--json",
            ],
            cwd=run_dir,
            expected_exit_codes=(0,),
        )
        if verify_payload.get("ok") is not True:
            raise RuntimeError(f"say token verify failed: {verify_payload}")
    else:
        say_token_id = ""

    call_record = build_call_record(
        scenario=scenario,
        runpack_path=runpack_path,
        call_id=call_id,
        verdict=verdict,
        reason_codes=normalized_reason_codes,
        intent_digest=intent_digest,
        policy_digest=policy_digest,
        say_token_id=say_token_id,
    )
    call_record_path.write_text(json.dumps(call_record, indent=2) + "\n", encoding="utf-8")

    _, build_payload = run_json_command(
        [
            gait_bin,
            "voice",
            "pack",
            "build",
            "--from",
            str(call_record_path),
            "--out",
            str(callpack_path),
            "--json",
        ],
        cwd=run_dir,
        expected_exit_codes=(0,),
    )
    if build_payload.get("ok") is not True:
        raise RuntimeError(f"callpack build failed: {build_payload}")
    _, verify_payload = run_json_command(
        [gait_bin, "voice", "pack", "verify", str(callpack_path), "--json"],
        cwd=run_dir,
        expected_exit_codes=(0,),
    )
    if verify_payload.get("ok") is not True:
        raise RuntimeError(f"callpack verify failed: {verify_payload}")

    print(f"framework={FRAMEWORK}")
    print(f"scenario={scenario}")
    print(f"verdict={verdict}")
    print(f"speak_emitted={'true' if speak_emitted else 'false'}")
    print(f"trace_path={trace_path}")
    print(f"call_record={call_record_path}")
    print(f"callpack_path={callpack_path}")
    if token_path.exists():
        print(f"token_path={token_path}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except RuntimeError as error:
        print(f"error={error}")
        raise SystemExit(1)
