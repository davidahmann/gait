---
name: gait-capture-runpack
description: Capture and verify deterministic Gait runpacks from normalized run input. Use when asked to record a run, produce run_id or runpack artifacts, generate ticket-ready proof, or validate artifact integrity before handoff.
disable-model-invocation: true
---
# Capture Runpack

Execute this workflow to record an artifact safely and deterministically.

## Workflow

1. Validate required input path: `<run_record.json>`.
2. Run record with JSON output:
   - `gait run record --input <run_record.json> --json`
3. Parse output fields:
   - `ok`, `run_id`, `bundle`, `manifest_digest`, `ticket_footer`
4. Verify artifact integrity:
   - `gait verify <run_id_or_bundle_path> --json`
5. Return a concise handoff block that includes:
   - `run_id`
   - `bundle`
   - `manifest_digest`
   - `ticket_footer`
   - verify status

## Safety Rules

- Keep default capture mode as `reference`.
- Do not switch to raw capture unless explicitly requested.
- For replay workflows, prefer `gait run replay` (stub mode default); require explicit unsafe flags for real tool replay.
- Do not invent `run_id`, digests, or verify results.
- Treat non-zero exit from `gait run record` or `gait verify` as blocking errors.

## Determinism Rules

- Always use `--json` and parse structured fields.
- Do not rely on text-only output for workflow decisions.
- Keep output grounded in recorded artifact values only.
