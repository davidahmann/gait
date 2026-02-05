package gate

import "time"

type TraceRecord struct {
	SchemaID         string     `json:"schema_id"`
	SchemaVersion    string     `json:"schema_version"`
	CreatedAt        time.Time  `json:"created_at"`
	ProducerVersion  string     `json:"producer_version"`
	TraceID          string     `json:"trace_id"`
	ToolName         string     `json:"tool_name"`
	ArgsDigest       string     `json:"args_digest"`
	PolicyDigest     string     `json:"policy_digest"`
	Verdict          string     `json:"verdict"`
	Violations       []string   `json:"violations,omitempty"`
	LatencyMS        float64    `json:"latency_ms,omitempty"`
	ApprovalTokenRef string     `json:"approval_token_ref,omitempty"`
	Signature        *Signature `json:"signature,omitempty"`
}

type Signature struct {
	Alg   string `json:"alg"`
	KeyID string `json:"key_id"`
	Sig   string `json:"sig"`
}

type IntentRequest struct {
	SchemaID        string         `json:"schema_id"`
	SchemaVersion   string         `json:"schema_version"`
	CreatedAt       time.Time      `json:"created_at"`
	ProducerVersion string         `json:"producer_version"`
	ToolName        string         `json:"tool_name"`
	Args            map[string]any `json:"args"`
	ArgsDigest      string         `json:"args_digest"`
	Context         IntentContext  `json:"context"`
}

type IntentContext struct {
	Identity  string `json:"identity"`
	RiskClass string `json:"risk_class"`
}

type GateResult struct {
	SchemaID        string    `json:"schema_id"`
	SchemaVersion   string    `json:"schema_version"`
	CreatedAt       time.Time `json:"created_at"`
	ProducerVersion string    `json:"producer_version"`
	Verdict         string    `json:"verdict"`
	ReasonCodes     []string  `json:"reason_codes"`
	Violations      []string  `json:"violations"`
}
