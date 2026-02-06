package scout

import "time"

type InventorySnapshot struct {
	SchemaID        string          `json:"schema_id"`
	SchemaVersion   string          `json:"schema_version"`
	CreatedAt       time.Time       `json:"created_at"`
	ProducerVersion string          `json:"producer_version"`
	SnapshotID      string          `json:"snapshot_id"`
	Workspace       string          `json:"workspace,omitempty"`
	Items           []InventoryItem `json:"items"`
}

type InventoryItem struct {
	ID          string   `json:"id"`
	Kind        string   `json:"kind"`
	Name        string   `json:"name"`
	Locator     string   `json:"locator"`
	RiskLevel   string   `json:"risk_level,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	LastSeenRun string   `json:"last_seen_run,omitempty"`
}

type AdoptionEvent struct {
	SchemaID        string             `json:"schema_id"`
	SchemaVersion   string             `json:"schema_version"`
	CreatedAt       time.Time          `json:"created_at"`
	ProducerVersion string             `json:"producer_version"`
	Command         string             `json:"command"`
	Success         bool               `json:"success"`
	ExitCode        int                `json:"exit_code"`
	ElapsedMS       int64              `json:"elapsed_ms"`
	Milestones      []string           `json:"milestones,omitempty"`
	Environment     AdoptionEnvContext `json:"environment"`
}

type AdoptionEnvContext struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}
