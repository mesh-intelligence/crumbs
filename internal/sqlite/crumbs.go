// Implements: prd002-sqlite-backend (R13: Table Interface, R14.2: Crumb hydration,
//             R15: Entity Persistence);
//             prd003-crumbs-interface (R1: Crumb struct, R3: Creating Crumbs,
//             R6: Retrieving, R7: Updating, R8: Deleting, R9: Filter Map);
//             prd001-cupboard-core (R3: Table Interface, R8: UUID v7).
package sqlite

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/mesh-intelligence/crumbs/pkg/types"
)

// crumbsTable implements types.Table for crumbs.
type crumbsTable struct {
	backend *Backend
}

// Compile-time assertion: crumbsTable implements types.Table.
var _ types.Table = (*crumbsTable)(nil)

// Get retrieves a crumb by ID. Returns ErrNotFound if absent,
// ErrInvalidID if id is empty (prd003-crumbs-interface R6.3, R6.4).
func (t *crumbsTable) Get(id string) (any, error) {
	t.backend.mu.RLock()
	defer t.backend.mu.RUnlock()

	if !t.backend.attached {
		return nil, types.ErrCupboardDetached
	}
	if id == "" {
		return nil, types.ErrInvalidID
	}

	row := t.backend.db.QueryRow(
		`SELECT crumb_id, name, state, created_at, updated_at FROM crumbs WHERE crumb_id = ?`,
		id,
	)
	c, err := hydrateCrumb(row)
	if err == sql.ErrNoRows {
		return nil, types.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting crumb %s: %w", id, err)
	}
	return c, nil
}

// Set persists a crumb. If id is empty, generates a UUID v7 and creates the
// crumb with state "draft". If id is provided, updates the existing crumb or
// creates it if not found. Returns the actual ID.
func (t *crumbsTable) Set(id string, data any) (string, error) {
	t.backend.mu.Lock()
	defer t.backend.mu.Unlock()

	if !t.backend.attached {
		return "", types.ErrCupboardDetached
	}

	crumb, ok := data.(*types.Crumb)
	if !ok {
		return "", types.ErrInvalidData
	}
	if crumb.Name == "" {
		return "", types.ErrInvalidName
	}

	now := time.Now().UTC()

	if id == "" {
		// Create: generate UUID v7, set defaults (prd003-crumbs-interface R3.2).
		newID, err := uuid.NewV7()
		if err != nil {
			return "", fmt.Errorf("generating UUID v7: %w", err)
		}
		crumb.CrumbID = newID.String()
		crumb.State = types.StateDraft
		crumb.CreatedAt = now
		crumb.UpdatedAt = now
		id = crumb.CrumbID
	} else {
		crumb.CrumbID = id
		crumb.UpdatedAt = now
	}

	// INSERT or UPDATE (prd002-sqlite-backend R15.6).
	var exists bool
	err := t.backend.db.QueryRow(`SELECT 1 FROM crumbs WHERE crumb_id = ?`, id).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("checking crumb existence: %w", err)
	}

	if err == sql.ErrNoRows {
		_, err = t.backend.db.Exec(
			`INSERT INTO crumbs (crumb_id, name, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
			crumb.CrumbID,
			crumb.Name,
			crumb.State,
			crumb.CreatedAt.Format(timeFormat),
			crumb.UpdatedAt.Format(timeFormat),
		)
	} else {
		_, err = t.backend.db.Exec(
			`UPDATE crumbs SET name = ?, state = ?, created_at = ?, updated_at = ? WHERE crumb_id = ?`,
			crumb.Name,
			crumb.State,
			crumb.CreatedAt.Format(timeFormat),
			crumb.UpdatedAt.Format(timeFormat),
			crumb.CrumbID,
		)
	}
	if err != nil {
		return "", fmt.Errorf("persisting crumb: %w", err)
	}

	// Persist to crumbs.jsonl atomically (prd002-sqlite-backend R5.1, R5.2).
	if err := t.persistCrumbsJSONL(); err != nil {
		return "", fmt.Errorf("persisting crumbs.jsonl: %w", err)
	}

	return crumb.CrumbID, nil
}

// Delete removes a crumb by ID. Returns ErrNotFound if absent,
// ErrInvalidID if id is empty (prd003-crumbs-interface R8.4, R8.5).
func (t *crumbsTable) Delete(id string) error {
	t.backend.mu.Lock()
	defer t.backend.mu.Unlock()

	if !t.backend.attached {
		return types.ErrCupboardDetached
	}
	if id == "" {
		return types.ErrInvalidID
	}

	result, err := t.backend.db.Exec(`DELETE FROM crumbs WHERE crumb_id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting crumb %s: %w", id, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return types.ErrNotFound
	}

	if err := t.persistCrumbsJSONL(); err != nil {
		return fmt.Errorf("persisting crumbs.jsonl: %w", err)
	}
	return nil
}

