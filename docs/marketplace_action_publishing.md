---
title: "GitHub Marketplace Action Publishing"
description: "How to publish gait-regress as a Marketplace action and why monorepo layouts usually do not auto-enable publish controls."
---

# GitHub Marketplace Action Publishing

## Why publish may not appear in this repo

GitHub Marketplace publishing for actions expects action-repo layout constraints that this monorepo does not currently follow.

Current monorepo state:
- action metadata is at `.github/actions/gait-regress/action.yml` (not repo root)
- repo contains workflow files under `.github/workflows/`

Typical Marketplace action expectations:
- a single `action.yml` at repository root
- action-focused repository shape suitable for listing
- owner acceptance of Marketplace developer terms

## Recommended Path

Use a dedicated public action repository, for example `gait-regress-action`.

Required setup:
1. place `action.yml` at repo root
2. include action runtime files + README
3. add metadata quality fields (`name`, `description`, `branding`)
4. create semantic release tags (`v1.0.0`, `v1`)
5. publish listing from that repository

Then consume from external repos via:

```yaml
- uses: <owner>/gait-regress-action@v1
```

## In-Repo Positioning

This monorepo keeps `.github/actions/gait-regress/` as the canonical source implementation and test fixture for workflow/template parity.

Marketplace distribution should reference the dedicated action repo.
