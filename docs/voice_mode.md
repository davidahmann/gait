---
title: "Voice Mode"
description: "Gate high-stakes spoken commitments before they are uttered with signed SayToken capability tokens and callpack artifacts."
---

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

## Frequently Asked Questions

### What is a SayToken?

A SayToken is a signed capability token minted only when gate evaluation returns allow. The voice adapter must hold a valid SayToken before producing gated speech.

### What is a callpack?

A callpack is a signed artifact (like a runpack but for voice calls) containing the event log, commitment intents, gate decisions, speech receipts, and tool effects for a single call.

### Which commitment types are supported?

v1 supports: refund, quote, eligibility, schedule, cancel, and account_change. Each is evaluated against YAML policy before the agent speaks.

### Does voice mode require a live telephony connection?

No. Voice mode is offline-first for verification, diff, and regression. Live gating integrates via a thin adapter at the speech boundary.

### Can I replay a voice call?

Yes. Callpacks support the same verify, diff, and replay workflows as runpacks. You can deterministically replay a gated call using recorded stubs.
