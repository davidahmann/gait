package guard

import "time"

type PackManifest struct {
	SchemaID        string      `json:"schema_id"`
	SchemaVersion   string      `json:"schema_version"`
	CreatedAt       time.Time   `json:"created_at"`
	ProducerVersion string      `json:"producer_version"`
	PackID          string      `json:"pack_id"`
	RunID           string      `json:"run_id"`
	CaseID          string      `json:"case_id,omitempty"`
	GeneratedAt     time.Time   `json:"generated_at"`
	Contents        []PackEntry `json:"contents"`
}

type PackEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Type   string `json:"type"`
}
