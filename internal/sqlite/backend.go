// Implements: prd002-sqlite-backend (R1: Directory Layout, R4: Startup Sequence,
//             R6: Shutdown Sequence, R8: Concurrency Model, R11: Cupboard Interface,
//             R12: Table Name Routing);
//             prd001-cupboard-core (R2: Cupboard Interface, R4: Attach, R5: Detach,
//             R6: Error Handling After Detach);
//             docs/ARCHITECTURE § SQLite Backend.
package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mesh-intelligence/crumbs/pkg/types"

	_ "modernc.org/sqlite"
)

// jsonlFiles lists all JSONL files that the backend creates on Attach
// (prd002-sqlite-backend R1.2). We create all nine files so the directory
// layout matches the spec from day one, even though only crumbs.jsonl is
// active in this task.
var jsonlFiles = []string{
	"crumbs.jsonl",
	"trails.jsonl",
	"links.jsonl",
	"properties.jsonl",
	"categories.jsonl",
	"crumb_properties.jsonl",
	"metadata.jsonl",
	"stashes.jsonl",
	"stash_history.jsonl",
}

// Backend implements types.Cupboard using SQLite as a query engine and
// JSONL files as the source of truth.
type Backend struct {
	mu       sync.RWMutex
	attached bool
	config   types.Config
	db       *sql.DB
	tables   map[string]types.Table
}

// Compile-time assertion: Backend implements types.Cupboard.
var _ types.Cupboard = (*Backend)(nil)

// NewBackend creates a new unattached Backend.
func NewBackend() *Backend {
	return &Backend{}
}

// Attach initializes the backend: creates the data directory, creates JSONL
// files, creates the SQLite schema, and loads crumbs.jsonl into SQLite.
// Returns ErrAlreadyAttached if called on an attached backend.
func (b *Backend) Attach(config types.Config) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.attached {
		return types.ErrAlreadyAttached
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("validating config: %w", err)
	}

	// R1.3: create DataDir if missing.
	if err := os.MkdirAll(config.DataDir, 0o755); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	// R1.4: create empty JSONL files if missing.
	for _, name := range jsonlFiles {
		p := filepath.Join(config.DataDir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			if err := os.WriteFile(p, nil, 0o644); err != nil {
				return fmt.Errorf("creating %s: %w", name, err)
			}
		}
	}

	// R4.1: delete cupboard.db if it exists (ephemeral cache).
	dbPath := filepath.Join(config.DataDir, "cupboard.db")
	_ = os.Remove(dbPath)

	// Open SQLite (modernc.org/sqlite, pure Go).
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("opening SQLite: %w", err)
	}

	// Create schema (crumbs table only for this task).
	if err := createSchema(db); err != nil {
		db.Close()
		return fmt.Errorf("creating schema: %w", err)
	}

	b.db = db
	b.config = config

	// Load crumbs.jsonl into SQLite.
	crumbsPath := filepath.Join(config.DataDir, "crumbs.jsonl")
	if err := b.loadCrumbs(crumbsPath); err != nil {
		db.Close()
		return fmt.Errorf("loading crumbs: %w", err)
	}

	// Create table accessors (R12.4: created once, reused).
	b.tables = map[string]types.Table{
		types.TableCrumbs: &crumbsTable{backend: b},
	}

	b.attached = true
	return nil
}

// loadCrumbs reads crumbs.jsonl and inserts each crumb into SQLite.
func (b *Backend) loadCrumbs(path string) error {
	crumbs, err := loadJSONL[types.Crumb](path)
	if err != nil {
		return err
	}
	tx, err := b.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	stmt, err := tx.Prepare(`INSERT INTO crumbs (crumb_id, name, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("preparing insert: %w", err)
	}
	defer stmt.Close()

	for _, c := range crumbs {
		_, err := stmt.Exec(
			c.CrumbID,
			c.Name,
			c.State,
			c.CreatedAt.Format(timeFormat),
			c.UpdatedAt.Format(timeFormat),
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("inserting crumb %s: %w", c.CrumbID, err)
		}
	}
	return tx.Commit()
}

// Detach closes the SQLite connection and marks the backend as detached.
// Subsequent operations return ErrCupboardDetached. Detach is idempotent.
func (b *Backend) Detach() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.attached {
		return nil
	}

	if b.db != nil {
		b.db.Close()
		b.db = nil
	}

	b.tables = nil
	b.attached = false
	return nil
}

// GetTable returns a Table for the given name. Returns ErrTableNotFound for
// unrecognized names and ErrCupboardDetached if the backend is detached.
func (b *Backend) GetTable(name string) (types.Table, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.attached {
		return nil, types.ErrCupboardDetached
	}

	t, ok := b.tables[name]
	if !ok {
		return nil, types.ErrTableNotFound
	}
	return t, nil
}

// timeFormat is RFC 3339 for timestamp serialization (prd002-sqlite-backend R2.11).
const timeFormat = "2006-01-02T15:04:05Z07:00"
