package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/nsec/askgod/api"
)

func (m *MCP) submitFlag(r *http.Request, args map[string]any) CallToolResult {
	flagStr, ok := args["flag"].(string)
	if !ok || flagStr == "" {
		return errorResult("Missing or invalid 'flag' argument")
	}

	body, err := json.Marshal(api.FlagPost{Flag: flagStr, Source: api.SourceMCP})
	if err != nil {
		return errorResult("Internal server error")
	}

	req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, "/1.0/team/flags", bytes.NewReader(body))
	req.RemoteAddr = r.RemoteAddr
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	m.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		return errorResult(strings.TrimSpace(rec.Body.String()))
	}

	var result api.Flag

	err = json.NewDecoder(rec.Body).Decode(&result)
	if err != nil {
		return errorResult("Internal server error")
	}

	msg := fmt.Sprintf("Correct flag! +%d points\n", result.Value)
	if result.ReturnString != "" {
		msg += fmt.Sprintf("Return message: %s\n", result.ReturnString)
	}

	return CallToolResult{Content: []Content{{Type: "text", Text: msg}}}
}

func errorResult(msg string) CallToolResult {
	return CallToolResult{Content: []Content{{Type: "text", Text: msg}}, IsError: true}
}
