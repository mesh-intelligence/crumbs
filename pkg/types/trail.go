package types

import "time"

// Trail states (prd006-trails-interface).
const (
	TrailDraft     = "draft"
	TrailPending   = "pending"
	TrailActive    = "active"
	TrailCompleted = "completed"
	TrailAbandoned = "abandoned"
)

// validTrailStates is the set of recognized trail state values.
var validTrailStates = map[string]bool{
	TrailDraft:     true,
	TrailPending:   true,
	TrailActive:    true,
	TrailCompleted: true,
	TrailAbandoned: true,
}

// Trail represents an exploration session that groups crumbs
// (prd006-trails-interface).
type Trail struct {
	TrailID     string     `json:"trail_id"`
	State       string     `json:"state"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// SetState transitions the trail to the specified state.
// Returns ErrInvalidState if the state is not recognized.
func (t *Trail) SetState(state string) error {
	if !validTrailStates[state] {
		return ErrInvalidState
	}
	t.State = state
	return nil
}

// Complete marks the trail as finished. Sets State to "completed" and
// CompletedAt to now. Returns ErrInvalidTransition if the current state
// is not "active".
// (prd006-trails-interface)
func (t *Trail) Complete() error {
	if t.State != TrailActive {
		return ErrInvalidTransition
	}
	t.State = TrailCompleted
	now := time.Now()
	t.CompletedAt = &now
	return nil
}

// Abandon marks the trail as discarded. Sets State to "abandoned" and
// CompletedAt to now. Returns ErrInvalidTransition if the trail is
// already in a terminal state (completed or abandoned).
// (prd006-trails-interface)
func (t *Trail) Abandon() error {
	if t.State == TrailCompleted || t.State == TrailAbandoned {
		return ErrInvalidTransition
	}
	t.State = TrailAbandoned
	now := time.Now()
	t.CompletedAt = &now
	return nil
}
