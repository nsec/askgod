package api

// URL: /1.0
// Access: admin

// Status represents the Askgod configuration.
type Status struct {
	IsAdmin bool `json:"is_admin" yaml:"is_admin"`
	IsTeam  bool `json:"is_team"  yaml:"is_team"`
	IsGuest bool `json:"is_guest" yaml:"is_guest"`

	EventName string `json:"event_name" yaml:"event_name"`

	Flags StatusFlags `json:"flags" yaml:"flags"`
}

// StatusFlags is a number of configuration flags that are useful to clients.
type StatusFlags struct {
	TeamSelfRegister bool `json:"team_self_register" yaml:"team_self_register"`
	TeamSelfUpdate   bool `json:"team_self_update"   yaml:"team_self_update"`

	BoardReadOnly   bool `json:"board_read_only"   yaml:"board_read_only"`
	BoardHideOthers bool `json:"board_hide_others" yaml:"board_hide_others"`
}
