package database

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
)

var dbUpdates = []dbUpdate{
	{version: 1, run: dbUpdateFromV0},
	{version: 2, run: dbUpdateFromV1},
	{version: 3, run: dbUpdateFromV2},
}

type dbUpdate struct {
	version int
	run     func(ctx context.Context, previousVersion int, version int, db *DB) error
}

func (u *dbUpdate) apply(ctx context.Context, currentVersion int, db *DB, logger log15.Logger) error {
	logger.Info("Updating DB schema", log15.Ctx{"current": currentVersion, "update": u.version})

	err := u.run(ctx, currentVersion, u.version, db)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "INSERT INTO schema (version, updated_at) VALUES ($1, $2);", u.version, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func dbUpdateFromV0(ctx context.Context, _ int, _ int, db *DB) error {
	_, err := db.ExecContext(ctx, "ALTER TABLE team ADD COLUMN tags VARCHAR;")

	return err
}

func dbUpdateFromV1(ctx context.Context, _, _ int, db *DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS config (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    value VARCHAR NOT NULL,
    UNIQUE(key)
);
	`)

	return err
}

func dbUpdateFromV2(ctx context.Context, _, _ int, db *DB) error {
	_, err := db.ExecContext(ctx, "ALTER TABLE score ADD COLUMN source VARCHAR NOT NULL DEFAULT 'unknown';")

	return err
}
