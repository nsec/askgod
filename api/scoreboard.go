package api

import (
	"time"
)

// URL: /1.0/scoreboard
// Access: guest

// ScoreboardEntry represents a line on the scoreboard
type ScoreboardEntry struct {
	Team           Team      `yaml:"team" json:"team"`
	Value          int64     `yaml:"value" json:"value"`
	LastSubmitTime time.Time `yaml:"last_submit_time" json:"last_submit_time"`
}
