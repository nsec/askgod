package api

// URL: /1.0
// Access: admin

// Status represents the Askgod configuration
type Status struct {
	IsAdmin bool `yaml:"is_admin" json:"is_admin"`
	IsTeam  bool `yaml:"is_team" json:"is_team"`
	IsGuest bool `yaml:"is_guest" json:"is_guest"`

	EventName string `yaml:"event_name" json:"event_name"`

	Flags StatusFlags `yaml:"flags" json:"flags"`
}

// StatusFlags is a number of configuration flags that are useful to clients
type StatusFlags struct {
	TeamSelfRegister bool `yaml:"team_self_register" json:"team_self_register"`
	TeamSelfUpdate   bool `yaml:"team_self_update" json:"team_self_update"`

	BoardReadOnly   bool `yaml:"board_read_only" json:"board_read_only"`
	BoardHideOthers bool `yaml:"board_hide_others" json:"board_hide_others"`
}
