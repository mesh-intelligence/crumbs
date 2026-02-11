package sqlite

import (
	"database/sql"
	"time"

	"github.com/mesh-intelligence/crumbs/pkg/types"
)

// builtinProperty describes a property to seed on first startup.
type builtinProperty struct {
	Name        string
	Description string
	ValueType   string
	Categories  []builtinCategory
}

// builtinCategory describes a category to seed for a categorical property.
type builtinCategory struct {
	Name    string
	Ordinal int
}

// builtinProperties defines the properties seeded when properties.jsonl is empty.
// Implements: prd002-sqlite-backend R9, prd004-properties-interface R9.
var builtinProperties = []builtinProperty{
	{
		Name:        "priority",
		Description: "Task priority level",
		ValueType:   types.ValueTypeCategorical,
		Categories: []builtinCategory{
			{"highest", 0},
			{"high", 1},
			{"medium", 2},
			{"low", 3},
			{"lowest", 4},
		},
	},
	{
		Name:        "type",
		Description: "Crumb type classification",
		ValueType:   types.ValueTypeCategorical,
		Categories: []builtinCategory{
			{"task", 0},
			{"epic", 1},
			{"bug", 2},
			{"chore", 3},
		},
	},
	{
		Name:        "description",
		Description: "Detailed description",
		ValueType:   types.ValueTypeText,
	},
	{
		Name:        "owner",
		Description: "Assigned worker or user ID",
		ValueType:   types.ValueTypeText,
	},
	{
		Name:        "labels",
		Description: "Capability tags or labels",
		ValueType:   types.ValueTypeList,
	},
}

// seedBuiltins creates built-in properties and categories if the properties
// table is empty (first startup). Does not modify existing data.
// Implements: prd002-sqlite-backend R9.4, prd004-properties-interface R9.4.
func (b *SQLiteBackend) seedBuiltins() error {
	var count int
	if err := b.db.QueryRow("SELECT COUNT(*) FROM properties").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now()
	propTable := &table{name: types.TableProperties, backend: b}

	for _, bp := range builtinProperties {
		propID := newUUID()
		var desc sql.NullString
		if bp.Description != "" {
			desc = sql.NullString{String: bp.Description, Valid: true}
		}

		_, err := b.db.Exec(`
			INSERT INTO properties (property_id, name, description, value_type, created_at)
			VALUES (?, ?, ?, ?, ?)`,
			propID, bp.Name, desc, bp.ValueType, now.Format(json_time_format))
		if err != nil {
			return err
		}

		for _, bc := range bp.Categories {
			catID := newUUID()
			_, err := b.db.Exec(`
				INSERT INTO categories (category_id, property_id, name, ordinal)
				VALUES (?, ?, ?, ?)`,
				catID, propID, bc.Name, bc.Ordinal)
			if err != nil {
				return err
			}
		}
	}

	// Persist seeded data to JSONL.
	if err := propTable.persistPropertiesJSONL(); err != nil {
		return err
	}
	return propTable.persistCategoriesJSONL()
}
