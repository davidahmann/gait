# Examples Gallery

This page points to runnable, deterministic examples.

## Core Walkthroughs

- 90-second first win: `scripts/demo_90s.sh`
- Scenario packs:
  - `examples/scenarios/incident_reproduction.sh`
  - `examples/scenarios/prompt_injection_block.sh`
  - `examples/scenarios/approval_flow.sh`

## Integration Examples

- `examples/integrations/openai_agents/`
- `examples/integrations/langchain/`
- `examples/integrations/autogen/`
- `examples/integrations/openclaw/`
- `examples/integrations/autogpt/`
- `examples/integrations/template/`

## Policy Examples

- Endpoint controls: `examples/policy/endpoint/`
- Skills trust: `examples/policy/skills/`
- Prompt injection guardrails: `examples/prompt-injection/`

## Skill Provenance Examples

- `examples/skills/registry_pack_example.json`
- `docs/contracts/skill_provenance.md`

## Validate Everything

```bash
make test-adoption
make test-adapter-parity
make test-skill-supply-chain
make test-uat-local
```
