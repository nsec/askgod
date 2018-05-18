package database

import (
	"database/sql"

	"github.com/lxc/lxd/shared/log15"
)

// DB represents the Askgod database
type DB struct {
	*sql.DB

	logger log15.Logger
}
