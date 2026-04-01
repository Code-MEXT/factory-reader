package db

import "context"

func (d *DB) Migrate(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS connections (
		id            SERIAL PRIMARY KEY,
		name          TEXT NOT NULL,
		protocol      TEXT NOT NULL,
		host          TEXT NOT NULL,
		port          INTEGER NOT NULL,
		topic         TEXT DEFAULT '',
		node_id       TEXT DEFAULT '',
		unit_id       INTEGER DEFAULT 1,
		rack          INTEGER DEFAULT 0,
		slot          INTEGER DEFAULT 1,
		db_number     INTEGER DEFAULT 1,
		start_address INTEGER DEFAULT 0,
		quantity      INTEGER DEFAULT 1,
		created_at    TIMESTAMPTZ DEFAULT NOW()
	);`
	_, err := d.Pool.Exec(ctx, query)
	return err
}
