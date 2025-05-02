package database

import (
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

// GetTeamPoints returns the current total for the team.
func (db *DB) GetTeamPoints(teamid int64) (int64, error) {
	total := int64(0)

	// Get the total
	err := db.QueryRow("SELECT COALESCE(SUM(score.value), 0) AS points FROM score WHERE teamid=$1", teamid).Scan(&total)
	if err != nil {
		return -1, err
	}

	return total, nil
}

// GetTeamFlags retrieves all the score entries for the team.
func (db *DB) GetTeamFlags(teamid int64) ([]api.Flag, error) {
	// Return a list of score entries
	resp := []api.Flag{}

	// Query all the scores from the database
	rows, err := db.Query("SELECT score.flagid, flag.description, score.value, score.notes, score.submit_time, flag.return_string FROM score LEFT JOIN flag ON flag.id=score.flagid WHERE score.teamid=$1 ORDER BY score.submit_time ASC;", teamid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the results
	for rows.Next() {
		row := api.Flag{}

		err := rows.Scan(&row.ID, &row.Description, &row.Value, &row.Notes, &row.SubmitTime, &row.ReturnString)
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

// GetTeamFlag retrieves a single score entry for the team.
func (db *DB) GetTeamFlag(teamid int64, id int64) (*api.Flag, error) {
	// Return a list of score entries
	resp := api.Flag{}

	// Query all the scores from the database
	err := db.QueryRow("SELECT score.flagid, flag.description, score.value, score.notes, score.submit_time, flag.return_string FROM score LEFT JOIN flag ON flag.id=score.flagid WHERE score.teamid=$1 AND score.flagid=$2 ORDER BY score.submit_time ASC;", teamid, id).Scan(
		&resp.ID, &resp.Description, &resp.Value, &resp.Notes, &resp.SubmitTime, &resp.ReturnString)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateTeamFlag updates a single score entry for the team.
func (db *DB) UpdateTeamFlag(teamid int64, id int64, flag api.FlagPut) error {
	// Update the database entry
	result, err := db.Exec("UPDATE score SET notes=$1 WHERE teamid=$2 AND flagid=$3;",
		flag.Notes, teamid, id)
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

// SubmitTeamFlag validates a submitted flag and adds it to the database.
func (db *DB) SubmitTeamFlag(teamid int64, flag api.FlagPost) (*api.Flag, *api.AdminFlag, error) {
	// Query the database entry
	row := api.AdminFlag{}
	tags := ""
	err := db.QueryRow("SELECT id, flag, value, return_string, description, tags FROM flag WHERE LOWER(flag)=LOWER($1);", flag.Flag).Scan(
		&row.ID, &row.Flag, &row.Value, &row.ReturnString, &row.Description, &tags)
	if err != nil {
		return nil, nil, err
	}

	row.Tags, err = utils.ParseTags(tags)
	if err != nil {
		return nil, nil, err
	}

	// Check if already submitted
	id := int64(-1)
	err = db.QueryRow("SELECT id FROM score WHERE teamid=$1 AND flagid=$2;", teamid, row.ID).Scan(&id)
	if err == nil {
		return nil, &row, os.ErrExist
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	// Add the flag
	id = -1
	err = db.QueryRow("INSERT INTO score (teamid, flagid, value, notes, submit_time) VALUES ($1, $2, $3, $4, $5) RETURNING id;",
		teamid, row.ID, row.Value, flag.Notes, time.Now()).Scan(&id)
	if err != nil {
		return nil, nil, err
	}

	// Query the new entry
	result := api.Flag{}
	err = db.QueryRow("SELECT score.flagid, flag.description, score.value, score.notes, score.submit_time, flag.return_string FROM score LEFT JOIN flag ON flag.id=score.flagid WHERE score.id=$1;", id).Scan(
		&result.ID, &result.Description, &result.Value, &result.Notes, &result.SubmitTime, &result.ReturnString)
	if err != nil {
		return nil, nil, err
	}

	return &result, &row, nil
}

// GetScores retrieves all the score entries from the database.
func (db *DB) GetScores() ([]api.AdminScore, error) {
	// Return a list of score entries
	resp := []api.AdminScore{}

	// Query all the scores from the database
	rows, err := db.Query("SELECT id, teamid, flagid, value, notes, submit_time FROM score ORDER BY id ASC;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the results
	for rows.Next() {
		row := api.AdminScore{}

		err := rows.Scan(&row.ID, &row.TeamID, &row.FlagID, &row.Value, &row.Notes, &row.SubmitTime)
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

// GetScore retrieves a single score entry from the database.
func (db *DB) GetScore(id int64) (*api.AdminScore, error) {
	// Query the database entry
	row := api.AdminScore{}
	err := db.QueryRow("SELECT id, teamid, flagid, value, notes, submit_time FROM score WHERE id=$1;", id).Scan(
		&row.ID, &row.TeamID, &row.FlagID, &row.Value, &row.Notes, &row.SubmitTime)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

// CreateScore adds a new score entry to the database.
func (db *DB) CreateScore(score api.AdminScorePost) (int64, error) {
	id := int64(-1)

	// Create the database entry
	err := db.QueryRow("INSERT INTO score (teamid, flagid, value, notes, submit_time) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		score.TeamID, score.FlagID, score.Value, score.Notes, time.Now()).Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

// UpdateScore updates an existing score entry.
func (db *DB) UpdateScore(id int64, score api.AdminScorePut) error {
	// Update the database entry
	result, err := db.Exec("UPDATE score SET value=$1, notes=$2 WHERE id=$3;",
		score.Value, score.Notes, id)
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

// DeleteScore deletes a single score entry from the database.
func (db *DB) DeleteScore(id int64) error {
	// Delete the database entry
	result, err := db.Exec("DELETE FROM score WHERE id=$1;", id)
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

// ClearScores wipes all score entries from the database.
func (db *DB) ClearScores() error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Wipe the table
	_, err = tx.Exec("DELETE FROM score;")
	if err != nil {
		errRollback := tx.Rollback()
		if err != nil {
			return errRollback
		}

		return err
	}

	// Reset the sequence
	_, err = tx.Exec("ALTER SEQUENCE score_id_seq RESTART;")
	if err != nil {
		errRollback := tx.Rollback()
		if err != nil {
			return errRollback
		}

		return err
	}

	// Commit
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
