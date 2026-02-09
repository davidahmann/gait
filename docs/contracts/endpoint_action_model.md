# Endpoint Action Model (Normative)

Status: normative for OSS `v1.7+`.

This document defines the deterministic endpoint taxonomy used by Gate for policy evaluation and traceability.

## Purpose

Endpoint taxonomy gives every tool-call target a stable execution class so policy can control:

- filesystem access
- process execution
- network egress
- destructive operations

## Endpoint Classes

Stable class identifiers:

- `fs.read`
- `fs.write`
- `fs.delete`
- `proc.exec`
- `net.http`
- `net.dns`
- `other`

Reserved (additive, not required in v1.7):

- `ui.click`
- `ui.type`
- `ui.navigate`

`other` means the action could not be classified into a stricter class. In fail-closed high-risk paths, `other` is treated as non-evaluable.

## Classification Rules

Classification is deterministic and based on normalized target fields:

- `kind=path`:
  - read-like operations -> `fs.read`
  - write-like operations -> `fs.write`
  - delete-like operations -> `fs.delete`
- `kind=host|url`:
  - DNS-like operations -> `net.dns`
  - otherwise -> `net.http`
- `kind=other` with exec-like operation/tool hint -> `proc.exec`
- all unresolved cases -> `other`

Each target may also include:

- `endpoint_domain` (for host/url targets)
- `destructive` (true for delete/exec style operations)

## Policy Controls

Gate supports endpoint constraints in rules:

- `path_allowlist`
- `path_denylist`
- `domain_allowlist`
- `domain_denylist`
- `egress_classes`
- `destructive_action`

Constraint violations produce deterministic reason/violation codes and can force `block` or `require_approval`.

## Fail-Closed Requirement

When fail-closed applies to high-risk intents:

- unknown endpoint classes (`""` or `other`) produce:
  - reason code: `fail_closed_endpoint_class_unknown`
  - violation: `endpoint_class_unknown`
- Gate blocks execution.

## Compatibility

- Existing v1 intents remain valid.
- Endpoint metadata fields are additive and optional in schema.
- Normalization infers endpoint metadata when not provided.
