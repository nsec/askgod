package api

import (
	"time"
)

// URL: /1.0/timeline
// Access: guest

// TimelineEntry represents the timeline for a team
type TimelineEntry struct {
	Team  Team                 `yaml:"team" json:"team"`
	Score []TimelineEntryScore `yaml:"score" json:"score"`
}

// TimelineEntryScore represents a score entry for a team
type TimelineEntryScore struct {
	SubmitTime time.Time `yaml:"submit_time" json:"submit_time"`
	Value      int64     `yaml:"value" json:"value"`
	Total      int64     `yaml:"total" json:"total"`
}
