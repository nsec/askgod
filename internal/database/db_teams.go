package database

import (
	"database/sql"
	"fmt"
	"net"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

// GetTeams retrieves all the team entries from the database
func (db *DB) GetTeams() ([]api.AdminTeam, error) {
	// Return a list of teams
	resp := []api.AdminTeam{}

	// Query all the teams from the database
	rows, err := db.Query("SELECT id, name, country, website, notes, subnets, tags FROM team ORDER BY id ASC;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the results
	for rows.Next() {
		row := api.AdminTeam{}
		tags := ""

		err := rows.Scan(&row.ID, &row.Name, &row.Country, &row.Website, &row.Notes, &row.Subnets, &tags)
		if err != nil {
			return nil, err
		}

		row.Tags, err = utils.ParseTags(tags)
		if err != nil {
			return nil, err
		}

		resp = append(resp, row)
	}

	// Check for any error that might have happened
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetTeam retrieves a single team entry from the database
func (db *DB) GetTeam(id int64) (*api.AdminTeam, error) {
	// Query the database entry
	row := api.AdminTeam{}
	tags := ""
	err := db.QueryRow("SELECT id, name, country, website, notes, subnets, tags FROM team WHERE id=$1;", id).Scan(
		&row.ID, &row.Name, &row.Country, &row.Website, &row.Notes, &row.Subnets, &tags)
	if err != nil {
		return nil, err
	}

	row.Tags, err = utils.ParseTags(tags)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

// GetTeamForIP retrieves a single team entry for the provided IP
func (db *DB) GetTeamForIP(ip net.IP) (*api.AdminTeam, error) {
	// Get all the teams
	teams, err := db.GetTeams()
	if err != nil {
		return nil, err
	}

	var resp *api.AdminTeam

	for _, team := range teams {
		if team.Subnets == "" {
			continue
		}

		subnets := strings.Split(team.Subnets, ",")
		for _, subnet := range subnets {
			subnet = strings.TrimSpace(subnet)

			_, netSubnet, err := net.ParseCIDR(subnet)
			if err != nil {
				db.logger.Error("Bad subnet", log15.Ctx{"error": err})
				continue
			}

			if netSubnet.Contains(ip) {
				if resp != nil {
					db.logger.Error("More than one team for client IP", log15.Ctx{"ip": ip.String()})
					return nil, fmt.Errorf("More than one team for client IP")
				}

				newTeam := api.AdminTeam(team)
				resp = &newTeam
			}
		}
	}

	if resp == nil {
		return nil, sql.ErrNoRows
	}

	return resp, nil
}

// CreateTeam adds a new team to the database
func (db *DB) CreateTeam(team api.AdminTeamPost) (int64, error) {
	id := int64(-1)

	// Create the database entry
	err := db.QueryRow("INSERT INTO team (name, country, website, notes, subnets, tags) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		team.Name, team.Country, team.Website, team.Notes, team.Subnets, utils.PackTags(team.Tags)).Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

// UpdateTeam updates an existing team
func (db *DB) UpdateTeam(id int64, team api.AdminTeamPut) error {
	// Update the database entry
	result, err := db.Exec("UPDATE team SET name=$1, country=$2, website=$3, notes=$4, subnets=$5, tags=$6 WHERE id=$7;",
		team.Name, team.Country, team.Website, team.Notes, team.Subnets, utils.PackTags(team.Tags), id)
	if err != nil {
		return err
	}

	// Check that a change indeed happened
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteTeam deletes a single team from the database
func (db *DB) DeleteTeam(id int64) error {
	// Delete the database entry
	result, err := db.Exec("DELETE FROM team WHERE id=$1;", id)
	if err != nil {
		return err
	}

	// Check that a change indeed happened
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ClearTeams wipes all team entries from the database
func (db *DB) ClearTeams() error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Wipe the table
	_, err = tx.Exec("DELETE FROM team;")
	if err != nil {
		tx.Rollback()
		return err
	}

	// Reset the sequence
	_, err = tx.Exec("ALTER SEQUENCE team_id_seq RESTART;")
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
