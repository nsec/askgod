package api

import (
	"time"
)

// URL: /1.0/team/flags
// Access: team

// Flag represents a score entry as seen by the team.
type Flag struct {
	FlagPut `yaml:",inline"`

	ID           int64     `json:"id"            yaml:"id"`
	Description  string    `json:"description"   yaml:"description"`
	ReturnString string    `json:"return_string" yaml:"return_string"`
	Value        int64     `json:"value"         yaml:"value"`
	SubmitTime   time.Time `json:"submit_time"   yaml:"submit_time"`
}

// FlagPut represents the editable fields of a team score entry.
type FlagPut struct {
	Notes string `json:"notes" yaml:"notes"`
}

// FlagPost represents the fields used to submit a new score entry.
type FlagPost struct {
	FlagPut `yaml:",inline"`

	Flag string `json:"flag" yaml:"flag"`
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
