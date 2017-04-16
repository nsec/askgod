package database

import (
	"gopkg.in/inconshreveable/log15.v2"
)

var dbUpdates = []dbUpdate{
	{version: 1, run: dbUpdateFromV0},
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

	_, err = db.Exec("INSERT INTO schema (version, updated_at) VALUES (?, strftime(\"%s\"));", u.version)
	if err != nil {
		return err
	}

	return nil
}

func dbUpdateFromV0(currentVersion int, version int, db *DB) error {
	_, err := db.Exec("ALTER TABLE team ADD COLUMN tags VARCHAR;")
	return err
}
