package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/inconshreveable/log15"
	// Import the postgres DB driver.
	_ "github.com/lib/pq"
)

// Connect sets up the database connection and returns a DB struct.
func Connect(driver string, host string, username string, password string, database string, connections int, tls bool, logger log15.Logger) (*DB, error) {
	// We only support postgres for now
	if driver != "postgres" {
		return nil, errors.New("database driver not supported")
	}

	// Connect to the backend
	logger.Info("Connecting to the database", log15.Ctx{
		"driver":      driver,
		"host":        host,
		"username":    username,
		"database":    database,
		"connections": connections,
		"tls":         tls,
	})

	sslmode := "require"

	if !tls {
		sslmode = "disable"
	}
	psqlDB, err := sql.Open(driver, fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s", host, username, password, database, sslmode))
	if err != nil {
		return nil, err
	}

	// Setup the DB struct
	db := DB{
		DB:     psqlDB,
		logger: logger,
	}

	// We don't want multiple clients during setup
	db.SetMaxOpenConns(1)

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Check if the database is initialized
	_, err = db.GetCurrentSchema()
	if err != nil {
		// Lets assume that the database is empty and create it
		err = db.createDatabase()
		if err != nil {
			return nil, err
		}
	}

	// Apply schema updates
	err = db.updateDatabase()
	if err != nil {
		return nil, err
	}

	// Set the connection limit for the DB pool
	db.SetMaxOpenConns(connections)

	return &db, nil
}
