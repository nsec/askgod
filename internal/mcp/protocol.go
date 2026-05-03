package mcp

import "encoding/json"

// Request is a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// Error is a JSON-RPC 2.0 error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// InitializeResult is the result of an MCP initialize request.
type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"` //nolint:tagliatelle
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"` //nolint:tagliatelle
}

// Capabilities describes what the server supports.
type Capabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

// ToolsCapability indicates the server supports tools.
type ToolsCapability struct{}

// ServerInfo holds the server name and version.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool is a tool available on the server.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema ToolInputSchema `json:"inputSchema"` //nolint:tagliatelle
}

// ToolInputSchema is a JSON Schema object describing the tool's parameters.
type ToolInputSchema struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties,omitempty"`
	Required   []string       `json:"required,omitempty"`
}

// CallToolParams holds the parameters for a tools/call request.
type CallToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// CallToolResult holds the result of a tool call.
type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"` //nolint:tagliatelle
}

// Content is a content block in a tool result.
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ListToolsResult holds the result of a tools/list request.
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}
