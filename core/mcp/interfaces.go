package mcp

import "context"

type ToolCall struct {
	Name   string
	Args   map[string]any
	Target string
}

type ToolResult struct {
	Status string
	Output map[string]any
}

type Adapter interface {
	CallTool(context.Context, ToolCall) (ToolResult, error)
}
