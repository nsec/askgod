package api

import (
	"time"
)

// URL: /1.0/scoreboard
// Access: guest

// ScoreboardEntry represents a line on the scoreboard.
type ScoreboardEntry struct {
	Team           Team      `json:"team"             yaml:"team"`
	Value          int64     `json:"value"            yaml:"value"`
	LastSubmitTime time.Time `json:"last_submit_time" yaml:"last_submit_time"`
}
