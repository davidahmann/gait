# ADR 0002: Atomic Write Strategy

- Status: Accepted
- Date: 2026-02-06
- Related epic: `H2`

## Context

Critical runtime files are written by multiple flows. Partial writes can corrupt state and break deterministic behavior.

## Decision

Use a shared atomic write utility for critical files:

1. write to temp file in destination directory
2. flush and sync file contents
3. apply explicit file mode
4. atomically rename into final path

## Alternatives Considered

1. Keep direct `os.WriteFile` usage in all call sites.
   - Rejected: partial-write corruption risk.
2. Add per-package custom atomic helpers.
   - Rejected: duplicated logic and inconsistent semantics.

## Consequences

- Improved crash consistency for state and artifact writes.
- Requires migration of existing critical write paths and targeted failure tests.
