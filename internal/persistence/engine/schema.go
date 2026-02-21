// Package engine implements the SQLite storage backend for Crumbs.
//
// Implements: prd002-sqlite-backend R3 (SQLite Schema)
package engine

import "database/sql"

// Schema DDL: 9 tables mirroring the JSONL file structure (R3.2).
const schemaSQL = `
CREATE TABLE crumbs (
    crumb_id   TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    state      TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE trails (
    trail_id     TEXT PRIMARY KEY,
    state        TEXT NOT NULL,
    created_at   TEXT NOT NULL,
    completed_at TEXT
);

CREATE TABLE links (
    link_id    TEXT PRIMARY KEY,
    link_type  TEXT NOT NULL,
    from_id    TEXT NOT NULL,
    to_id      TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE properties (
    property_id TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    value_type  TEXT NOT NULL,
    created_at  TEXT NOT NULL
);

CREATE TABLE categories (
    category_id TEXT PRIMARY KEY,
    property_id TEXT NOT NULL,
    name        TEXT NOT NULL,
    ordinal     INTEGER NOT NULL,
    FOREIGN KEY (property_id) REFERENCES properties(property_id)
);

CREATE TABLE crumb_properties (
    crumb_id    TEXT NOT NULL,
    property_id TEXT NOT NULL,
    value_type  TEXT NOT NULL,
    value       TEXT NOT NULL,
    PRIMARY KEY (crumb_id, property_id),
    FOREIGN KEY (crumb_id) REFERENCES crumbs(crumb_id),
    FOREIGN KEY (property_id) REFERENCES properties(property_id)
);

CREATE TABLE metadata (
    metadata_id TEXT PRIMARY KEY,
    table_name  TEXT NOT NULL,
    crumb_id    TEXT NOT NULL,
    property_id TEXT,
    content     TEXT NOT NULL,
    created_at  TEXT NOT NULL,
    FOREIGN KEY (crumb_id) REFERENCES crumbs(crumb_id)
);

CREATE TABLE stashes (
    stash_id   TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    stash_type TEXT NOT NULL,
    value      TEXT NOT NULL,
    version    INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE stash_history (
    history_id TEXT PRIMARY KEY,
    stash_id   TEXT NOT NULL,
    version    INTEGER NOT NULL,
    value      TEXT NOT NULL,
    operation  TEXT NOT NULL,
    changed_by TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (stash_id) REFERENCES stashes(stash_id),
    FOREIGN KEY (changed_by) REFERENCES crumbs(crumb_id)
);
`

// Index DDL for common query patterns (R3.3).
const indexSQL = `
CREATE INDEX idx_crumbs_state ON crumbs(state);
CREATE INDEX idx_trails_state ON trails(state);
CREATE INDEX idx_links_type_from ON links(link_type, from_id);
CREATE INDEX idx_links_type_to ON links(link_type, to_id);
CREATE INDEX idx_crumb_properties_crumb ON crumb_properties(crumb_id);
CREATE INDEX idx_crumb_properties_property ON crumb_properties(property_id);
CREATE INDEX idx_metadata_crumb ON metadata(crumb_id);
CREATE INDEX idx_metadata_table ON metadata(table_name);
CREATE INDEX idx_categories_property ON categories(property_id);
CREATE INDEX idx_stashes_name ON stashes(name);
CREATE INDEX idx_stash_history_stash ON stash_history(stash_id);
CREATE INDEX idx_stash_history_version ON stash_history(stash_id, version);
`

// CreateSchema executes all CREATE TABLE and CREATE INDEX statements
// against the provided database connection.
func CreateSchema(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return err
	}
	if _, err := db.Exec(indexSQL); err != nil {
		return err
	}
	return nil
}
