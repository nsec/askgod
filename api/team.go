package api

// URL: /1.0/team
// Access: team

// Team represents a team as seen by its members.
type Team struct {
	TeamPut `yaml:",inline"`
	ID      int64 `json:"id"      yaml:"id"`
}

// TeamPut represents the editable fields of a team as seen by its members.
type TeamPut struct {
	Name    string `json:"name"    yaml:"name"`
	Country string `json:"country" yaml:"country"`
	Website string `json:"website" yaml:"website"`
}

// URL: /1.0/teams
// Access: admin

// AdminTeam represents a team in the database.
type AdminTeam struct {
	AdminTeamPut `yaml:",inline"`

	ID int64 `json:"id" yaml:"id"`
}

// AdminTeamPut represents the editable fields of a team in the database.
type AdminTeamPut struct {
	TeamPut `yaml:",inline"`

	Notes   string            `json:"notes"   yaml:"notes"`
	Subnets string            `json:"subnets" yaml:"subnets"`
	Tags    map[string]string `json:"tags"    yaml:"tags"`
}

// AdminTeamPost represents the fields allowed when creating a new team.
type AdminTeamPost struct {
	AdminTeamPut `yaml:",inline"`
}
