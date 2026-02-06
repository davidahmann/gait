---
name: gait-incident-to-regression
description: Convert a Gait run artifact into a deterministic regression workflow. Use when asked to initialize fixtures from run_id or runpack path, run graders, produce CI-friendly outputs, or summarize drift and failures.
disable-model-invocation: true
---
# Incident To Regression

Execute this workflow to transform an observed run into repeatable CI checks.

## Workflow

1. Resolve source run artifact:
   - use `<run_id>` or `<runpack_path>`
2. Initialize fixture deterministically:
   - `gait regress init --from <run_id_or_path> --json`
3. Parse and report:
   - `ok`, `run_id`, `fixture_name`, `fixture_dir`, `config_path`, `next_commands`
4. Run regression suite:
   - `gait regress run --json`
5. If CI output is requested, add JUnit:
   - `gait regress run --json --junit junit.xml`
6. Return concise summary:
   - source run
   - fixture path
   - pass/fail status
   - failed graders count
   - output paths

## Safety Rules

- Keep replay deterministic defaults.
- For replay workflows, prefer `gait run replay` (stub mode default); require explicit unsafe flags for real tool replay.
- Do not pass `--allow-nondeterministic` unless explicitly requested.
- Treat non-zero regress run exits as regressions, not soft warnings.

## Determinism Rules

- Always run `regress init` before `regress run` for new incidents.
- Always consume `--json` output fields for decisions.
- Keep fixture names stable and explicit when user provides naming constraints.
