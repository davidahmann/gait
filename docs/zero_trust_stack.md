---
title: "Zero Trust Stack"
description: "How Gait implements zero trust for agent tool calls with fail-closed enforcement, signed traces, and offline verification."
---

# Gait In Your Zero Trust Stack

This document explains where Gait fits, what it integrates with, and what it intentionally does not replace.

## One-Line Positioning

Guardrails scan content. Gait evaluates structured action intent before execution and emits verifiable proof.

Short version: scanners and monitoring observe; Gait gates execution. You want both.

## Layer Model

1. Identity and credential systems (CyberArk, HashiCorp Vault, cloud IAM):
   - establish who can act and how credentials are issued
2. Guardrail and gateway systems:
   - scan prompts and outputs for injection/leakage patterns
3. Gait execution boundary:
   - evaluate `IntentRequest` against policy
   - fail closed for high-risk ambiguity
   - emit signed traces and runpack evidence
4. Monitoring and SIEM systems (Splunk, Datadog, Elastic):
   - ingest Gait artifacts/events for fleet analytics and response

## What Gait Owns

- deterministic policy decisions at tool-call boundary (`gait gate eval`)
- approval enforcement and signed approval audit records
- signed trace records and verifiable runpack artifacts
- deterministic incident-to-regression loop

## What Gait Does Not Own

- identity provider lifecycle
- credential vault lifecycle
- prompt/output scanning engines
- hosted SIEM/dashboard products

## Integration Patterns

### Identity And Auth Context Passthrough

`IntentRequest.context` supports additive optional fields for external identity context:

- `auth_context` (object)
- `credential_scopes` (array of strings)
- `environment_fingerprint` (string)

These are passthrough fields for policy and audit context. They do not change default OSS execution semantics.

### Tool Registry To Policy

Keep external registry ownership in your platform stack; convert allowlist exports into Gait policy:

- `docs/external_tool_registry_policy.md`

### SIEM Export

Use `gait mcp proxy` (one-shot) or `gait mcp serve` (long-running local service) and ingest JSONL/OTEL exports in existing observability stack:

- `docs/siem_ingestion_recipes.md`

## Runtime Governance vs ACP

Runtime governance (as commonly used) is often observational and alert-driven.

ACP in Gait is decision-and-proof at execution time:

- runtime governance asks: "is behavior acceptable?" and monitors
- ACP asks: "is this action allowed now?" and decides before side effects

Use this language consistently in docs and customer conversations.

## Frequently Asked Questions

### How does Gait implement zero trust for agents?

Every tool call is treated as untrusted until evaluated against policy. Non-allow verdicts do not execute. Every decision is signed and traceable.

### What is fail-closed enforcement?

When policy evaluation encounters ambiguity — no matching rule, missing context evidence, or invalid approval token — the default is block. Execution requires an explicit allow.

### How are gate traces signed?

Traces are signed with Ed25519 keys using JCS (RFC 8785) canonicalization for deterministic JSON serialization. This prevents ordering attacks and ensures verifiable integrity.

### Can I verify traces offline?

Yes. `gait trace verify` validates trace signatures and schema offline with no network dependency.

### What is the difference between gate evaluation and prompt filtering?

Gate evaluation operates on structured tool-call intent at execution time. Prompt filtering operates on free-form text at generation time. Gait enforces at the action boundary, not the prompt boundary.
