# Retention Profiles (OSS)

Recommended retention defaults for local/customer-hosted deployments.

## Short Horizon (Dev/Smoke)

- `trace_ttl: 24h`
- `session_ttl: 72h`
- `export_ttl: 24h`

## Medium Horizon (Team Production)

- `trace_ttl: 168h` (7 days)
- `session_ttl: 336h` (14 days)
- `export_ttl: 168h` (7 days)

## Long Horizon (High Audit Pressure)

- `trace_ttl: 720h` (30 days)
- `session_ttl: 1440h` (60 days)
- `export_ttl: 720h` (30 days)

## Notes

- Apply the profile in `.gait/config.yaml` under `retention`.
- For `mcp serve`, also enforce `--*-max-age` and `--*-max-count` flags.
- Keep dry-run retention checks in CI before policy changes where possible.
