package api

// URL: /1.0/team
// Access: team

// Team represents a team as seen by its members
type Team struct {
	TeamPut `yaml:",inline"`
	ID      int64 `yaml:"id" json:"id"`
}

// TeamPut represents the editable fields of a team as seen by its members
type TeamPut struct {
	Name    string `yaml:"name" json:"name"`
	Country string `yaml:"country" json:"country"`
	Website string `yaml:"website" json:"website"`
}

// URL: /1.0/teams
// Access: admin

// AdminTeam represents a team in the database
type AdminTeam struct {
	AdminTeamPut `yaml:",inline"`

	ID int64 `yaml:"id" json:"id"`
}

// AdminTeamPut represents the editable fields of a team in the database
type AdminTeamPut struct {
	TeamPut `yaml:",inline"`

	Notes   string `yaml:"notes" json:"notes"`
	Subnets string `yaml:"subnets" json:"subnets"`
}

// AdminTeamPost represents the fields allowed when creating a new team
type AdminTeamPost struct {
	AdminTeamPut `yaml:",inline"`
}
