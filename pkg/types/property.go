// Property and Category entities for extensible attributes on crumbs.
// Implements: prd004-properties-interface R1, R2, R3, R7, R8 (Property, Category, value types, entity methods);
//
//	docs/ARCHITECTURE § Main Interface.
package types

import (
	"sort"
	"time"
)

// Value type constants.
const (
	ValueTypeCategorical = "categorical"
	ValueTypeText        = "text"
	ValueTypeInteger     = "integer"
	ValueTypeBoolean     = "boolean"
	ValueTypeTimestamp   = "timestamp"
	ValueTypeList        = "list"
)

// Property defines a custom attribute that can be assigned to crumbs.
type Property struct {
	// PropertyID is a UUID v7, generated on creation.
	PropertyID string

	// Name is a unique human-readable name (e.g., "priority", "labels").
	Name string

	// Description is an optional explanation of the property's purpose.
	Description string

	// ValueType is the type of values this property accepts.
	// One of: categorical, text, integer, boolean, timestamp, list.
	ValueType string

	// CreatedAt is the timestamp of creation.
	CreatedAt time.Time
}

// Category defines an enumeration value for categorical properties.
type Category struct {
	// CategoryID is a UUID v7, generated on creation.
	CategoryID string

	// PropertyID is the categorical property this category belongs to.
	PropertyID string

	// Name is the display name for this category (e.g., "high", "medium").
	Name string

	// Ordinal determines display order; lower ordinals sort first.
	Ordinal int
}

// CategoryDefiner provides category storage operations for the DefineCategory entity method.
// The SQLite backend implements this interface to allow Property entity methods to access storage.
type CategoryDefiner interface {
	// DefineCategory creates a new category for a property and persists it.
	// Returns the created Category with CategoryID populated.
	// Returns ErrDuplicateName if a category with the same name exists for this property.
	DefineCategory(propertyID, name string, ordinal int) (*Category, error)

	// GetCategories retrieves all categories for a property ordered by ordinal then name.
	// Returns an empty slice (not nil) if no categories exist.
	GetCategories(propertyID string) ([]*Category, error)
}

// DefineCategory creates a new category for this categorical property.
// Per prd004-properties-interface R7.
//
// Validates that the property's ValueType is "categorical" (ErrInvalidValueType if not).
// Validates that name is non-empty (ErrInvalidName if empty).
// Creates a Category with UUID v7, PropertyID, name, and ordinal.
// The definer parameter provides backend access for persistence and uniqueness validation.
func (p *Property) DefineCategory(definer CategoryDefiner, name string, ordinal int) (*Category, error) {
	if p.ValueType != ValueTypeCategorical {
		return nil, ErrInvalidValueType
	}
	if name == "" {
		return nil, ErrInvalidName
	}
	return definer.DefineCategory(p.PropertyID, name, ordinal)
}

// GetCategories retrieves all categories for this categorical property.
// Per prd004-properties-interface R8.
//
// Validates that the property's ValueType is "categorical" (ErrInvalidValueType if not).
// Returns categories ordered by ordinal ascending, then name ascending for ties.
// Returns an empty slice (not nil) if no categories are defined.
func (p *Property) GetCategories(definer CategoryDefiner) ([]*Category, error) {
	if p.ValueType != ValueTypeCategorical {
		return nil, ErrInvalidValueType
	}
	categories, err := definer.GetCategories(p.PropertyID)
	if err != nil {
		return nil, err
	}
	// Ensure ordering by ordinal ascending, then name ascending
	sort.Slice(categories, func(i, j int) bool {
		if categories[i].Ordinal != categories[j].Ordinal {
			return categories[i].Ordinal < categories[j].Ordinal
		}
		return categories[i].Name < categories[j].Name
	})
	return categories, nil
}
