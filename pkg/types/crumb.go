package types

import "time"

// Crumb state constants (prd003-crumbs-interface R2.1).
const (
	StateDraft   = "draft"
	StatePending = "pending"
	StateReady   = "ready"
	StateTaken   = "taken"
	StatePebble  = "pebble"
	StateDust    = "dust"
)

// validStates provides O(1) lookup for state validation.
var validStates = map[string]bool{
	StateDraft:   true,
	StatePending: true,
	StateReady:   true,
	StateTaken:   true,
	StatePebble:  true,
	StateDust:    true,
}

// Crumb represents a work item in the Crumbs storage system
// (prd003-crumbs-interface R1.1). Entity methods modify the struct in memory;
// the caller must call Table.Set to persist changes.
type Crumb struct {
	CrumbID    string         `json:"crumb_id"`
	Name       string         `json:"name"`
	State      string         `json:"state"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Properties map[string]any `json:"properties"`
}

// SetState transitions the crumb to the specified state (prd003-crumbs-interface
// R4.2). It validates that state is one of the values defined in R2.1 and returns
// ErrInvalidState if not. The operation is idempotent: setting the current state
// succeeds without error. UpdatedAt is always refreshed.
func (c *Crumb) SetState(state string) error {
	if !validStates[state] {
		return ErrInvalidState
	}
	c.State = state
	c.UpdatedAt = time.Now()
	return nil
}

// Pebble transitions the crumb to the pebble state, marking it as successfully
// completed (prd003-crumbs-interface R4.3). The current state must be "taken";
// otherwise ErrInvalidTransition is returned.
func (c *Crumb) Pebble() error {
	if c.State != StateTaken {
		return ErrInvalidTransition
	}
	c.State = StatePebble
	c.UpdatedAt = time.Now()
	return nil
}

// Dust transitions the crumb to the dust state, marking it as failed or
// abandoned (prd003-crumbs-interface R4.4). Dust can be called from any state
// and is idempotent. UpdatedAt is always refreshed.
func (c *Crumb) Dust() error {
	c.State = StateDust
	c.UpdatedAt = time.Now()
	return nil
}
