# ADR 0001: Error Taxonomy And Exit Mapping

- Status: Accepted
- Date: 2026-02-06
- Related epic: `H1`

## Context

Operational faults are inconsistently surfaced and sometimes collapse into input-oriented errors, reducing automation reliability.

## Decision

Introduce a canonical runtime error taxonomy and map categories to stable CLI exit behavior without breaking existing public contracts.

Key requirements:

- machine-readable `error_category` and `error_code`
- explicit retryability signal
- deterministic mapping in `--json` outputs

## Alternatives Considered

1. Keep existing ad-hoc error handling.
   - Rejected: nondeterministic behavior for automation.
2. Add new exit codes for every category immediately.
   - Rejected: high compatibility risk.

## Consequences

- Improved operator and CI diagnosability.
- Additional test fixtures are required to lock error-surface behavior.
