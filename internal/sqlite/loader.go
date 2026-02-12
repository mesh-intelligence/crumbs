// This file implements JSONL loading and built-in property seeding for startup.
// Implements: prd002-sqlite-backend R4 (startup sequence), R4.2 (malformed lines),
//             R4.4 (transactional loading), R9 (built-in properties seeding).
package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/mesh-intelligence/crumbs/pkg/types"
)

// jsonlTableMapping maps JSONL filenames to their SQLite tables and column lists.
// The order matters: tables with foreign keys must load after their referenced tables.
var jsonlTableMapping = []struct {
	file    string
	table   string
	columns []string
}{
	{"crumbs.jsonl", "crumbs", []string{"crumb_id", "name", "state", "created_at", "updated_at"}},
	{"trails.jsonl", "trails", []string{"trail_id", "state", "created_at", "completed_at"}},
	{"properties.jsonl", "properties", []string{"property_id", "name", "description", "value_type", "created_at"}},
	{"categories.jsonl", "categories", []string{"category_id", "property_id", "name", "ordinal"}},
	{"crumb_properties.jsonl", "crumb_properties", []string{"crumb_id", "property_id", "value_type", "value"}},
	{"links.jsonl", "links", []string{"link_id", "link_type", "from_id", "to_id", "created_at"}},
	{"metadata.jsonl", "metadata", []string{"metadata_id", "table_name", "crumb_id", "property_id", "content", "created_at"}},
	{"stashes.jsonl", "stashes", []string{"stash_id", "name", "stash_type", "value", "version", "created_at", "updated_at"}},
	{"stash_history.jsonl", "stash_history", []string{"history_id", "stash_id", "version", "value", "operation", "changed_by", "created_at"}},
}

// loadAllJSONL reads each JSONL file from DataDir and inserts records into the
// corresponding SQLite tables. Loading is transactional: all succeed or the
// database remains empty (prd002-sqlite-backend R4.4). Malformed lines are
// skipped per R4.2.
func loadAllJSONL(db *sql.DB, dataDir string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("beginning load transaction: %w", err)
	}
	defer tx.Rollback()

	// Disable foreign keys during loading, re-enable after.
	if _, err := tx.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disabling foreign keys for load: %w", err)
	}

	for _, mapping := range jsonlTableMapping {
		path := filepath.Join(dataDir, mapping.file)
		records, err := readJSONL(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", mapping.file, err)
		}

		if len(records) == 0 {
			continue
		}

		if err := insertRecords(tx, mapping.table, mapping.columns, records); err != nil {
			return fmt.Errorf("loading %s into %s: %w", mapping.file, mapping.table, err)
		}
	}

	if _, err := tx.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("re-enabling foreign keys: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing load transaction: %w", err)
	}

	return nil
}

// insertRecords inserts parsed JSONL records into a SQLite table.
func insertRecords(tx *sql.Tx, table string, columns []string, records []json.RawMessage) error {
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		joinColumns(columns),
		joinColumns(placeholders),
	)

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("preparing insert for %s: %w", table, err)
	}
	defer stmt.Close()

	for _, rec := range records {
		var obj map[string]any
		if err := json.Unmarshal(rec, &obj); err != nil {
			// Skip malformed records (prd002-sqlite-backend R4.2).
			continue
		}

		args := make([]any, len(columns))
		for i, col := range columns {
			val, ok := obj[col]
			if !ok {
				args[i] = nil
				continue
			}
			// JSON values (stash value, etc.) need to be re-serialized as strings.
			switch v := val.(type) {
			case map[string]any, []any:
				b, err := json.Marshal(v)
				if err != nil {
					args[i] = nil
					continue
				}
				args[i] = string(b)
			default:
				args[i] = val
			}
		}

		if _, err := stmt.Exec(args...); err != nil {
			// Skip records that violate constraints (prd002-sqlite-backend R4.2).
			continue
		}
	}

	return nil
}

// joinColumns joins column names with commas.
func joinColumns(cols []string) string {
	result := ""
	for i, c := range cols {
		if i > 0 {
			result += ", "
		}
		result += c
	}
	return result
}

// builtInProperty describes a property to seed on first startup.
type builtInProperty struct {
	name        string
	valueType   string
	description string
	categories  []builtInCategory
}

// builtInCategory describes a category to seed with a built-in property.
type builtInCategory struct {
	name    string
	ordinal int
}

