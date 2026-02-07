// Category operations for the SQLite backend.
// Implements: prd-properties-interface R7, R8 (DefineCategory, GetCategories);
//
//	docs/ARCHITECTURE § Table Interfaces.
package sqlite

import (
	"github.com/mesh-intelligence/crumbs/pkg/types"
)

// Ensure Backend implements CategoryDefiner.
var _ types.CategoryDefiner = (*Backend)(nil)

// DefineCategory creates a new category for a property and persists it.
// Per prd-properties-interface R7.
//
// Generates a UUID v7 for CategoryID.
// Validates that name is unique within the property (ErrDuplicateName if exists).
// Persists to SQLite and categories.jsonl.
// Returns the created Category with all fields populated.
func (b *Backend) DefineCategory(propertyID, name string, ordinal int) (*types.Category, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.attached {
		return nil, types.ErrCupboardDetached
	}

	// Check for duplicate name within the property
	var count int
	err := b.db.QueryRow(
		"SELECT COUNT(*) FROM categories WHERE property_id = ? AND name = ?",
		propertyID, name,
	).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, types.ErrDuplicateName
	}

	// Generate UUID v7 for CategoryID
	categoryID := generateUUID()

	// Insert into SQLite
	_, err = b.db.Exec(
		`INSERT INTO categories (category_id, property_id, name, ordinal)
		 VALUES (?, ?, ?, ?)`,
		categoryID, propertyID, name, ordinal,
	)
	if err != nil {
		return nil, err
	}

	// Persist to JSONL
	cat := &categoryJSON{
		CategoryID: categoryID,
		PropertyID: propertyID,
		Name:       name,
		Ordinal:    ordinal,
	}
	if err := b.saveCategoryToJSONL(cat); err != nil {
		return nil, err
	}

	return &types.Category{
		CategoryID: categoryID,
		PropertyID: propertyID,
		Name:       name,
		Ordinal:    ordinal,
	}, nil
}

// GetCategories retrieves all categories for a property ordered by ordinal then name.
// Per prd-properties-interface R8.
//
// Returns an empty slice (not nil) if no categories exist.
func (b *Backend) GetCategories(propertyID string) ([]*types.Category, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.attached {
		return nil, types.ErrCupboardDetached
	}

	rows, err := b.db.Query(
		`SELECT category_id, property_id, name, ordinal
		 FROM categories
		 WHERE property_id = ?
		 ORDER BY ordinal ASC, name ASC`,
		propertyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*types.Category
	for rows.Next() {
		var cat types.Category
		if err := rows.Scan(&cat.CategoryID, &cat.PropertyID, &cat.Name, &cat.Ordinal); err != nil {
			return nil, err
		}
		categories = append(categories, &cat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return empty slice (not nil) per R8.5
	if categories == nil {
		categories = []*types.Category{}
	}

	return categories, nil
}
