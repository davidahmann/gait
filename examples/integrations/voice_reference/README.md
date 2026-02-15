# Voice Reference Adapter (v1)

This adapter demonstrates the voice-boundary contract:

1. normalize a `CommitmentIntent`
2. evaluate policy and mint `SayToken` via `gait voice token mint`
3. fail closed unless verdict is `allow`
4. verify token binding before gated speech
5. emit a deterministic `call_record` and build/verify a signed `callpack`

Run from repo root:

```bash
go build -o ./gait ./cmd/gait
python3 examples/integrations/voice_reference/quickstart.py --scenario allow
python3 examples/integrations/voice_reference/quickstart.py --scenario block
python3 examples/integrations/voice_reference/quickstart.py --scenario require_approval
```

Expected output fields:

- `framework=voice_reference`
- `scenario=<allow|block|require_approval>`
- `verdict=<allow|block|require_approval>`
- `speak_emitted=<true|false>`
- `trace_path=...`
- `call_record=...`
- `callpack_path=...`

Generated artifacts:

- `gait-out/integrations/voice_reference/intent_<scenario>.json`
- `gait-out/integrations/voice_reference/trace_<scenario>.json`
- `gait-out/integrations/voice_reference/call_record_<scenario>.json`
- `gait-out/integrations/voice_reference/callpack_<scenario>.zip`
