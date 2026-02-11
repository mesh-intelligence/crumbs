// Package sqlite implements the SQLite backend for the Crumbs system.
// Implements: prd002-sqlite-backend (R3: SQLite Schema, R3.2: crumbs table, R3.3: indexes);
//             docs/ARCHITECTURE § SQLite Backend.
package sqlite

import "database/sql"

// createSchema creates the crumbs table and indexes in the SQLite database.
// Other tables (trails, properties, metadata, links, stashes) are added by
// future tasks following the same pattern.
func createSchema(db *sql.DB) error {
	_, err := db.Exec(schemaCrumbs)
	if err != nil {
		return err
	}
	_, err = db.Exec(indexCrumbsState)
	return err
}

// SQL schema for the crumbs table (prd002-sqlite-backend R3.2).
const schemaCrumbs = `CREATE TABLE IF NOT EXISTS crumbs (
	crumb_id   TEXT PRIMARY KEY,
	name       TEXT NOT NULL,
	state      TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);`

// Index for common crumbs queries (prd002-sqlite-backend R3.3).
const indexCrumbsState = `CREATE INDEX IF NOT EXISTS idx_crumbs_state ON crumbs(state);`
