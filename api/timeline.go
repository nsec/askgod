package api

import (
	"time"
)

// URL: /1.0/timeline
// Access: guest

// TimelineEntry represents the timeline for a team.
type TimelineEntry struct {
	Team  Team                 `json:"team"  yaml:"team"`
	Score []TimelineEntryScore `json:"score" yaml:"score"`
}

// TimelineEntryScore represents a score entry for a team.
type TimelineEntryScore struct {
	SubmitTime time.Time `json:"submit_time" yaml:"submit_time"`
	Value      int64     `json:"value"       yaml:"value"`
	Total      int64     `json:"total"       yaml:"total"`
}
