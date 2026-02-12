# Launch KPI Scorecard (v2.3)

Track these metrics for every launch window (day 1, day 7, day 30).

## Activation Metrics (M*)

- `M1`: median install-to-`gait demo` success <= 5 minutes
- `M2`: median install-to-`gait regress run` success <= 15 minutes
- `M3`: wrapper quickstart completion rate >= 90%
- `M4`: deterministic CI regress template pass rate >= 95%

## Conformance Metrics (C*)

- `C1`: Intent schema conformance checks pass 100%
- `C2`: receipt/ticket-footer verification contract passes 100%
- `C3`: additive-field compatibility checks pass 100% for enterprise consumer projections

## Distribution Metrics (D*)

- `D1`: official skill install success >= 95% (`codex`, `claude`)
- `D2`: skill workflow end-to-end checks pass with deterministic outputs
- `D3`: reusable CI template validates in both dispatch + workflow-call mode

## Blessed Lane Scorecard

Lane candidates:

- `coding_agent_local`
- `ci_workflow`
- `it_workflow`

Weighted decision factors:

- setup time (25%)
- failure rate (25%)
- determinism pass rate (20%)
- policy correctness (20%)
- distribution reach (10%)

Generate decision artifact:

```bash
python3 scripts/check_integration_lane_scorecard.py \
  --input gait-out/adoption_metrics.json \
  --out gait-out/integration_lane_scorecard.json
```

Decision rule:

- select highest weighted score
- expansion allowed only when selected score >= `0.75` and confidence delta >= `0.03`

## v2.3 Metrics Snapshot (Release Note Input)

Snapshot path:

- `gait-out/v2_3_metrics_snapshot.json`

Required keys:

- `M1`, `M2`, `M3`, `M4`
- `C1`, `C2`, `C3`
- `D1`, `D2`, `D3`
- `release_gate_passed`

Produced by:

```bash
bash scripts/test_v2_3_acceptance.sh ./gait
```

## If Below Baseline

Prioritize in order:

1. onboarding conversion path (`scripts/quickstart.sh`, `docs/integration_checklist.md`)
2. contract stability (`scripts/test_intent_receipt_conformance.sh`, `scripts/test_contracts.sh`)
3. distribution reliability (`scripts/test_skill_supply_chain.sh`, CI template checks)
4. only then consider new lane expansion
