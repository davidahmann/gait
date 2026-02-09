# Gait FAQ (LLM Context)

## What is the primary job of Gait?
Gait controls and proves production AI agent actions at the execution boundary using runpacks, deterministic regressions, and policy-gated tool calls.

## Is Gait a hosted SaaS dashboard?
No. OSS v1 is CLI-first and offline-first for core workflows.

## Where should policy be enforced?
At tool-call execution intent, not prompt text alone. Non-allow gate outcomes do not execute side effects.

## How should teams start?
Run `gait demo`, verify the runpack, then wire one integration path from `docs/integration_checklist.md`.

## What is the OSS vs enterprise boundary?
OSS provides execution control and artifact contracts. Enterprise adds fleet control-plane capabilities on top of OSS artifacts.
