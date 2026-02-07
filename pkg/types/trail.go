// Trail entity represents an exploratory work session.
// Implements: prd-trails-interface R1, R2 (Trail struct, state values);
//
//	docs/ARCHITECTURE § Main Interface.
package types

import (
	"slices"
	"time"
)

// Trail state values.
const (
	TrailStateDraft     = "draft"
	TrailStatePending   = "pending"
	TrailStateActive    = "active"
	TrailStateCompleted = "completed"
	TrailStateAbandoned = "abandoned"
)

// validTrailStates lists all valid trail state values.
var validTrailStates = []string{
	TrailStateDraft, TrailStatePending, TrailStateActive,
	TrailStateCompleted, TrailStateAbandoned,
}

// Trail represents an exploratory work session that groups crumbs.
// Trail branching uses branches_from links in the links table (ARCHITECTURE Decision 10).
type Trail struct {
	// TrailID is a UUID v7, generated on creation.
	TrailID string

	// State is the trail state (draft, pending, active, completed, abandoned).
	State string

	// CreatedAt is the timestamp of creation.
	CreatedAt time.Time

	// CompletedAt is the timestamp when completed or abandoned; nil if active.
	CompletedAt *time.Time
}

// SetState transitions the trail to the specified state.
// Returns ErrInvalidState if the state is not recognized or the transition is not allowed.
// State transitions per prd-trails-interface R2.3:
//   - draft → pending, active
//   - pending → active
//   - active → completed, abandoned (via Complete/Abandon methods)
//   - completed, abandoned are terminal
//
// Caller must save via Table.Set.
func (t *Trail) SetState(state string) error {
	if !slices.Contains(validTrailStates, state) {
		return ErrInvalidState
	}

	// Terminal states cannot transition
	if t.State == TrailStateCompleted || t.State == TrailStateAbandoned {
		return ErrInvalidState
	}

	// Validate allowed transitions
	switch t.State {
	case TrailStateDraft:
		// draft → pending or active only
		if state != TrailStatePending && state != TrailStateActive {
			return ErrInvalidState
		}
	case TrailStatePending:
		// pending → active only
		if state != TrailStateActive {
			return ErrInvalidState
		}
	case TrailStateActive:
		// active → completed or abandoned (should use Complete/Abandon methods)
		if state != TrailStateCompleted && state != TrailStateAbandoned {
			return ErrInvalidState
		}
	case "":
		// Empty state (new trail) can be set to any valid state
	default:
		// Unknown current state
		return ErrInvalidState
	}

	t.State = state
	return nil
}

// Complete marks the trail as completed.
// Returns ErrInvalidState if the trail is not in active state.
// Sets CompletedAt to now. Caller must save via Table.Set.
// When persisted, the backend removes belongs_to links so crumbs become permanent.
func (t *Trail) Complete() error {
	if t.State != TrailStateActive {
		return ErrInvalidState
	}
	t.State = TrailStateCompleted
	now := time.Now()
	t.CompletedAt = &now
	return nil
}

// Abandon marks the trail as abandoned.
// Returns ErrInvalidState if the trail is not in active state.
// Sets CompletedAt to now. Caller must save via Table.Set.
// When persisted, the backend deletes all crumbs belonging to this trail.
func (t *Trail) Abandon() error {
	if t.State != TrailStateActive {
		return ErrInvalidState
	}
	t.State = TrailStateAbandoned
	now := time.Now()
	t.CompletedAt = &now
	return nil
}
