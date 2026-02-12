# Intent And Receipt Spec

This is the compact contract for boundary intent + receipt continuity in Gait.

If your integration claims compatibility, it must:

- emit valid `IntentRequest` payloads at the tool boundary
- preserve receipt continuity in runpack refs (`receipts[]`)
- emit deterministic ticket footer output from `gait run receipt`
- pass conformance checks end to end

Conformance source of truth:

- `docs/contracts/intent_receipt_conformance.md`
- `scripts/test_intent_receipt_conformance.sh`

Run locally:

```bash
bash scripts/test_intent_receipt_conformance.sh ./gait
```
