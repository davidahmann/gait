# PLAN Voice Mode (v1): Pre-Utterance Commitment Gating + Signed Callpack Evidence

Date: 2026-02-15  
Scope: Voice mode v1 in the existing Gait repo (additive, offline-first)

## Goal

Make voice agents safe to run in production by enforcing policy before high-stakes speech and producing signed, offline-verifiable artifacts that prove what happened.

## One-Line Product Promise

Gate high-stakes commitments before they are spoken, then emit a signed `callpack` that binds intent, policy decision, evidence references, speech events, and downstream effects.

## Global Decisions (Locked)

- Voice mode ships as an additive surface inside this repo.
- `gate` remains the authoritative policy decision engine.
- `SayToken` mint/verify lives in Go core and is cryptographically signed.
- Voice adapter remains thin and non-authoritative (no policy logic in adapter).
- `verify`, `diff`, and regression workflows remain offline-first and deterministic.
- Privacy defaults remain hash-first; raw transcript/audio remains opt-in only.

## v1 Scope

1. **Pre-Utterance Commitment Gate**
   - Add `CommitmentIntent` support for commitment classes:
     - `refund`, `quote`, `eligibility`, `schedule`, `cancel`, `account_change`
   - Evaluate with existing YAML policy engine.
   - Emit signed trace for all outcomes (`allow`, `block`, `dry_run`, `require_approval`).

2. **Non-Bypassable Speak Boundary**
   - Add `SayToken` capability token minted only on `allow`.
   - Adapter must require valid `SayToken` for gated `tts.request -> tts.emitted`.
   - `SayToken` binds at minimum to:
     - `intent_digest`, `policy_digest`, `call_id`, `turn_index`, `call_seq`, `ttl`
   - Optional bounds:
     - quote ranges, refund ceilings, policy-imposed constraints.

3. **Callpack Artifact**
   - Add `callpack_<call_id>.zip` using existing deterministic pack/sign/verify pipelines.
   - Include canonical event log, commitment intents, decisions, approvals, speech receipts, tool effects, and references.

4. **Canonical Event Model**
   - Append-only event log with monotonic `call_seq`.
   - Required fields per record: `call_id`, `call_seq`, `turn_index`, `created_at`.
   - Required v1 event types:
     - `asr.final`
     - `commitment.declared`
     - `gate.decision`
     - `approval.granted`
     - `tts.request`
     - `tts.emitted`
     - `tool.intent`
     - `tool.result`

5. **Privacy Defaults + Dispute Mode**
   - Default mode stores digests and reference receipts only.
   - Explicit dispute mode stores encrypted transcript/audio artifacts with manifest flags and retention semantics.

6. **Reference Adapter**
   - Ship one voice reference adapter that enforces:
     - commitment-before-speak
     - existing tool-call boundary semantics
   - Adapter only normalizes, calls CLI/core, enforces `SayToken`, and writes events.

7. **Regression Workflow**
   - Support `callpack -> fixtures -> regress run`.
   - Deterministic graders for commitment class correctness, bounds compliance, and approval enforcement.

## Explicit Non-Goals (v1)

- No call-center recording product scope.
- No coaching/QA dashboard scope.
- No speech model hosting scope.
- No broad CCaaS connector suite in v1.
- No mandatory hosted dependency for verify/diff/review.
- No automatic raw-audio commitment extraction by default; v1 expects adapter-declared `CommitmentIntent`.

## Proposed CLI Surface (Additive)

- `gait voice pack build --from <call_log|call_id> [--json]`
- `gait voice pack inspect <callpack.zip> [--json]`
- `gait voice pack verify <callpack.zip> [--json]`
- `gait voice pack diff <left.zip> <right.zip> [--json]`
- `gait voice token verify --token <say_token.json> [--json]`
- `gait regress bootstrap --from <callpack.zip> --name <case> [--json]` (existing command, new input shape)

Note: exact command naming can be adjusted during implementation, but artifact and contract shape are locked.

## Schemas and Contracts (Planned)

New schemas:

- `schemas/v1/voice/commitment_intent.schema.json`
- `schemas/v1/voice/say_token.schema.json`
- `schemas/v1/voice/call_event.schema.json`
- `schemas/v1/voice/callpack_manifest.schema.json`

Existing contract reuse:

- Pack manifest/signature conventions from `pack`/`guard`.
- Gate trace and approval/delegation artifacts from current gate schemas.
- RFC 8785 (JCS) canonicalization and deterministic zip packaging rules.

## Repository Touch Map (Planned)

- `cmd/gait/voice.go` (new command family)
- `cmd/gait/main.go` (command dispatch wiring)
- `core/gate/*` (commitment intent + say token mint/verify integration)
- `core/pack/*` (callpack build/inspect/verify/diff support)
- `core/schema/*` and `core/schema/v1/*` (new voice types + validation)
- `schemas/v1/voice/*` (new schema files)
- `examples/integrations/voice_reference/*` (thin adapter)
- `docs/voice_mode.md` (operator + integrator guide)
- `scripts/test_voice_acceptance.sh` (acceptance gate)
- `Makefile` (`test-voice-acceptance` target)

