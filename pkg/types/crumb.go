// Implements: prd003-crumbs-interface (R1: Crumb struct, R2: State values);
//             prd001-cupboard-core (R8: UUID v7 identifiers);
//             docs/ARCHITECTURE § Lifecycle of entities.
package types

import "time"

// Crumb represents a work item in the Crumbs system.
// Entity methods modify the struct in memory; callers must call Table.Set
// to persist changes (prd003-crumbs-interface R4.5).
type Crumb struct {
	CrumbID   string    `json:"crumb_id"`
	Name      string    `json:"name"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Crumb state constants (prd003-crumbs-interface R2.1).
const (
	StateDraft   = "draft"
	StatePending = "pending"
	StateReady   = "ready"
	StateTaken   = "taken"
	StatePebble  = "pebble"
	StateDust    = "dust"
)

// validStates is the set of recognized crumb states.
var validStates = map[string]bool{
	StateDraft:   true,
	StatePending: true,
	StateReady:   true,
	StateTaken:   true,
	StatePebble:  true,
	StateDust:    true,
}

// validTransitions maps each state to the set of states it can transition to.
// The "any→dust" rule is handled separately in ValidateTransition.
var validTransitions = map[string]map[string]bool{
	StateDraft:   {StatePending: true},
	StatePending: {StateReady: true},
	StateReady:   {StateTaken: true},
	StateTaken:   {StatePebble: true, StateDust: true},
}

// ValidateTransition checks whether a state transition from → to is allowed.
// Returns ErrInvalidState if either state is unrecognized, ErrInvalidTransition
// if the transition is not permitted. The "any→dust" transition is always valid
// (prd003-crumbs-interface R4.4). A self-transition (from == to) is valid
// (prd003-crumbs-interface R4.2).
func ValidateTransition(from, to string) error {
	if !validStates[from] || !validStates[to] {
		return ErrInvalidState
	}
	if from == to {
		return nil
	}
	if to == StateDust {
		return nil
	}
	if allowed, ok := validTransitions[from]; ok && allowed[to] {
		return nil
	}
	return ErrInvalidTransition
}
