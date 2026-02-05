package runpack

import "time"

type Manifest struct {
	SchemaID        string         `json:"schema_id"`
	SchemaVersion   string         `json:"schema_version"`
	CreatedAt       time.Time      `json:"created_at"`
	ProducerVersion string         `json:"producer_version"`
	RunID           string         `json:"run_id"`
	CaptureMode     string         `json:"capture_mode"`
	Files           []ManifestFile `json:"files"`
	ManifestDigest  string         `json:"manifest_digest"`
	Signatures      []Signature    `json:"signatures,omitempty"`
}

type ManifestFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type Signature struct {
	Alg          string `json:"alg"`
	KeyID        string `json:"key_id"`
	Sig          string `json:"sig"`
	SignedDigest string `json:"signed_digest,omitempty"`
}

type Run struct {
	SchemaID        string        `json:"schema_id"`
	SchemaVersion   string        `json:"schema_version"`
	CreatedAt       time.Time     `json:"created_at"`
	ProducerVersion string        `json:"producer_version"`
	RunID           string        `json:"run_id"`
	Env             RunEnv        `json:"env"`
	Timeline        []TimelineEvt `json:"timeline"`
}

type RunEnv struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Runtime string `json:"runtime"`
}

type TimelineEvt struct {
	Event string    `json:"event"`
	TS    time.Time `json:"ts"`
	Ref   string    `json:"ref,omitempty"`
}

type IntentRecord struct {
	SchemaID        string         `json:"schema_id"`
	SchemaVersion   string         `json:"schema_version"`
	CreatedAt       time.Time      `json:"created_at"`
	ProducerVersion string         `json:"producer_version"`
	RunID           string         `json:"run_id"`
	IntentID        string         `json:"intent_id"`
	ToolName        string         `json:"tool_name"`
	ArgsDigest      string         `json:"args_digest"`
	Args            map[string]any `json:"args,omitempty"`
}

type ResultRecord struct {
	SchemaID        string         `json:"schema_id"`
	SchemaVersion   string         `json:"schema_version"`
	CreatedAt       time.Time      `json:"created_at"`
	ProducerVersion string         `json:"producer_version"`
	RunID           string         `json:"run_id"`
	IntentID        string         `json:"intent_id"`
	Status          string         `json:"status"`
	ResultDigest    string         `json:"result_digest"`
	Result          map[string]any `json:"result,omitempty"`
}

type Refs struct {
	SchemaID        string       `json:"schema_id"`
	SchemaVersion   string       `json:"schema_version"`
	CreatedAt       time.Time    `json:"created_at"`
	ProducerVersion string       `json:"producer_version"`
	RunID           string       `json:"run_id"`
	Receipts        []RefReceipt `json:"receipts"`
}

type RefReceipt struct {
	RefID            string         `json:"ref_id"`
	SourceType       string         `json:"source_type"`
	SourceLocator    string         `json:"source_locator"`
	QueryDigest      string         `json:"query_digest"`
	ContentDigest    string         `json:"content_digest"`
	RetrievedAt      time.Time      `json:"retrieved_at"`
	RedactionMode    string         `json:"redaction_mode"`
	SensitivityLabel string         `json:"sensitivity_label,omitempty"`
	RetrievalParams  map[string]any `json:"retrieval_params,omitempty"`
}