## Workstreams

## Workstream 0: Contract Freeze

Tasks:

- Freeze commitment classes and reason-code taxonomy.
- Freeze required event model and field semantics.
- Freeze `SayToken` claim set and replay constraints.

Acceptance:

- Contract doc and schema drafts reviewed and locked before coding starts.

## Workstream 1: CommitmentIntent in Gate

Tasks:

- Add commitment intent normalization and digesting.
- Extend policy matching for commitment classes and bounds fields.
- Preserve stable verdict and exit-code behavior.

Acceptance:

- Gate tests cover allow/block/require_approval/dry_run for commitment intents.
- Deterministic digest and reason-code output verified by golden tests.

## Workstream 2: SayToken Capability

Tasks:

- Add token mint path on `allow`.
- Add token verify path with TTL, single-use semantics, and call binding.
- Bind token to `intent_digest`, `policy_digest`, `call_id`, `turn_index`, `call_seq`.

Acceptance:

- Invalid/expired/replayed/mismatched token always blocks gated speech.
- Token verification is deterministic and offline.

## Workstream 3: Callpack Artifact

Tasks:

- Add callpack assembly via deterministic zip pipeline.
- Include canonical event log, intent/decision artifacts, and references.
- Add inspect/verify/diff support.

Acceptance:

- `verify` succeeds on untampered callpack and fails deterministically on byte changes.
- `diff` output is stable on repeated runs for same inputs.

## Workstream 4: Privacy and Dispute Mode

Tasks:

- Add manifest flags for `hash_only` vs `dispute_encrypted`.
- Add encrypted artifact handling and retention metadata.
- Keep default mode free of raw transcript/audio.

Acceptance:

- Manifest flags are unambiguous and validated by schema.
- Verify/diff behavior remains deterministic in both modes.

## Workstream 5: Reference Adapter

Tasks:

- Implement one thin voice adapter example.
- Enforce "no speak without valid say token" for gated classes.
- Emit required events and assemble callpack path.

Acceptance:

- Adapter conformance tests prove fail-closed speech behavior.

## Workstream 6: Regression Workflow

Tasks:

- Extend regress bootstrap to accept callpack input.
- Add deterministic graders for bounds/approval/token violations.

Acceptance:

- Known-good fixtures pass deterministically.
- Known-bad fixtures fail deterministically with stable explanations and exit codes.

## Acceptance Criteria (v1, Testable)

- `AC1 Speak non-bypassable`: zero `tts.emitted` for gated classes without valid `SayToken`.
- `AC2 Fail-closed`: gate failure/unavailability blocks gated speech emission.
- `AC3 Deterministic verify`: tamper fails `verify` deterministically.
- `AC4 Deterministic diff`: identical inputs yield identical stable JSON; changed inputs yield stable minimal diffs.
- `AC5 Timeline reconstructable`: third party can reconstruct commitment -> decision -> approval -> speech -> tool effects from callpack alone.
- `AC6 Regression ready`: callpack bootstraps deterministic regress fixtures and graders.
- `AC7 Privacy explicit`: manifest clearly indicates capture mode and expected behavior.

## Validation Gates (Implementation Phase)

Minimum required per PR slice:

- `make lint-fast`
- `make test-fast`
- `go test ./core/gate ./core/pack ./core/schema/validate ./cmd/gait`
- `bash scripts/test_voice_acceptance.sh` (new)
- `make test-adoption` (ensure no regression in existing lane)
- `make test-adapter-parity` (ensure no regression in adapter contracts)

## Phase Sequence

1. Contract freeze + schemas.
2. Gate + SayToken logic.
3. Callpack build/verify/diff.
4. Reference adapter.
5. Regress integration.
6. End-to-end acceptance + docs.

## Risks and Mitigations

- **Risk:** Speech chunking semantics create bypass gaps.
  - **Mitigation:** define strict event ordering and token validity across chunked emissions.
- **Risk:** Token replay in distributed runtime paths.
  - **Mitigation:** single-use nonce/call_seq binding and deterministic replay checks.
- **Risk:** Privacy mode confusion.
  - **Mitigation:** explicit manifest flags plus schema validation and CLI warnings.
- **Risk:** Scope creep into call-center platform features.
  - **Mitigation:** enforce non-goals at review gates and release checklist.

## Definition of Done (v1 Voice Mode)

- Commitment gating before speech is enforceable and fail-closed.
- `SayToken` is required and non-bypassable for gated commitment speech.
- Signed callpack artifacts verify and diff offline deterministically.
- Privacy posture is explicit with stable behavior in both modes.
- Regress can bootstrap and run from callpacks with deterministic pass/fail output.
