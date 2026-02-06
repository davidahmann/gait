# Security Policy

## Reporting a Vulnerability

Please report security issues privately.

- Email: dahmann@lumyn.cc

Do not open public issues for security vulnerabilities.

## Supported Versions

Only the latest tagged release is supported with security fixes.

## Approval Key Management

Approval and trace signing keys are operational security assets:

- Store private keys in a secret manager or HSM-backed system.
- Do not commit keys, tokens, or customer artifacts to git.
- Rotate signing keys on a fixed schedule and immediately after suspected compromise.
- Publish and distribute public verification keys before key cutover.
- Keep approval signing keys separate from release signing keys.

Operational procedure:

- Follow `docs/approval_runbook.md` for token minting, TTL/scope policy, and incident audit workflow.
