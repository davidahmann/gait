export interface NavItem {
  title: string;
  href: string;
  children?: NavItem[];
}

export const navigation: NavItem[] = [
  {
    title: 'Start Here',
    href: '/docs',
    children: [
      { title: 'Install', href: '/docs/install' },
      { title: 'Mental Model', href: '/docs/concepts/mental_model' },
      { title: 'Architecture', href: '/docs/architecture' },
      { title: 'Flows', href: '/docs/flows' },
      { title: 'Simple Scenario', href: '/docs/scenarios/simple_agent_tool_boundary' },
      { title: 'Demo Output Legend', href: '/docs/demo_output_legend' },
      { title: 'Local UI Playground', href: '/docs/ui_localhost' },
    ],
  },
  {
    title: 'Durable Jobs & Packs',
    href: '/docs/contracts/packspec_v1',
    children: [
      { title: 'PackSpec v1', href: '/docs/contracts/packspec_v1' },
      { title: 'PackSpec TCK', href: '/docs/contracts/packspec_tck' },
      { title: 'Artifact Graph', href: '/docs/contracts/artifact_graph' },
      { title: 'Skill Provenance', href: '/docs/contracts/skill_provenance' },
    ],
  },
  {
    title: 'Voice Mode',
    href: '/docs/voice_mode',
    children: [
      { title: 'Voice Mode Guide', href: '/docs/voice_mode' },
    ],
  },
  {
    title: 'Context Evidence',
    href: '/docs/contracts/contextspec_v1',
    children: [
      { title: 'ContextSpec v1', href: '/docs/contracts/contextspec_v1' },
    ],
  },
  {
    title: 'Policy & Gates',
    href: '/docs/policy_authoring',
    children: [
      { title: 'Policy Authoring', href: '/docs/policy_authoring' },
      { title: 'Policy Rollout', href: '/docs/policy_rollout' },
      { title: 'Approval Runbook', href: '/docs/approval_runbook' },
      { title: 'Zero Trust Stack', href: '/docs/zero_trust_stack' },
      { title: 'External Tool Registry Policy', href: '/docs/external_tool_registry_policy' },
    ],
  },
  {
    title: 'Integrate',
    href: '/docs/integration_checklist',
    children: [
      { title: 'Integration Checklist', href: '/docs/integration_checklist' },
      { title: 'Integration Boundary Guide', href: '/docs/agent_integration_boundary' },
      { title: 'MCP Capability Matrix', href: '/docs/mcp_capability_matrix' },
      { title: 'Python SDK', href: '/docs/sdk/python' },
      { title: 'Cloud Runtime Patterns', href: '/docs/deployment/cloud_runtime_patterns' },
      { title: 'CI Regress Kit', href: '/docs/ci_regress_kit' },
    ],
  },
  {
    title: 'Hardening',
    href: '/docs/hardening/v2_2_contract',
    children: [
      { title: 'v2.2 Contract', href: '/docs/hardening/v2_2_contract' },
      { title: 'Production Runbook', href: '/docs/hardening/prime_time_runbook' },
      { title: 'Release Checklist', href: '/docs/hardening/release_checklist' },
      { title: 'Runtime SLO', href: '/docs/slo/runtime_slo' },
    ],
  },
  {
    title: 'Contracts',
    href: '/docs/contracts/primitive_contract',
    children: [
      { title: 'Primitive Contract', href: '/docs/contracts/primitive_contract' },
      { title: 'Intent+Receipt Spec', href: '/docs/contracts/intent_receipt_spec' },
      { title: 'Intent+Receipt Conformance', href: '/docs/contracts/intent_receipt_conformance' },
      { title: 'Endpoint Action Model', href: '/docs/contracts/endpoint_action_model' },
      { title: 'UI Contract', href: '/docs/contracts/ui_contract' },
    ],
  },
  {
    title: 'Blog',
    href: '/docs/blog/openclaw_24h_boundary_enforcement',
    children: [
      { title: '2,880 Tool Calls Gate-Checked', href: '/docs/blog/openclaw_24h_boundary_enforcement' },
    ],
  },
  {
    title: 'Ecosystem',
    href: '/docs/ecosystem/awesome',
    children: [
      { title: 'Community Index', href: '/docs/ecosystem/awesome' },
      { title: 'Contribute', href: '/docs/ecosystem/contribute' },
      { title: 'Homebrew', href: '/docs/homebrew' },
      { title: 'Packaging', href: '/docs/packaging' },
    ],
  },
];
