# Gait Architecture

This document is the canonical architecture view for OSS v1.

## Component Architecture

```mermaid
flowchart LR
    operator["Developer / Platform Engineer / CI"] --> cli["gait CLI (cmd/gait)"]

    subgraph core["Go Core (core/*)"]
        dispatch["Command dispatch + exit contracts"]
        runpack["Runpack (core/runpack)"]
        gate["Gate + policy eval (core/gate)"]
        regress["Regress (core/regress)"]
        doctor["Doctor (core/doctor)"]
        scout["Scout signal + snapshots (core/scout)"]
        guard["Guard/evidence (core/guard)"]
        sign["JCS + signing (core/jcs, core/sign)"]
        schema["Schema validation (core/schema)"]
        fsx["Deterministic FS utilities (core/fsx, core/zipx)"]
    end

    cli --> dispatch
    dispatch --> runpack
    dispatch --> gate
    dispatch --> regress
    dispatch --> doctor
    dispatch --> scout
    dispatch --> guard
    runpack --> sign
    gate --> sign
    regress --> runpack
    doctor --> schema
    dispatch --> schema
    dispatch --> fsx

    subgraph adapters["Adoption Layer (non-authoritative)"]
        py["Python SDK (sdk/python)"]
        sidecar["Sidecar patterns (examples/sidecar)"]
        framework["Framework adapters (examples/integrations/*)"]
        mcpserve["Local MCP interception service (gait mcp serve)"]
    end

    py --> cli
    sidecar --> cli
    framework --> py
    mcpserve --> gate

    subgraph artifacts["Artifact Surface (durable contract)"]
        runpackZip["gait-out/runpack_{run_id}.zip"]
        sessionChain["gait-out/sessions/*_chain.json"]
        traceJson["trace_{id}.json"]
        delegationAudit["delegation_audit_{id}.json"]
        regressJson["regress_result.json"]
        junitXml["junit.xml"]
        evidenceZip["Evidence packs (gait guard/incident pack)"]
    end

    runpack --> runpackZip
    runpack --> sessionChain
    gate --> traceJson
    gate --> delegationAudit
    regress --> regressJson
    regress --> junitXml
    guard --> evidenceZip
```

## Runtime Boundaries

- **Authoritative boundary**: Go core owns policy decisions, canonicalization, signing, schema validation, determinism, and exit codes.
- **Adoption boundary**: Python and sidecars should only serialize intent, call CLI, and interpret structured results.
- **Durable contract boundary**: artifacts and schemas are the long-lived API surface, not in-memory implementation details.

## State And Persistence

- Working artifacts: `./gait-out/`
- Session journals and chains: `./gait-out/sessions/*` (append-only journal + checkpoint chain)
- MCP serve runtime traces: `./gait-out/mcp-serve/traces`
- Regress fixtures/config: `fixtures/` and `gait.yaml`
- Optional local caches: `~/.gait/runpacks`, `~/.gait/registry`
- Schema contracts: `schemas/v1/*` with matching Go types/validators under `core/schema/*`

## Failure Posture

- Default-safe behavior is enforced in command handlers: non-`allow` gate outcomes do not execute side effects.
- High-risk production profile (`oss-prod`) remains fail-closed when policy or credentials cannot be evaluated.
- Delegation-constrained policies remain fail-closed when required delegation evidence is absent/invalid.
- Verification and regression failures return stable non-zero exit codes for automation.
