package policytest

import "time"

type PolicyTestResult struct {
	SchemaID          string    `json:"schema_id"`
	SchemaVersion     string    `json:"schema_version"`
	CreatedAt         time.Time `json:"created_at"`
	ProducerVersion   string    `json:"producer_version"`
	PolicyDigest      string    `json:"policy_digest"`
	IntentDigest      string    `json:"intent_digest"`
	Verdict           string    `json:"verdict"`
	ReasonCodes       []string  `json:"reason_codes"`
	Violations        []string  `json:"violations"`
	MatchedRule       string    `json:"matched_rule,omitempty"`
	DelegationMatched bool      `json:"delegation_matched,omitempty"`
	DelegationDepth   int       `json:"delegation_depth,omitempty"`
	DelegationScope   string    `json:"delegation_scope,omitempty"`
}
