# ADR 0003: Lock Contention Strategy

- Status: Accepted
- Date: 2026-02-06
- Related epic: `H3`

## Context

Concurrent operations on shared state (notably gate rate-limit state) can produce nondeterministic failures or stale lock artifacts.

## Decision

Adopt a bounded contention model:

- lock metadata includes owner and timestamp
- stale lock detection is explicit
- retry behavior is bounded with deterministic timeout
- timeout failures are categorized as contention failures

## Alternatives Considered

1. Infinite retries.
   - Rejected: can hang automation.
2. Immediate failure on first lock miss.
   - Rejected: fragile under short-lived contention.

## Consequences

- Deterministic behavior under concurrency pressure.
- Additional integration tests are required for concurrent invocations.
