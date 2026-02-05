package registry

import "time"

type RegistryPack struct {
	SchemaID        string         `json:"schema_id"`
	SchemaVersion   string         `json:"schema_version"`
	CreatedAt       time.Time      `json:"created_at"`
	ProducerVersion string         `json:"producer_version"`
	PackName        string         `json:"pack_name"`
	PackVersion     string         `json:"pack_version"`
	Artifacts       []PackArtifact `json:"artifacts"`
	Signatures      []SignatureRef `json:"signatures,omitempty"`
}

type PackArtifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Media  string `json:"media_type,omitempty"`
}

type SignatureRef struct {
	Alg          string `json:"alg"`
	KeyID        string `json:"key_id"`
	Sig          string `json:"sig"`
	SignedDigest string `json:"signed_digest,omitempty"`
}
