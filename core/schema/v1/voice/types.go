package voice

import "time"

type CommitmentIntent struct {
	SchemaID                string            `json:"schema_id"`
	SchemaVersion           string            `json:"schema_version"`
	CreatedAt               time.Time         `json:"created_at"`
	ProducerVersion         string            `json:"producer_version"`
	CallID                  string            `json:"call_id"`
	TurnIndex               int               `json:"turn_index"`
	CallSeq                 int               `json:"call_seq"`
	CommitmentClass         string            `json:"commitment_class"`
	UtteranceDigest         string            `json:"utterance_digest,omitempty"`
	Context                 CommitmentContext `json:"context"`
	Currency                string            `json:"currency,omitempty"`
	QuoteMinCents           int64             `json:"quote_min_cents,omitempty"`
	QuoteMaxCents           int64             `json:"quote_max_cents,omitempty"`
	RefundCeilingCents      int64             `json:"refund_ceiling_cents,omitempty"`
	ApprovalRequired        bool              `json:"approval_required,omitempty"`
	ApprovalArtifactRefs    []string          `json:"approval_artifact_refs,omitempty"`
	EvidenceReferenceDigests []string         `json:"evidence_reference_digests,omitempty"`
}

type CommitmentContext struct {
	Identity               string `json:"identity"`
	Workspace              string `json:"workspace"`
	RiskClass              string `json:"risk_class"`
	SessionID              string `json:"session_id,omitempty"`
	RequestID              string `json:"request_id,omitempty"`
	EnvironmentFingerprint string `json:"environment_fingerprint,omitempty"`
}

type SayToken struct {
	SchemaID           string     `json:"schema_id"`
	SchemaVersion      string     `json:"schema_version"`
	CreatedAt          time.Time  `json:"created_at"`
	ProducerVersion    string     `json:"producer_version"`
	TokenID            string     `json:"token_id"`
	CommitmentClass    string     `json:"commitment_class"`
	IntentDigest       string     `json:"intent_digest"`
	PolicyDigest       string     `json:"policy_digest"`
	CallID             string     `json:"call_id"`
	TurnIndex          int        `json:"turn_index"`
	CallSeq            int        `json:"call_seq"`
	Currency           string     `json:"currency,omitempty"`
	QuoteMinCents      int64      `json:"quote_min_cents,omitempty"`
	QuoteMaxCents      int64      `json:"quote_max_cents,omitempty"`
	RefundCeilingCents int64      `json:"refund_ceiling_cents,omitempty"`
	ExpiresAt          time.Time  `json:"expires_at"`
	Signature          *Signature `json:"signature,omitempty"`
}

type Signature struct {
	Alg          string `json:"alg"`
	KeyID        string `json:"key_id"`
	Sig          string `json:"sig"`
	SignedDigest string `json:"signed_digest,omitempty"`
}

type CallEvent struct {
	SchemaID        string         `json:"schema_id"`
	SchemaVersion   string         `json:"schema_version"`
	CreatedAt       time.Time      `json:"created_at"`
	CallID          string         `json:"call_id"`
	CallSeq         int            `json:"call_seq"`
	TurnIndex       int            `json:"turn_index"`
	EventType       string         `json:"event_type"`
	CommitmentClass string         `json:"commitment_class,omitempty"`
	IntentDigest    string         `json:"intent_digest,omitempty"`
	PolicyDigest    string         `json:"policy_digest,omitempty"`
	SayTokenID      string         `json:"say_token_id,omitempty"`
	PayloadDigest   string         `json:"payload_digest,omitempty"`
	Attributes      map[string]any `json:"attributes,omitempty"`
}

type CallpackManifest struct {
	SchemaID               string    `json:"schema_id"`
	SchemaVersion          string    `json:"schema_version"`
	CreatedAt              time.Time `json:"created_at"`
	ProducerVersion        string    `json:"producer_version"`
	CallID                 string    `json:"call_id"`
	PrivacyMode            string    `json:"privacy_mode"`
	EventCount             int       `json:"event_count"`
	CommitmentCount        int       `json:"commitment_count"`
	DecisionCount          int       `json:"decision_count"`
	SpeakReceiptCount      int       `json:"speak_receipt_count"`
	ReferenceDigestCount   int       `json:"reference_digest_count"`
	EnvironmentFingerprint string    `json:"environment_fingerprint,omitempty"`
}

type GateDecision struct {
	CallID          string   `json:"call_id"`
	CallSeq         int      `json:"call_seq"`
	TurnIndex       int      `json:"turn_index"`
	CommitmentClass string   `json:"commitment_class"`
	Verdict         string   `json:"verdict"`
	ReasonCodes     []string `json:"reason_codes,omitempty"`
	IntentDigest    string   `json:"intent_digest,omitempty"`
	PolicyDigest    string   `json:"policy_digest,omitempty"`
	ApprovalRef     string   `json:"approval_ref,omitempty"`
}

type SpeakReceipt struct {
	CallID          string    `json:"call_id"`
	CallSeq         int       `json:"call_seq"`
	TurnIndex       int       `json:"turn_index"`
	CommitmentClass string    `json:"commitment_class"`
	SayTokenID      string    `json:"say_token_id"`
	SpokenDigest    string    `json:"spoken_digest"`
	EmittedAt       time.Time `json:"emitted_at"`
}

type ReferenceDigest struct {
	RefID  string `json:"ref_id"`
	SHA256 string `json:"sha256"`
}

