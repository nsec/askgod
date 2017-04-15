package api

import (
	"encoding/json"
	"time"
)

// URL: /1.0/events
// Access: various

// Event represents an event entry (over websocket)
type Event struct {
	Server    string          `yaml:"server" json:"server"`
	Type      string          `yaml:"type" json:"type"`
	Timestamp time.Time       `yaml:"timestamp" json:"timestamp"`
	Metadata  json.RawMessage `yaml:"metadata" json:"metadata"`
}

// EventLogging represents a logging type event entry (admin only)
type EventLogging struct {
	Message string            `yaml:"message" json:"message"`
	Level   string            `yaml:"level" json:"level"`
	Context map[string]string `yaml:"context" json:"context"`
}

// EventFlag represents a flag submission event entry (admin only)
type EventFlag struct {
	Team AdminTeam  `yaml:"team" json:"team"`
	Flag *AdminFlag `yaml:"flag" json:"flag"`

	Input  string `yaml:"input" json:"input"`
	Value  int64  `yaml:"value" json:"value"`
	Result string `yaml:"result" json:"result"`
}

// EventTimeline represents a change to the timeline (guest only)
type EventTimeline struct {
	Team  Team               `yaml:"team" json:"team"`
	Score TimelineEntryScore `yaml:"score" json:"score"`
}

// EventTeam represents a team change event entry (guest only)
type EventTeam Team
