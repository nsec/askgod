package api

import (
	"time"
)

// URL: /1.0/scores
// Access: admin

// AdminScore represents a score entry in the database
type AdminScore struct {
	AdminScorePost `yaml:",inline"`

	ID         int64     `yaml:"id" json:"id"`
	SubmitTime time.Time `yaml:"submit_time" json:"submit_time"`
}

// AdminScorePut represents the editable fields of a score entry in the database
type AdminScorePut struct {
	Value int64  `yaml:"value" json:"value"`
	Notes string `yaml:"notes" json:"notes"`
}

// AdminScorePost represents the fields allowed when creating a new score entry
type AdminScorePost struct {
	AdminScorePut `yaml:",inline"`

	TeamID int64 `yaml:"team_id" json:"team_id"`
	FlagID int64 `yaml:"flag_id" json:"flag_id"`
}
