package types

import "time"

// Crumb states (prd003-crumbs-interface R2.1).
const (
	CrumbDraft   = "draft"
	CrumbPending = "pending"
	CrumbReady   = "ready"
	CrumbTaken   = "taken"
	CrumbPebble  = "pebble"
	CrumbDust    = "dust"
)

// validCrumbStates is the set of recognized crumb state values.
var validCrumbStates = map[string]bool{
	CrumbDraft:   true,
	CrumbPending: true,
	CrumbReady:   true,
	CrumbTaken:   true,
	CrumbPebble:  true,
	CrumbDust:    true,
}

// Crumb represents a work item (prd003-crumbs-interface R1).
type Crumb struct {
	CrumbID    string         `json:"crumb_id"`
	Name       string         `json:"name"`
	State      string         `json:"state"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Properties map[string]any `json:"properties"`
}

// SetState transitions the crumb to the specified state.
// Returns ErrInvalidState if the state is not recognized.
// Idempotent: setting the current state succeeds without error.
// (prd003-crumbs-interface R4.2)
func (c *Crumb) SetState(state string) error {
	if !validCrumbStates[state] {
		return ErrInvalidState
	}
	c.State = state
	c.UpdatedAt = time.Now()
	return nil
}

// Pebble transitions the crumb to the pebble state (completed successfully).
// Returns ErrInvalidTransition if the current state is not "taken".
// (prd003-crumbs-interface R4.3)
func (c *Crumb) Pebble() error {
	if c.State != CrumbTaken {
		return ErrInvalidTransition
	}
	c.State = CrumbPebble
	c.UpdatedAt = time.Now()
	return nil
}

// Dust transitions the crumb to the dust state (failed or abandoned).
// Can be called from any state. Idempotent.
// (prd003-crumbs-interface R4.4)
func (c *Crumb) Dust() error {
	c.State = CrumbDust
	c.UpdatedAt = time.Now()
	return nil
}

// SetProperty assigns a value to a property in the Properties map.
// Type validation is deferred to Table.Set (prd003-crumbs-interface R5.7).
func (c *Crumb) SetProperty(propertyID string, value any) error {
	if c.Properties == nil {
		c.Properties = make(map[string]any)
	}
	c.Properties[propertyID] = value
	c.UpdatedAt = time.Now()
	return nil
}

// GetProperty retrieves a single property value from the Properties map.
// Returns ErrPropertyNotFound if the property is not in the map.
// (prd003-crumbs-interface R5.3)
func (c *Crumb) GetProperty(propertyID string) (any, error) {
	if c.Properties == nil {
		return nil, ErrPropertyNotFound
	}
	v, ok := c.Properties[propertyID]
	if !ok {
		return nil, ErrPropertyNotFound
	}
	return v, nil
}

// GetProperties returns all property values.
// (prd003-crumbs-interface R5.4)
func (c *Crumb) GetProperties() map[string]any {
	if c.Properties == nil {
		return map[string]any{}
	}
	return c.Properties
}

// ClearProperty resets a property to nil. The backend resolves the
// type-based default when persisted via Table.Set.
// Returns ErrPropertyNotFound if the property is not in the map.
// (prd003-crumbs-interface R5.5)
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
