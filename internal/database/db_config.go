package database

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nsec/askgod/api"
)

var ErrEmptyConfig = errors.New("no configuration in database")

// GetConfig retrieves the configuration
func (db *DB) GetConfig() (*api.ConfigPut, error) {
	// Query all the teams from the database
	rows, err := db.Query("SELECT key, value FROM config;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the results
	dbConfig := map[string]string{}
	count := 0
	for rows.Next() {
		count++
		key := ""
		value := ""

		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}

		dbConfig[key] = value
	}

	if count == 0 {
		return nil, ErrEmptyConfig
	}

	// Apply mapping
	resp := api.ConfigPut{
		Scoring: api.ConfigScoring{
			EventName:  dbConfig["scoring.event_name"],
			HideOthers: dbConfig["scoring.hide_others"] == "true",
			ReadOnly:   dbConfig["scoring.read_only"] == "true",
		},
		Teams: api.ConfigTeams{
			SelfRegister: dbConfig["teams.self_register"] == "true",
			SelfUpdate:   dbConfig["teams.self_update"] == "true",
			Hidden:       strings.Split(dbConfig["teams.hidden"], ","),
		},
		Subnets: api.ConfigSubnets{
			Admins: strings.Split(dbConfig["subnets.admins"], ","),
			Teams:  strings.Split(dbConfig["subnets.teams"], ","),
			Guests: strings.Split(dbConfig["subnets.guests"], ","),
		},
	}

	// Check for any error that might have happened
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateConfig updates the configuration
func (db *DB) UpdateConfig(config api.ConfigPut) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Wipe the table
	_, err = tx.Exec("DELETE FROM config;")
	if err != nil {
		tx.Rollback()
		return err
	}

	// Setup mapping
	dbConfig := map[string]string{
		"scoring.event_name":  config.Scoring.EventName,
		"scoring.hide_others": fmt.Sprintf("%v", config.Scoring.HideOthers),
		"scoring.read_only":   fmt.Sprintf("%v", config.Scoring.ReadOnly),
		"teams.self_register": fmt.Sprintf("%v", config.Teams.SelfRegister),
		"teams.self_update":   fmt.Sprintf("%v", config.Teams.SelfUpdate),
		"teams.hidden":        strings.Join(config.Teams.Hidden, ","),
		"subnets.admins":      strings.Join(config.Subnets.Admins, ","),
		"subnets.teams":       strings.Join(config.Subnets.Teams, ","),
		"subnets.guests":      strings.Join(config.Subnets.Guests, ","),
	}

	// Insert the new config
	for k, v := range dbConfig {
		_, err = tx.Exec("INSERT INTO config (key, value) VALUES ($1, $2);", k, v)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
