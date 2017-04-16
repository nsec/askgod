package database

import (
	"time"
)

const schema string = `
CREATE TABLE IF NOT EXISTS flag (
    id SERIAL PRIMARY KEY,
    flag VARCHAR,
    value INTEGER NOT NULL DEFAULT 0,
    return_string VARCHAR,
    description VARCHAR,
    tags VARCHAR,
    UNIQUE(flag)
);

CREATE TABLE IF NOT EXISTS team (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    country VARCHAR(2),
    website VARCHAR(255),
    notes VARCHAR,
    subnets VARCHAR,
    tags VARCHAR
);

CREATE TABLE IF NOT EXISTS score (
    id SERIAL PRIMARY KEY,
    teamid INTEGER NOT NULL,
    flagid INTEGER NOT NULL,
    value INTEGER NOT NULL DEFAULT 0,
    submit_time TIMESTAMP WITH TIME ZONE,
    notes VARCHAR,
    FOREIGN KEY (teamid) REFERENCES team (id) ON DELETE CASCADE,
    FOREIGN KEY (flagid) REFERENCES flag (id) ON DELETE CASCADE,
    UNIQUE(teamid, flagid)
);

CREATE TABLE IF NOT EXISTS schema (
    id SERIAL PRIMARY KEY,
    version INTEGER,
    updated_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(version)
);
`

// GetCurrentSchema returns the current DB schema version
func (db *DB) GetCurrentSchema() (int, error) {
	version := -1
	err := db.QueryRow("SELECT max(version) FROM schema;").Scan(&version)
	if err != nil {
		return -1, err
	}

	return version, nil
}

func (db *DB) getLatestSchema() int {
	if len(dbUpdates) == 0 {
		return 0
	}

	return dbUpdates[len(dbUpdates)-1].version
}

func (db *DB) createDatabase() error {
	// Setup a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Apply the latest schema
	db.logger.Info("Creating initial database schema")
	_, err = tx.Exec(schema)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Create the initial schema entry
	db.logger.Info("Inserting initial schema entry")
	_, err = tx.Exec("INSERT INTO schema (version, updated_at) VALUES ($1, $2);", db.getLatestSchema(), time.Now())
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

func (db *DB) updateDatabase() error {
	current, err := db.GetCurrentSchema()
	if err != nil {
		return err
	}

	for _, update := range dbUpdates {
		if update.version <= current {
			continue
		}

		err := update.apply(current, db, db.logger)
		if err != nil {
			return err
		}

		current = update.version
	}

	return nil
}
