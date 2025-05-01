package api

import (
	"time"
)

// URL: /1.0/scores
// Access: admin

// AdminScore represents a score entry in the database.
type AdminScore struct {
	AdminScorePost `yaml:",inline"`

	ID         int64     `json:"id"          yaml:"id"`
	SubmitTime time.Time `json:"submit_time" yaml:"submit_time"`
}

// AdminScorePut represents the editable fields of a score entry in the database.
type AdminScorePut struct {
	Value int64  `json:"value" yaml:"value"`
	Notes string `json:"notes" yaml:"notes"`
}

// AdminScorePost represents the fields allowed when creating a new score entry.
type AdminScorePost struct {
	AdminScorePut `yaml:",inline"`

	TeamID int64 `json:"team_id" yaml:"team_id"`
	FlagID int64 `json:"flag_id" yaml:"flag_id"`
}
