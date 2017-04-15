package api

import (
	"time"
)

// URL: /1.0/team/flags
// Access: team

// Flag represents a score entry as seen by the team
type Flag struct {
	FlagPost `yaml:",inline"`

	ID           int64     `yaml:"id" json:"id"`
	ReturnString string    `yaml:"return_string" json:"return_string"`
	Value        int64     `yaml:"value" json:"value"`
	SubmitTime   time.Time `yaml:"submit_time" json:"submit_time"`
}

// FlagPut represents the editable fields of a team score entry
type FlagPut struct {
	Notes string `yaml:"notes" json:"notes"`
}

// FlagPost represents the fields used to submit a new score entry
type FlagPost struct {
	FlagPut `yaml:",inline"`

	Flag string `yaml:"flag" json:"flag"`
}

// URL: /1.0/flags
// Access: admin

// AdminFlag represents a score entry in the database
type AdminFlag struct {
	AdminFlagPost `yaml:",inline"`

	ID int64 `yaml:"id" json:"id"`
}

// AdminFlagPut represents the editable fields of a score entry in the database
type AdminFlagPut struct {
	Flag         string `yaml:"flag" json:"flag"`
	Value        int64  `yaml:"value" json:"value"`
	ReturnString string `yaml:"return_string" json:"return_string"`
	Description  string `yaml:"description" json:"description"`
	Tags         string `yaml:"tags" json:"tags"`
}

// AdminFlagPost represents the fields allowed when creating a new score entry
type AdminFlagPost struct {
	AdminFlagPut `yaml:",inline"`
}
