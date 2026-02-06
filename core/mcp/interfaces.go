package mcp

import (
	"context"
	"time"
)

type ToolCall struct {
	Name          string          `json:"name"`
	Args          map[string]any  `json:"args,omitempty"`
	Target        string          `json:"target,omitempty"`
	Targets       []Target        `json:"targets,omitempty"`
	ArgProvenance []ArgProvenance `json:"arg_provenance,omitempty"`
	Context       CallContext     `json:"context,omitempty"`
	CreatedAt     time.Time       `json:"created_at,omitempty"`
}

type Target struct {
	Kind        string `json:"kind"`
	Value       string `json:"value"`
	Operation   string `json:"operation,omitempty"`
	Sensitivity string `json:"sensitivity,omitempty"`
}

type ArgProvenance struct {
	ArgPath         string `json:"arg_path"`
	Source          string `json:"source"`
	SourceRef       string `json:"source_ref,omitempty"`
	IntegrityDigest string `json:"integrity_digest,omitempty"`
}

type CallContext struct {
	Identity  string `json:"identity,omitempty"`
	Workspace string `json:"workspace,omitempty"`
	RiskClass string `json:"risk_class,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	RunID     string `json:"run_id,omitempty"`
}

type ToolResult struct {
	Status string         `json:"status"`
	Output map[string]any `json:"output,omitempty"`
}

type Adapter interface {
	CallTool(context.Context, ToolCall) (ToolResult, error)
}
