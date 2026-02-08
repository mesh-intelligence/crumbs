// Crumb entity represents a work item in the task coordination system.
// Implements: prd-crumbs-interface R1, R4, R5 (Crumb struct, state methods, property methods);
//
//	docs/ARCHITECTURE § Main Interface.
package types

import (
	"slices"
	"time"
)

// Crumb state values.
// Terminal states: pebble (success) and dust (failed/abandoned).
const (
	StateDraft   = "draft"
	StatePending = "pending"
	StateReady   = "ready"
	StateTaken   = "taken"
	StatePebble  = "pebble" // completed successfully (permanent, enduring)
	StateDust    = "dust"   // failed or abandoned (swept away)
)

// Crumb represents a work item.
type Crumb struct {
	// CrumbID is a UUID v7, generated on creation.
	CrumbID string

	// Name is a human-readable name (required, non-empty).
	Name string

	// State is the crumb state (draft, pending, ready, taken, pebble, dust).
	State string

	// CreatedAt is the timestamp of creation.
	CreatedAt time.Time

	// UpdatedAt is the timestamp of last modification.
	UpdatedAt time.Time

	// Properties holds property values (property_id to value).
	Properties map[string]any
}

// validCrumbStates lists all valid crumb state values.
var validCrumbStates = []string{
	StateDraft, StatePending, StateReady, StateTaken,
	StatePebble, StateDust,
}

// SetState transitions the crumb to the specified state.
// Returns ErrInvalidState if the state is not recognized.
// Updates UpdatedAt. Caller must save via Table.Set.
func (c *Crumb) SetState(state string) error {
	if !slices.Contains(validCrumbStates, state) {
		return ErrInvalidState
	}
	c.State = state
	c.UpdatedAt = time.Now()
	return nil
}

// Pebble transitions the crumb to the pebble state (completed successfully).
// Returns ErrInvalidTransition if current state is not taken.
// Updates UpdatedAt. Caller must save via Table.Set.
func (c *Crumb) Pebble() error {
	if c.State != StateTaken {
		return ErrInvalidTransition
	}
	c.State = StatePebble
	c.UpdatedAt = time.Now()
	return nil
}

// Dust transitions the crumb to the dust state (failed or abandoned).
// Can be called from any state. Idempotent.
// Updates UpdatedAt. Caller must save via Table.Set.
func (c *Crumb) Dust() error {
	c.State = StateDust
	c.UpdatedAt = time.Now()
	return nil
}

// SetProperty assigns a value to a property.
// Initializes Properties map if nil.
// Updates UpdatedAt. Caller must save via Table.Set.
// Note: Type validation is deferred to Table.Set per PRD R5.7.
func (c *Crumb) SetProperty(propertyID string, value any) error {
	if c.Properties == nil {
		c.Properties = make(map[string]any)
	}
	c.Properties[propertyID] = value
	c.UpdatedAt = time.Now()
	return nil
}

// GetProperty retrieves a single property value.
// Returns ErrPropertyNotFound if the property does not exist.
func (c *Crumb) GetProperty(propertyID string) (any, error) {
	if c.Properties == nil {
		return nil, ErrPropertyNotFound
	}
	value, ok := c.Properties[propertyID]
	if !ok {
		return nil, ErrPropertyNotFound
	}
	return value, nil
}

// GetProperties retrieves all property values.
// Returns an empty map if no properties are set.
func (c *Crumb) GetProperties() map[string]any {
	if c.Properties == nil {
		return make(map[string]any)
	}
	return c.Properties
}

// ClearProperty resets a property to nil.
// Per prd-crumbs-interface R5.5, the map entry is preserved (properties are
// never unset). The type-based default is resolved by Table.Set during
// persistence (per R5.7, validation is deferred to persist).
// Returns ErrPropertyNotFound if the property does not exist.
// Updates UpdatedAt. Caller must save via Table.Set.
func (c *Crumb) ClearProperty(propertyID string) error {
	if c.Properties == nil {
		return ErrPropertyNotFound
	}
	if _, ok := c.Properties[propertyID]; !ok {
		return ErrPropertyNotFound
	}
	c.Properties[propertyID] = nil
	c.UpdatedAt = time.Now()
	return nil
}