// Fetch queries crumbs matching the filter. An empty filter returns all crumbs.
// Supported filter keys: "states" ([]string). Results ordered by created_at DESC.
func (t *crumbsTable) Fetch(filter map[string]any) ([]any, error) {
	t.backend.mu.RLock()
	defer t.backend.mu.RUnlock()

	if !t.backend.attached {
		return nil, types.ErrCupboardDetached
	}

	query := `SELECT crumb_id, name, state, created_at, updated_at FROM crumbs`
	var args []any
	var where string

	if states, ok := filter["states"]; ok {
		sl, ok := states.([]string)
		if !ok {
			return nil, types.ErrInvalidFilter
		}
		if len(sl) > 0 {
			placeholders := ""
			for i, s := range sl {
				if i > 0 {
					placeholders += ", "
				}
				placeholders += "?"
				args = append(args, s)
			}
			where = " WHERE state IN (" + placeholders + ")"
		}
	}

	query += where + " ORDER BY created_at DESC"
	rows, err := t.backend.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching crumbs: %w", err)
	}
	defer rows.Close()

	var result []any
	for rows.Next() {
		c, err := hydrateCrumbFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("hydrating crumb: %w", err)
		}
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating crumbs: %w", err)
	}

	// Return empty slice, not nil (prd003-crumbs-interface R10.3).
	if result == nil {
		result = []any{}
	}
	return result, nil
}

// hydrateCrumb converts a single SQL row into a *types.Crumb
// (prd002-sqlite-backend R14.2).
func hydrateCrumb(row *sql.Row) (*types.Crumb, error) {
	var c types.Crumb
	var createdAt, updatedAt string
	err := row.Scan(&c.CrumbID, &c.Name, &c.State, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	c.CreatedAt, err = time.Parse(timeFormat, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parsing created_at: %w", err)
	}
	c.UpdatedAt, err = time.Parse(timeFormat, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing updated_at: %w", err)
	}
	return &c, nil
}

// hydrateCrumbFromRows converts a row from sql.Rows into a *types.Crumb.
func hydrateCrumbFromRows(rows *sql.Rows) (*types.Crumb, error) {
	var c types.Crumb
	var createdAt, updatedAt string
	err := rows.Scan(&c.CrumbID, &c.Name, &c.State, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	c.CreatedAt, err = time.Parse(timeFormat, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parsing created_at: %w", err)
	}
	c.UpdatedAt, err = time.Parse(timeFormat, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing updated_at: %w", err)
	}
	return &c, nil
}

// persistCrumbsJSONL reads all crumbs from SQLite and writes them to
// crumbs.jsonl atomically. Must be called with b.mu held for writing.
func (t *crumbsTable) persistCrumbsJSONL() error {
	rows, err := t.backend.db.Query(
		`SELECT crumb_id, name, state, created_at, updated_at FROM crumbs ORDER BY created_at`,
	)
	if err != nil {
		return fmt.Errorf("querying crumbs for JSONL: %w", err)
	}
	defer rows.Close()

	var crumbs []types.Crumb
	for rows.Next() {
		c, err := hydrateCrumbFromRows(rows)
		if err != nil {
			return fmt.Errorf("hydrating crumb for JSONL: %w", err)
		}
		crumbs = append(crumbs, *c)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating crumbs for JSONL: %w", err)
	}

	path := filepath.Join(t.backend.config.DataDir, "crumbs.jsonl")
	return persistJSONL(path, crumbs)
}
