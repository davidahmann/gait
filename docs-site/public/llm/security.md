# Gait Security and Safety

- Fail-closed by default for ambiguous high-risk policy outcomes.
- Structured intent model for policy decisions (not free-form prompt filtering).
- Deterministic and offline verification for all artifact types (runpacks, jobpacks, callpacks).
- Ed25519 signatures and SHA-256 manifest integrity in PackSpec v1.
- Signed traces and explicit reason codes for blocked actions.
- SayToken capability tokens for voice agent commitment gating — gated speech cannot execute without a valid token.
- Context evidence envelopes with fail-closed enforcement when evidence is missing for high-risk actions.
- Durable jobs with deterministic stop reasons and checkpoint integrity.
- No hosted service dependency required for core operation.
