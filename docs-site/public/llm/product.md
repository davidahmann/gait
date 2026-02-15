# Gait Product Summary

Gait is an offline-first runtime for production AI agents that dispatches durable jobs, captures signed evidence, and enforces fail-closed policy at the tool boundary.

It provides seven OSS primitives:

1. **Jobs**: Dispatch multi-step, multi-hour agent work with checkpoints, pause/resume/cancel, approval gates, and deterministic stop reasons.
2. **Packs**: Unified portable artifact envelope (PackSpec v1) for run, job, and call evidence with Ed25519 signatures and SHA-256 manifest.
3. **Gate**: Evaluate structured tool-call intent against YAML policy with fail-closed enforcement. Non-allow outcomes do not execute side effects.
4. **Regress**: Convert any incident or failed run into a deterministic CI regression fixture with JUnit output and stable exit codes.
5. **Voice**: Gate high-stakes spoken commitments (refunds, quotes, eligibility) before they are uttered. Signed SayToken capability tokens and callpack artifacts for voice boundaries.
6. **Context Evidence**: Deterministic proof of what context the model was working from at decision time. Privacy-aware envelopes with fail-closed enforcement when evidence is missing.
7. **Doctor**: Diagnose first-run environment issues with stable JSON output.

Gait is vendor-neutral and offline-first for core workflows: capture, verify, diff, policy evaluation, regressions, and voice/context verification all run without network dependencies.
