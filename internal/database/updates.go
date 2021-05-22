package database

import (
	"time"

	"github.com/lxc/lxd/shared/log15"
)

var dbUpdates = []dbUpdate{
	{version: 1, run: dbUpdateFromV0},
	{version: 2, run: dbUpdateFromV1},
}

type dbUpdate struct {
	version int
	run     func(previousVersion int, version int, db *DB) error
}

func (u *dbUpdate) apply(currentVersion int, db *DB, logger log15.Logger) error {
	logger.Info("Updating DB schema", log15.Ctx{"current": currentVersion, "update": u.version})

	err := u.run(currentVersion, u.version, db)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO schema (version, updated_at) VALUES ($1, $2);", u.version, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func dbUpdateFromV0(currentVersion int, version int, db *DB) error {
	_, err := db.Exec("ALTER TABLE team ADD COLUMN tags VARCHAR;")
	return err
}

func dbUpdateFromV1(currentVersion int, version int, db *DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS config (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    value VARCHAR NOT NULL,
    UNIQUE(key)
);
	`)
	return err
}
