// Package sqlite implements the SQLite backend for the Crumbs cupboard.
// Implements: prd002-sqlite-backend; docs/ARCHITECTURE § System Components.
package sqlite

import (
	"database/sql"
	"fmt"
)

// createSchema creates all tables and indexes in the SQLite database.
// Each table matches the JSONL file structure; SQLite serves as a query cache
// rebuilt from JSONL on every startup.
// Implements: prd002-sqlite-backend R3.
func createSchema(db *sql.DB) error {
	stmts := []string{
		// Crumbs table (prd002 R3.2, R14.2).
		`CREATE TABLE IF NOT EXISTS crumbs (
			crumb_id   TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			state      TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,

		// Trails table (prd002 R3.2, R14.3).
		`CREATE TABLE IF NOT EXISTS trails (
			trail_id     TEXT PRIMARY KEY,
			state        TEXT NOT NULL,
			created_at   TEXT NOT NULL,
			completed_at TEXT
		)`,

		// Properties table (prd002 R3.2, R14.4).
		`CREATE TABLE IF NOT EXISTS properties (
			property_id TEXT PRIMARY KEY,
			name        TEXT NOT NULL UNIQUE,
			description TEXT,
			value_type  TEXT NOT NULL,
			created_at  TEXT NOT NULL
		)`,

		// Categories table (prd002 R3.2).
		`CREATE TABLE IF NOT EXISTS categories (
			category_id TEXT PRIMARY KEY,
			property_id TEXT NOT NULL,
			name        TEXT NOT NULL,
			ordinal     INTEGER NOT NULL,
			FOREIGN KEY (property_id) REFERENCES properties(property_id)
		)`,

		// Crumb properties junction table (prd002 R3.2, R3.4).
		`CREATE TABLE IF NOT EXISTS crumb_properties (
			crumb_id    TEXT NOT NULL,
			property_id TEXT NOT NULL,
			value       TEXT NOT NULL,
			PRIMARY KEY (crumb_id, property_id),
			FOREIGN KEY (crumb_id) REFERENCES crumbs(crumb_id),
			FOREIGN KEY (property_id) REFERENCES properties(property_id)
		)`,

		// Metadata table (prd002 R3.2, R14.5).
		`CREATE TABLE IF NOT EXISTS metadata (
			metadata_id TEXT PRIMARY KEY,
			table_name  TEXT NOT NULL,
			crumb_id    TEXT NOT NULL,
			property_id TEXT,
			content     TEXT NOT NULL,
			created_at  TEXT NOT NULL,
			FOREIGN KEY (crumb_id) REFERENCES crumbs(crumb_id)
		)`,

		// Links table (prd002 R3.2, R14.6).
		`CREATE TABLE IF NOT EXISTS links (
			link_id    TEXT PRIMARY KEY,
			link_type  TEXT NOT NULL,
			from_id    TEXT NOT NULL,
			to_id      TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,

		// Stashes table (prd002 R3.2, R14.7).
		`CREATE TABLE IF NOT EXISTS stashes (
			stash_id       TEXT PRIMARY KEY,
			name           TEXT NOT NULL,
			stash_type     TEXT NOT NULL,
			value          TEXT NOT NULL,
			version        INTEGER NOT NULL,
			created_at     TEXT NOT NULL,
			updated_at     TEXT NOT NULL,
			last_operation TEXT NOT NULL,
			changed_by     TEXT
		)`,

		// Stash history table (prd002 R3.2).
		`CREATE TABLE IF NOT EXISTS stash_history (
			history_id TEXT PRIMARY KEY,
			stash_id   TEXT NOT NULL,
			version    INTEGER NOT NULL,
			value      TEXT NOT NULL,
			operation  TEXT NOT NULL,
			changed_by TEXT,
			created_at TEXT NOT NULL,
			FOREIGN KEY (stash_id) REFERENCES stashes(stash_id)
		)`,

		// Indexes (prd002 R3.3).
		`CREATE INDEX IF NOT EXISTS idx_crumbs_state ON crumbs(state)`,
		`CREATE INDEX IF NOT EXISTS idx_trails_state ON trails(state)`,
		`CREATE INDEX IF NOT EXISTS idx_links_type_from ON links(link_type, from_id)`,
		`CREATE INDEX IF NOT EXISTS idx_links_type_to ON links(link_type, to_id)`,
		`CREATE INDEX IF NOT EXISTS idx_crumb_properties_crumb ON crumb_properties(crumb_id)`,
		`CREATE INDEX IF NOT EXISTS idx_crumb_properties_property ON crumb_properties(property_id)`,
		`CREATE INDEX IF NOT EXISTS idx_metadata_crumb ON metadata(crumb_id)`,
		`CREATE INDEX IF NOT EXISTS idx_metadata_table ON metadata(table_name)`,
		`CREATE INDEX IF NOT EXISTS idx_categories_property ON categories(property_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stashes_name ON stashes(name)`,
		`CREATE INDEX IF NOT EXISTS idx_stash_history_stash ON stash_history(stash_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stash_history_version ON stash_history(stash_id, version)`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("creating schema: %w", err)
		}
	}
	return nil
}
