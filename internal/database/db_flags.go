package database

import (
	"database/sql"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

// GetFlags retrieves all the flag entries from the database
func (db *DB) GetFlags() ([]api.AdminFlag, error) {
	// Return a list of flags
	resp := []api.AdminFlag{}

	// Query all the flags from the database
	rows, err := db.Query("SELECT id, flag, value, return_string, description, tags FROM flag ORDER BY id ASC;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the results
	for rows.Next() {
		row := api.AdminFlag{}
		tags := ""

		err := rows.Scan(&row.ID, &row.Flag, &row.Value, &row.ReturnString, &row.Description, &tags)
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

// GetFlag retrieves a single flag entry from the database
func (db *DB) GetFlag(id int64) (*api.AdminFlag, error) {
	// Query the database entry
	row := api.AdminFlag{}
	tags := ""

	err := db.QueryRow("SELECT id, flag, value, return_string, description, tags FROM flag WHERE id=$1;", id).Scan(
		&row.ID, &row.Flag, &row.Value, &row.ReturnString, &row.Description, &tags)
	if err != nil {
		return nil, err
	}

	row.Tags, err = utils.ParseTags(tags)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

// CreateFlag adds a new flag to the database
func (db *DB) CreateFlag(flag api.AdminFlagPost) (int64, error) {
	id := int64(-1)

	// Create the database entry
	err := db.QueryRow("INSERT INTO flag (flag, value, return_string, description, tags) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		flag.Flag, flag.Value, flag.ReturnString, flag.Description, utils.PackTags(flag.Tags)).Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

// UpdateFlag updates an existing flag
func (db *DB) UpdateFlag(id int64, flag api.AdminFlagPut) error {
	// Update the database entry
	result, err := db.Exec("UPDATE flag SET flag=$1, value=$2, return_string=$3, description=$4, tags=$5 WHERE id=$6;",
		flag.Flag, flag.Value, flag.ReturnString, flag.Description, utils.PackTags(flag.Tags), id)
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

// DeleteFlag deletes a single flag from the database
func (db *DB) DeleteFlag(id int64) error {
	// Delete the database entry
	result, err := db.Exec("DELETE FROM flag WHERE id=$1;", id)
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

// ClearFlags wipes all flag entries from the database
func (db *DB) ClearFlags() error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Wipe the table
	_, err = tx.Exec("DELETE FROM flag;")
	if err != nil {
		tx.Rollback()
		return err
	}

	// Reset the sequence
	_, err = tx.Exec("ALTER SEQUENCE flag_id_seq RESTART;")
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
