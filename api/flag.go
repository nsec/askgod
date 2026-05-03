package api

import (
	"slices"
	"time"
)

// Valid values for the Source field on flag and score entries.
const (
	SourceUnknown  = "unknown"
	SourceCLI      = "cli"
	SourceCLIAgent = "cli+agent"
	SourceMCP      = "mcp"
	SourceWeb      = "web"
	SourceWebAgent = "web+agent"
)

// validSources lists every accepted Source value (excluding the empty string,
// which is normalized to SourceUnknown by NormalizeSource).
var validSources = []string{
	SourceUnknown,
	SourceCLI,
	SourceCLIAgent,
	SourceMCP,
	SourceWeb,
	SourceWebAgent,
}

// NormalizeSource validates the provided source and returns the canonical
// value to store. An empty source is translated to SourceUnknown for
// compatibility with older clients. The returned bool indicates whether the
// input was a recognized value.
func NormalizeSource(source string) (string, bool) {
	if source == "" {
		return SourceUnknown, true
	}

	if slices.Contains(validSources, source) {
		return source, true
	}

	return source, false
}

// URL: /1.0/team/flags
// Access: team

// Flag represents a score entry as seen by the team.
type Flag struct {
	FlagPut `yaml:",inline"`

	ID           int64     `json:"id"            yaml:"id"`
	Description  string    `json:"description"   yaml:"description"`
	ReturnString string    `json:"return_string" yaml:"return_string"`
	Value        int64     `json:"value"         yaml:"value"`
	Source       string    `json:"source"        yaml:"source"`
	SubmitTime   time.Time `json:"submit_time"   yaml:"submit_time"`
}

// FlagPut represents the editable fields of a team score entry.
type FlagPut struct {
	Notes string `json:"notes" yaml:"notes"`
}

// FlagPost represents the fields used to submit a new score entry.
//
// Source tracks where the flag submission came from. Valid values are:
//   - ""           => compatibility mechanism, translated to "unknown" internally
//   - "cli"        => human submission through CLI
//   - "cli+agent"  => agent submission through CLI
//   - "mcp"        => submitted through the MCP server
//   - "web"        => human submission through website
//   - "web+agent"  => agent submission through website
type FlagPost struct {
	FlagPut `yaml:",inline"`

	Flag   string `json:"flag"   yaml:"flag"`
	Source string `json:"source" yaml:"source"`
}

// URL: /1.0/flags
// Access: admin

// AdminFlag represents a score entry in the database.
type AdminFlag struct {
	AdminFlagPost `yaml:",inline"`

	ID int64 `json:"id" yaml:"id"`
}

// AdminFlagPut represents the editable fields of a score entry in the database.
type AdminFlagPut struct {
	Flag         string            `json:"flag"          yaml:"flag"`
	Value        int64             `json:"value"         yaml:"value"`
	ReturnString string            `json:"return_string" yaml:"return_string"`
	Description  string            `json:"description"   yaml:"description"`
	Tags         map[string]string `json:"tags"          yaml:"tags"`
}

// AdminFlagPost represents the fields allowed when creating a new score entry.
type AdminFlagPost struct {
	AdminFlagPut `yaml:",inline"`
}
