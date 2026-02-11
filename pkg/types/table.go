// Implements: prd001-cupboard-core (R3: Table Interface, R7.2: table operation errors, R7.3: entity method errors);
//             docs/ARCHITECTURE § Main Interface.
package types

import "errors"

// Table provides uniform CRUD operations for all entity types.
// Get and Fetch return any; callers use type assertions to access
// entity-specific fields (prd001-cupboard-core R3.6).
type Table interface {
	// Get retrieves an entity by ID. Returns ErrNotFound if absent.
	Get(id string) (any, error)

	// Set persists an entity. If id is empty, generates a UUID v7 and creates
	// the entity. If id is provided, updates the existing entity or creates it
	// if not found. Returns the actual ID and any error.
	Set(id string, data any) (string, error)

	// Delete removes an entity by ID. Returns ErrNotFound if absent.
	Delete(id string) error

	// Fetch queries entities matching the filter. An empty filter returns all
	// entities in the table.
	Fetch(filter map[string]any) ([]any, error)
}

// Table operation errors (prd001-cupboard-core R7.2).
var (
	ErrNotFound    = errors.New("entity not found")
	ErrInvalidID   = errors.New("invalid entity ID")
	ErrInvalidData = errors.New("invalid entity data")
)

// Entity method errors (prd001-cupboard-core R7.3).
var (
	ErrInvalidState      = errors.New("invalid state value")
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrInvalidName       = errors.New("invalid name")
	ErrPropertyNotFound  = errors.New("property not found")
	ErrTypeMismatch      = errors.New("type mismatch")
	ErrInvalidCategory   = errors.New("invalid category")
	ErrInvalidStashType  = errors.New("invalid stash type or operation")
	ErrLockHeld          = errors.New("lock is held")
	ErrNotLockHolder     = errors.New("caller is not the lock holder")
	ErrInvalidHolder     = errors.New("holder cannot be empty")
	ErrAlreadyInTrail    = errors.New("crumb already belongs to a trail")
	ErrNotInTrail        = errors.New("crumb does not belong to the trail")
	ErrSchemaNotFound    = errors.New("schema not found")
	ErrInvalidContent    = errors.New("content must not be empty")
	ErrInvalidFilter     = errors.New("invalid filter value type")
)
