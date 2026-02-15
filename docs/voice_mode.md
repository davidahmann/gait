# Voice Mode (v1)

Voice mode adds a non-bypassable commitment boundary before gated speech and emits signed, offline-verifiable call evidence using the same PackSpec pipeline.

## Contract

Voice mode is additive and keeps existing architecture boundaries:

- Gate remains authoritative for policy decisions.
- `CommitmentIntent` is normalized and evaluated by Gate.
- `SayToken` is minted only on `allow` and binds to intent digest, policy digest, call identity, turn, and call sequence.
- Call evidence is emitted as `pack_type=call` and verified/diffed with existing `pack` workflows.
- Regress bootstraps from callpack by extracting embedded source runpack.

## Core Commands

```bash
gait voice token mint --intent <commitment_intent.json> --policy <policy.yaml> --json
gait voice token verify --token <say_token.json> --intent-digest <sha256> --policy-digest <sha256> --json
gait voice pack build --from <call_record.json> --json
gait voice pack verify <callpack.zip> --json
gait voice pack inspect <callpack.zip> --json
gait voice pack diff <left.zip> <right.zip> --json
```

Regress path:

```bash
gait regress bootstrap --from <callpack.zip> --json
```

## Callpack Contents

`pack_type=call` artifacts include:

- `call_payload.json`
- `callpack_manifest.json`
- `call_events.jsonl`
- `commitments.jsonl`
- `gate_decisions.jsonl`
- `speak_receipts.jsonl`
- `reference_digests.json`
- `source/runpack.zip`

`gait voice pack verify` enforces deterministic integrity checks and validates that every speak receipt has a prior `allow` gate decision for the same commitment class and turn.

## Privacy Modes

Supported call payload privacy modes:

- `hash_only` (default)
- `dispute_encrypted`

Privacy mode is explicit in `call_payload.json` and `callpack_manifest.json`, and remains stable under verify/diff workflows.

## Reference Adapter

A thin reference adapter is available at `examples/integrations/voice_reference/`.

Run it with:

```bash
python3 examples/integrations/voice_reference/quickstart.py --scenario allow
python3 examples/integrations/voice_reference/quickstart.py --scenario block
python3 examples/integrations/voice_reference/quickstart.py --scenario require_approval
```

## Test Gates

- `make test-voice-acceptance`
- `make test-adoption`
- CI lane: `voice-acceptance` in `.github/workflows/ci.yml`
