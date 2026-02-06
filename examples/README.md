# Examples (Offline-Safe)

All examples in this folder run without network, secrets, or cloud accounts.

## Included Paths

- `stub-replay/`: deterministic stub replay flow from a demo runpack
- `policy-test/`: allow/block/require-approval policy evaluation examples
- `regress-run/`: incident-to-regression fixture workflow
- `prompt-injection/`: deterministic prompt-injection style blocking example
- `python/`: thin Python adapter example (calls local `gait` binary)
- `integrations/openai_agents/`: wrapped tool path with allow/block + trace outputs
- `integrations/langchain/`: wrapped tool path with allow/block + trace outputs
- `integrations/autogen/`: wrapped tool path with allow/block + trace outputs

## Recommended Order

1. `stub-replay`
2. `policy-test`
3. `regress-run`
4. `prompt-injection`
5. `integrations/openai_agents`
6. `integrations/langchain`
7. `integrations/autogen`
