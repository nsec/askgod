// Package mcp implements an MCP server using Streamable HTTP transport.
package mcp

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/inconshreveable/log15"
)

// MCP implements an MCP server using Streamable HTTP transport.
type MCP struct {
	logger  log15.Logger
	handler http.Handler
}

// NewMCP creates a new MCP server instance.
func NewMCP(handler http.Handler, logger log15.Logger) *MCP {
	return &MCP{
		handler: handler,
		logger:  logger,
	}
}

// ServeHTTP handles incoming MCP requests over Streamable HTTP.
func (m *MCP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)

		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)

		return
	}

	var req Request

	err = json.Unmarshal(body, &req)
	if err != nil {
		m.writeError(w, nil, -32700, "Parse error")

		return
	}

	// Notifications have no id — respond with 202 Accepted.
	if req.ID == nil {
		w.WriteHeader(http.StatusAccepted)

		return
	}

	var id any

	err = json.Unmarshal(*req.ID, &id)
	if err != nil {
		id = nil
	}

	switch req.Method {
	case "initialize":
		m.handleInitialize(w, id)
	case "tools/list":
		m.handleToolsList(w, id)
	case "tools/call":
		m.handleToolsCall(w, r, id, req.Params)
	default:
		m.writeError(w, id, -32601, "Method not found")
	}
}

func (m *MCP) handleInitialize(w http.ResponseWriter, id any) {
	result := InitializeResult{
		ProtocolVersion: "2025-03-26",
		Capabilities: Capabilities{
			Tools: &ToolsCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "askgod",
			Version: "1.0.0",
		},
	}

	m.writeResult(w, id, result)
}

func (m *MCP) handleToolsList(w http.ResponseWriter, id any) {
	result := ListToolsResult{
		Tools: []Tool{
			{
				Name:        "submit_flag",
				Description: "Submit a CTF flag for your team. Returns whether the flag was valid, already submitted, or invalid.",
				InputSchema: ToolInputSchema{
					Type: "object",
					Properties: map[string]any{
						"flag": map[string]any{
							"type":        "string",
							"description": "The flag string to submit",
						},
					},
					Required: []string{"flag"},
				},
			},
		},
	}

	m.writeResult(w, id, result)
}

func (m *MCP) handleToolsCall(w http.ResponseWriter, r *http.Request, id any, params json.RawMessage) {
	var p CallToolParams

	err := json.Unmarshal(params, &p)
	if err != nil {
		m.writeError(w, id, -32602, "Invalid params")

		return
	}

	var result CallToolResult

	switch p.Name {
	case "submit_flag":
		result = m.submitFlag(r, p.Arguments)
	default:
		m.writeError(w, id, -32602, "Unknown tool: "+p.Name)

		return
	}

	m.writeResult(w, id, result)
}

func (m *MCP) writeResult(w http.ResponseWriter, id any, result any) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		m.logger.Error("Failed to encode response", log15.Ctx{"error": err})
	}
}

func (m *MCP) writeError(w http.ResponseWriter, id any, code int, message string) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		m.logger.Error("Failed to encode error response", log15.Ctx{"error": err})
	}
}
