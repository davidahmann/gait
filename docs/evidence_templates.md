# Evidence Templates (Epic A7.2)

Use these templates to standardize evidence handoff in incidents, pull requests, and postmortems.
If your agent touched prod, attach the runpack.

Conventions:

- Always include `run_id`.
- Always include an executable verify command: `gait verify <run_id>`.
- Prefer immutable digest references from `ticket_footer`.
- Persist the emitted `ticket_footer` at record time (`gait demo --json` or recorder output).

## Incident Ticket Template

```text
Title: [Agent Incident] <short summary>

Impact:
- User/system impact:
- Start time (UTC):
- Detection source:

Gait Evidence:
- run_id: <run_id>
- verify: gait verify <run_id>
- ticket_footer: GAIT run_id=<run_id> manifest=sha256:<digest> verify="gait verify <run_id>"
- gate_trace: trace_<trace_id>.json
- approval_audit (if applicable): approval_audit_<trace_id>.json

Reproduction:
1) gait regress init --from <run_id> --json
2) gait regress run --json --junit=./gait-out/junit.xml

Current Status:
- [ ] mitigated
- [ ] root cause identified
- [ ] regression added
```

## PR Description Template

```text
## Why
- Brief issue summary:

## Evidence
- run_id: <run_id>
- verify: gait verify <run_id>
- regress_result: gait regress run --json
- policy_result (if policy changed): gait policy test <policy.yaml> <intent.json> --json

## Change Scope
- Components touched:
- Risk class:

## Validation
- [ ] make lint
- [ ] make test
- [ ] gait verify <run_id>
- [ ] regress fixture updated or confirmed unchanged
```

## Postmortem Section Template

```text
### Deterministic Evidence
- run_id: <run_id>
- verify command: gait verify <run_id>
- manifest digest: sha256:<digest>
- associated regress fixture: fixtures/<fixture_name>/

### Enforcement Outcome
- gate verdict:
- reason codes:
- approval required: yes/no
- approval ref/token id (if used):

### Preventive Action
- Regression command: gait regress run --json
- Policy hardening changes:
- Owner + due date:
```

## Design-Partner Incident Conversion Runbook (OSS-Safe)

Use this sequence to convert one real incident into deterministic OSS artifacts:

1. capture/normalize run input into a runpack:
   `gait run record --input <incident_run_record.json> --json`
2. verify runpack integrity:
   `gait verify <run_id|runpack_path> --json`
3. mint deterministic ticket footer:
   `gait run receipt --from <run_id|runpack_path> --json`
4. convert incident to regression fixture:
   `gait regress init --from <run_id|runpack_path> --json`
5. enforce regression:
   `gait regress run --json --junit ./gait-out/junit.xml`
6. prove boundary policy behavior for the incident intent:
   `gait gate eval --policy <policy.yaml> --intent <intent.json> --trace-out ./gait-out/trace_incident.json --json`

Outputs remain fully local and compatible with existing schema/verify contracts.
