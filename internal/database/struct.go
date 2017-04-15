package database

import (
	"database/sql"

	"gopkg.in/inconshreveable/log15.v2"
)

// DB represents the Askgod database
type DB struct {
	*sql.DB

	logger log15.Logger
}