// builtInProperties defines the properties seeded on first startup
// (prd002-sqlite-backend R9.1).
var builtInProperties = []builtInProperty{
	{
		name:        types.PropertyPriority,
		valueType:   types.ValueTypeCategorical,
		description: "Task priority (0=highest, 4=lowest)",
		categories: []builtInCategory{
			{"highest", 0},
			{"high", 1},
			{"medium", 2},
			{"low", 3},
			{"lowest", 4},
		},
	},
	{
		name:        types.PropertyType,
		valueType:   types.ValueTypeCategorical,
		description: "Crumb type (task, epic, bug, etc.)",
		categories: []builtInCategory{
			{"task", 0},
			{"epic", 1},
			{"bug", 2},
			{"chore", 3},
		},
	},
	{
		name:        types.PropertyDescription,
		valueType:   types.ValueTypeText,
		description: "Detailed description",
	},
	{
		name:        types.PropertyOwner,
		valueType:   types.ValueTypeText,
		description: "Assigned worker/user ID",
	},
	{
		name:        types.PropertyLabels,
		valueType:   types.ValueTypeList,
		description: "Capability tags",
	},
}

// seedBuiltInProperties creates the built-in properties and categories if the
// properties table is empty (first run). This only runs when properties.jsonl
// was empty on startup (prd002-sqlite-backend R9.4).
func seedBuiltInProperties(db *sql.DB, dataDir string) error {
	// Check if properties already exist.
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM properties").Scan(&count); err != nil {
		return fmt.Errorf("counting properties: %w", err)
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("beginning seed transaction: %w", err)
	}
	defer tx.Rollback()

	for _, bp := range builtInProperties {
		propID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("generating property UUID: %w", err)
		}

		_, err = tx.Exec(
			"INSERT INTO properties (property_id, name, description, value_type, created_at) VALUES (?, ?, ?, ?, ?)",
			propID.String(), bp.name, bp.description, bp.valueType, nowStr,
		)
		if err != nil {
			return fmt.Errorf("seeding property %s: %w", bp.name, err)
		}

		for _, cat := range bp.categories {
			catID, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("generating category UUID: %w", err)
			}
			_, err = tx.Exec(
				"INSERT INTO categories (category_id, property_id, name, ordinal) VALUES (?, ?, ?, ?)",
				catID.String(), propID.String(), cat.name, cat.ordinal,
			)
			if err != nil {
				return fmt.Errorf("seeding category %s for %s: %w", cat.name, bp.name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing seed transaction: %w", err)
	}

	// Persist seeded data to JSONL files so they are not empty on next startup.
	if err := persistSeededJSONL(db, dataDir); err != nil {
		return fmt.Errorf("persisting seeded data: %w", err)
	}

	return nil
}

// persistSeededJSONL writes the seeded properties and categories to their
// JSONL files after first-run seeding.
func persistSeededJSONL(db *sql.DB, dataDir string) error {
	// Persist properties.jsonl.
	rows, err := db.Query(
		"SELECT property_id, name, description, value_type, created_at FROM properties ORDER BY created_at ASC",
	)
	if err != nil {
		return fmt.Errorf("querying properties for seed JSONL: %w", err)
	}
	defer rows.Close()

	var propRecords []json.RawMessage
	for rows.Next() {
		var id, name, valueType, createdAt string
		var desc sql.NullString
		if err := rows.Scan(&id, &name, &desc, &valueType, &createdAt); err != nil {
			return fmt.Errorf("scanning property for seed JSONL: %w", err)
		}
		rec := map[string]any{
			"property_id": id,
			"name":        name,
			"description": desc.String,
			"value_type":  valueType,
			"created_at":  createdAt,
		}
		data, err := json.Marshal(rec)
		if err != nil {
			return fmt.Errorf("marshaling property for seed JSONL: %w", err)
		}
		propRecords = append(propRecords, data)
	}
	rows.Close()
	if err := writeJSONL(filepath.Join(dataDir, "properties.jsonl"), propRecords); err != nil {
		return fmt.Errorf("writing properties.jsonl: %w", err)
	}

	// Persist categories.jsonl.
	catRows, err := db.Query(
		"SELECT category_id, property_id, name, ordinal FROM categories ORDER BY property_id, ordinal",
	)
	if err != nil {
		return fmt.Errorf("querying categories for seed JSONL: %w", err)
	}
	defer catRows.Close()

	var catRecords []json.RawMessage
	for catRows.Next() {
		var id, propID, name string
		var ordinal int
		if err := catRows.Scan(&id, &propID, &name, &ordinal); err != nil {
			return fmt.Errorf("scanning category for seed JSONL: %w", err)
		}
		rec := map[string]any{
			"category_id": id,
			"property_id": propID,
			"name":        name,
			"ordinal":     ordinal,
		}
		data, err := json.Marshal(rec)
		if err != nil {
			return fmt.Errorf("marshaling category for seed JSONL: %w", err)
		}
		catRecords = append(catRecords, data)
	}
	catRows.Close()
	return writeJSONL(filepath.Join(dataDir, "categories.jsonl"), catRecords)
}
