package types

import "time"

// Crumb represents a work item in the system.
// Implements: prd003-crumbs-interface (R1: struct, R2: states, R4: state methods, R5: property methods).
type Crumb struct {
	CrumbID    string         `json:"crumb_id"`
	Name       string         `json:"name"`
	State      string         `json:"state"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Properties map[string]any `json:"properties"`
}

// Crumb state constants per prd003-crumbs-interface R2.1.
const (
	StateDraft   = "draft"
	StatePending = "pending"
	StateReady   = "ready"
	StateTaken   = "taken"
	StatePebble  = "pebble"
	StateDust    = "dust"
)

// validCrumbStates is the set of recognized crumb states.
var validCrumbStates = map[string]bool{
	StateDraft:   true,
	StatePending: true,
	StateReady:   true,
	StateTaken:   true,
	StatePebble:  true,
	StateDust:    true,
}

// SetState transitions the crumb to the specified state.
// Returns ErrInvalidState if state is not recognized.
// Idempotent: setting to the current state succeeds without error.
// See prd003-crumbs-interface R4.2.
func (c *Crumb) SetState(state string) error {
	if !validCrumbStates[state] {
		return ErrInvalidState
	}
	c.State = state
	c.UpdatedAt = time.Now()
	return nil
}

// Pebble transitions the crumb to the pebble state (completed successfully).
// Returns ErrInvalidTransition if the current state is not taken.
// See prd003-crumbs-interface R4.3.
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
// See prd003-crumbs-interface R4.4.
func (c *Crumb) Dust() error {
	c.State = StateDust
	c.UpdatedAt = time.Now()
	return nil
}

// SetProperty assigns a value to a property in the Properties map.
// Validates that the property exists. Type validation and category
// validation are deferred to Table.Set per prd003 R5.7.
// See prd003-crumbs-interface R5.2.
func (c *Crumb) SetProperty(propertyID string, value any) error {
	if c.Properties == nil {
		return ErrPropertyNotFound
	}
	if _, exists := c.Properties[propertyID]; !exists {
		return ErrPropertyNotFound
	}
	c.Properties[propertyID] = value
	c.UpdatedAt = time.Now()
	return nil
}

// GetProperty retrieves a single property value from the Properties map.
// Returns ErrPropertyNotFound if the property does not exist.
// See prd003-crumbs-interface R5.3.
func (c *Crumb) GetProperty(propertyID string) (any, error) {
	if c.Properties == nil {
		return nil, ErrPropertyNotFound
	}
	val, exists := c.Properties[propertyID]
	if !exists {
		return nil, ErrPropertyNotFound
	}
	return val, nil
}

// GetProperties returns the full Properties map.
// See prd003-crumbs-interface R5.4.
func (c *Crumb) GetProperties() map[string]any {
	if c.Properties == nil {
		return map[string]any{}
	}
	return c.Properties
}

// ClearProperty resets a property to nil. The backend should set it to
// the type-based default on persist. The property entry remains in the map.
// Returns ErrPropertyNotFound if the property does not exist.
// See prd003-crumbs-interface R5.5.
func (c *Crumb) ClearProperty(propertyID string) error {
	if c.Properties == nil {
		return ErrPropertyNotFound
	}
	if _, exists := c.Properties[propertyID]; !exists {
		return ErrPropertyNotFound
	}
	c.Properties[propertyID] = nil
	c.UpdatedAt = time.Now()
	return nil
}
