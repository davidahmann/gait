# ADR 0004: Network Retry Policy For Registry Remote Fetch

- Status: Accepted
- Date: 2026-02-06
- Related epic: `H4`

## Context

Remote registry fetches currently have limited resilience under transient failures. Reliability requires retries, while trust must remain fail-closed.

## Decision

Add bounded retry/backoff for transient remote fetch failures only:

- retry transport timeouts and selected 5xx responses
- no retry for policy/trust failures (allowlist/signature/pin violations)
- deterministic max-attempt budget
- explicit categorization of transient vs permanent failures

## Alternatives Considered

1. No retries.
   - Rejected: poor resilience to transient faults.
2. Retry all failures.
   - Rejected: can hide trust or configuration faults.

## Consequences

- Better reliability for optional remote operations.
- Slightly longer worst-case latency bounded by configured budget.
