package types

import "time"

// Trail represents an exploratory work session that groups crumbs.
// Implements: prd006-trails-interface (R1: struct, R2: states, R5: Complete, R6: Abandon).
type Trail struct {
	TrailID     string     `json:"trail_id"`
	State       string     `json:"state"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// Trail state constants per prd006-trails-interface R2.1.
const (
	TrailStateDraft     = "draft"
	TrailStatePending   = "pending"
	TrailStateActive    = "active"
	TrailStateCompleted = "completed"
	TrailStateAbandoned = "abandoned"
)

// Complete marks the trail as finished.
// Returns ErrInvalidState if the trail is not in active state.
// The caller must persist via Table.Set; the backend performs cascade
// operations (removing belongs_to links).
// See prd006-trails-interface R5.
func (t *Trail) Complete() error {
	if t.State != TrailStateActive {
		return ErrInvalidState
	}
	t.State = TrailStateCompleted
	now := time.Now()
	t.CompletedAt = &now
	return nil
}

// Abandon marks the trail as discarded.
// Returns ErrInvalidState if the trail is not in active state.
// The caller must persist via Table.Set; the backend performs cascade
// operations (deleting crumbs that belong to this trail).
// See prd006-trails-interface R6.
func (t *Trail) Abandon() error {
	if t.State != TrailStateActive {
		return ErrInvalidState
	}
	t.State = TrailStateAbandoned
	now := time.Now()
	t.CompletedAt = &now
	return nil
}
