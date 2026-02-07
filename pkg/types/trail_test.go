package types

import (
	"errors"
	"testing"
)

func TestTrail_Complete(t *testing.T) {
	tests := []struct {
		name         string
		initialState string
		wantErr      error
		wantState    string
	}{
		{"from active", TrailStateActive, nil, TrailStateCompleted},
		{"from draft", TrailStateDraft, ErrInvalidState, TrailStateDraft},
		{"from pending", TrailStatePending, ErrInvalidState, TrailStatePending},
		{"from completed", TrailStateCompleted, ErrInvalidState, TrailStateCompleted},
		{"from abandoned", TrailStateAbandoned, ErrInvalidState, TrailStateAbandoned},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trail := &Trail{State: tt.initialState}

			err := trail.Complete()

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if trail.State != tt.wantState {
				t.Errorf("Complete() state = %v, want %v", trail.State, tt.wantState)
			}
			if err == nil && trail.CompletedAt == nil {
				t.Error("Complete() should set CompletedAt")
			}
		})
	}
}

func TestTrail_Abandon(t *testing.T) {
	tests := []struct {
		name         string
		initialState string
		wantErr      error
		wantState    string
	}{
		{"from active", TrailStateActive, nil, TrailStateAbandoned},
		{"from draft", TrailStateDraft, ErrInvalidState, TrailStateDraft},
		{"from pending", TrailStatePending, ErrInvalidState, TrailStatePending},
		{"from completed", TrailStateCompleted, ErrInvalidState, TrailStateCompleted},
		{"from abandoned", TrailStateAbandoned, ErrInvalidState, TrailStateAbandoned},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trail := &Trail{State: tt.initialState}

			err := trail.Abandon()

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Abandon() error = %v, wantErr %v", err, tt.wantErr)
			}
			if trail.State != tt.wantState {
				t.Errorf("Abandon() state = %v, want %v", trail.State, tt.wantState)
			}
			if err == nil && trail.CompletedAt == nil {
				t.Error("Abandon() should set CompletedAt")
			}
		})
	}
}

func TestTrail_SetState(t *testing.T) {
	tests := []struct {
		name         string
		initialState string
		targetState  string
		wantErr      error
		wantState    string
	}{
		// Valid transitions from draft
		{"draft to pending", TrailStateDraft, TrailStatePending, nil, TrailStatePending},
		{"draft to active", TrailStateDraft, TrailStateActive, nil, TrailStateActive},

		// Invalid transitions from draft
		{"draft to completed", TrailStateDraft, TrailStateCompleted, ErrInvalidState, TrailStateDraft},
		{"draft to abandoned", TrailStateDraft, TrailStateAbandoned, ErrInvalidState, TrailStateDraft},

		// Valid transitions from pending
		{"pending to active", TrailStatePending, TrailStateActive, nil, TrailStateActive},

		// Invalid transitions from pending
		{"pending to draft", TrailStatePending, TrailStateDraft, ErrInvalidState, TrailStatePending},
		{"pending to completed", TrailStatePending, TrailStateCompleted, ErrInvalidState, TrailStatePending},
		{"pending to abandoned", TrailStatePending, TrailStateAbandoned, ErrInvalidState, TrailStatePending},

		// Valid transitions from active (via SetState, though Complete/Abandon are preferred)
		{"active to completed", TrailStateActive, TrailStateCompleted, nil, TrailStateCompleted},
		{"active to abandoned", TrailStateActive, TrailStateAbandoned, nil, TrailStateAbandoned},

		// Invalid transitions from active
		{"active to draft", TrailStateActive, TrailStateDraft, ErrInvalidState, TrailStateActive},
		{"active to pending", TrailStateActive, TrailStatePending, ErrInvalidState, TrailStateActive},

		// Terminal states cannot transition
		{"completed to active", TrailStateCompleted, TrailStateActive, ErrInvalidState, TrailStateCompleted},
		{"completed to draft", TrailStateCompleted, TrailStateDraft, ErrInvalidState, TrailStateCompleted},
		{"abandoned to active", TrailStateAbandoned, TrailStateActive, ErrInvalidState, TrailStateAbandoned},
		{"abandoned to draft", TrailStateAbandoned, TrailStateDraft, ErrInvalidState, TrailStateAbandoned},

		// Empty state (new trail) can be set to any valid state
		{"empty to draft", "", TrailStateDraft, nil, TrailStateDraft},
		{"empty to pending", "", TrailStatePending, nil, TrailStatePending},
		{"empty to active", "", TrailStateActive, nil, TrailStateActive},
		{"empty to completed", "", TrailStateCompleted, nil, TrailStateCompleted},
		{"empty to abandoned", "", TrailStateAbandoned, nil, TrailStateAbandoned},

		// Invalid target state
		{"draft to invalid", TrailStateDraft, "invalid", ErrInvalidState, TrailStateDraft},
		{"empty to invalid", "", "invalid", ErrInvalidState, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trail := &Trail{State: tt.initialState}

			err := trail.SetState(tt.targetState)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SetState() error = %v, wantErr %v", err, tt.wantErr)
			}
			if trail.State != tt.wantState {
				t.Errorf("SetState() state = %v, want %v", trail.State, tt.wantState)
			}
		})
	}
}

func TestTrail_StateConstants(t *testing.T) {
	// Verify state constants have expected values
	if TrailStateDraft != "draft" {
		t.Errorf("TrailStateDraft = %q, want 'draft'", TrailStateDraft)
	}
	if TrailStatePending != "pending" {
		t.Errorf("TrailStatePending = %q, want 'pending'", TrailStatePending)
	}
	if TrailStateActive != "active" {
		t.Errorf("TrailStateActive = %q, want 'active'", TrailStateActive)
	}
	if TrailStateCompleted != "completed" {
		t.Errorf("TrailStateCompleted = %q, want 'completed'", TrailStateCompleted)
	}
	if TrailStateAbandoned != "abandoned" {
		t.Errorf("TrailStateAbandoned = %q, want 'abandoned'", TrailStateAbandoned)
	}
}
