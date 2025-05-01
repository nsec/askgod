package api

import (
	"encoding/json"
	"time"
)

// URL: /1.0/events
// Access: various

// Event represents an event entry (over websocket).
type Event struct {
	Server    string          `json:"server"    yaml:"server"`
	Type      string          `json:"type"      yaml:"type"`
	Timestamp time.Time       `json:"timestamp" yaml:"timestamp"`
	Metadata  json.RawMessage `json:"metadata"  yaml:"metadata"`
}

// EventLogging represents a logging type event entry (admin only).
type EventLogging struct {
	Message string            `json:"message" yaml:"message"`
	Level   string            `json:"level"   yaml:"level"`
	Context map[string]string `json:"context" yaml:"context"`
}

// EventFlag represents a flag submission event entry (admin only).
type EventFlag struct {
	Team AdminTeam  `json:"team" yaml:"team"`
	Flag *AdminFlag `json:"flag" yaml:"flag"`

	Input string `json:"input" yaml:"input"`
	Value int64  `json:"value" yaml:"value"`
	Type  string `json:"type"  yaml:"type"`
}

// EventTimeline represents a change to the timeline (guest only).
type EventTimeline struct {
	TeamID int64               `json:"teamid" yaml:"teamid"`
	Team   *TeamPut            `json:"team"   yaml:"team"`
	Score  *TimelineEntryScore `json:"score"  yaml:"score"`
	Type   string              `json:"type"   yaml:"type"`
}

// EventInternal represents an internal syncronisation event.
type EventInternal struct {
	Type string `json:"type" yaml:"type"`
}
