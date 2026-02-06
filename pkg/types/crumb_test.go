package types

import (
	"errors"
	"testing"
	"time"
)

func TestCrumb_SetState(t *testing.T) {
	tests := []struct {
		name      string
		state     string
		wantErr   error
		wantState string
	}{
		{"valid draft", StateDraft, nil, StateDraft},
		{"valid pending", StatePending, nil, StatePending},
		{"valid ready", StateReady, nil, StateReady},
		{"valid taken", StateTaken, nil, StateTaken},
		{"valid pebble", StatePebble, nil, StatePebble},
		{"valid dust", StateDust, nil, StateDust},
		{"invalid state", "invalid", ErrInvalidState, StateDraft},
		{"empty state", "", ErrInvalidState, StateDraft},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Crumb{State: StateDraft}
			before := c.UpdatedAt

			err := c.SetState(tt.state)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SetState() error = %v, wantErr %v", err, tt.wantErr)
			}
			if c.State != tt.wantState {
				t.Errorf("SetState() state = %v, want %v", c.State, tt.wantState)
			}
			if err == nil && !c.UpdatedAt.After(before) {
				t.Error("SetState() should update UpdatedAt")
			}
		})
	}
}

func TestCrumb_SetState_Idempotent(t *testing.T) {
	c := &Crumb{State: StateReady}
	err := c.SetState(StateReady)
	if err != nil {
		t.Errorf("SetState() idempotent call should succeed, got %v", err)
	}
	if c.State != StateReady {
		t.Errorf("SetState() state should remain %v", StateReady)
	}
}

func TestCrumb_Pebble(t *testing.T) {
	tests := []struct {
		name         string
		initialState string
		wantErr      error
		wantState    string
	}{
		{"from taken", StateTaken, nil, StatePebble},
		{"from draft", StateDraft, ErrInvalidTransition, StateDraft},
		{"from pending", StatePending, ErrInvalidTransition, StatePending},
		{"from ready", StateReady, ErrInvalidTransition, StateReady},
		{"from pebble", StatePebble, ErrInvalidTransition, StatePebble},
		{"from dust", StateDust, ErrInvalidTransition, StateDust},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Crumb{State: tt.initialState}
			before := c.UpdatedAt

			err := c.Pebble()

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Pebble() error = %v, wantErr %v", err, tt.wantErr)
			}
			if c.State != tt.wantState {
				t.Errorf("Pebble() state = %v, want %v", c.State, tt.wantState)
			}
			if err == nil && !c.UpdatedAt.After(before) {
				t.Error("Pebble() should update UpdatedAt")
			}
		})
	}
}

func TestCrumb_Dust(t *testing.T) {
	states := []string{StateDraft, StatePending, StateReady, StateTaken, StatePebble, StateDust}

	for _, state := range states {
		t.Run("from "+state, func(t *testing.T) {
			c := &Crumb{State: state}
			before := c.UpdatedAt

			err := c.Dust()

			if err != nil {
				t.Errorf("Dust() should succeed from any state, got %v", err)
			}
			if c.State != StateDust {
				t.Errorf("Dust() state = %v, want %v", c.State, StateDust)
			}
			if !c.UpdatedAt.After(before) {
				t.Error("Dust() should update UpdatedAt")
			}
		})
	}
}

func TestCrumb_SetProperty(t *testing.T) {
	t.Run("sets property on nil map", func(t *testing.T) {
		c := &Crumb{}
		before := c.UpdatedAt

		err := c.SetProperty("priority", int64(3))

		if err != nil {
			t.Errorf("SetProperty() error = %v", err)
		}
		if c.Properties == nil {
			t.Error("SetProperty() should initialize Properties map")
		}
		if c.Properties["priority"] != int64(3) {
			t.Errorf("SetProperty() value = %v, want %v", c.Properties["priority"], int64(3))
		}
		if !c.UpdatedAt.After(before) {
			t.Error("SetProperty() should update UpdatedAt")
		}
	})

	t.Run("overwrites existing property", func(t *testing.T) {
		c := &Crumb{Properties: map[string]any{"priority": int64(1)}}

		err := c.SetProperty("priority", int64(5))

		if err != nil {
			t.Errorf("SetProperty() error = %v", err)
		}
		if c.Properties["priority"] != int64(5) {
			t.Errorf("SetProperty() value = %v, want %v", c.Properties["priority"], int64(5))
		}
	})
}

func TestCrumb_GetProperty(t *testing.T) {
	t.Run("returns property value", func(t *testing.T) {
		c := &Crumb{Properties: map[string]any{"priority": int64(3)}}

		val, err := c.GetProperty("priority")

		if err != nil {
			t.Errorf("GetProperty() error = %v", err)
		}
		if val != int64(3) {
			t.Errorf("GetProperty() value = %v, want %v", val, int64(3))
		}
	})

	t.Run("returns error for missing property", func(t *testing.T) {
		c := &Crumb{Properties: map[string]any{}}

		_, err := c.GetProperty("missing")

		if !errors.Is(err, ErrPropertyNotFound) {
			t.Errorf("GetProperty() error = %v, want %v", err, ErrPropertyNotFound)
		}
	})

	t.Run("returns error for nil map", func(t *testing.T) {
		c := &Crumb{}

		_, err := c.GetProperty("priority")

		if !errors.Is(err, ErrPropertyNotFound) {
			t.Errorf("GetProperty() error = %v, want %v", err, ErrPropertyNotFound)
		}
	})
}

func TestCrumb_GetProperties(t *testing.T) {
	t.Run("returns properties map", func(t *testing.T) {
		props := map[string]any{"priority": int64(3), "status": "active"}
		c := &Crumb{Properties: props}

		result := c.GetProperties()

		if len(result) != 2 {
			t.Errorf("GetProperties() len = %v, want 2", len(result))
		}
	})

	t.Run("returns empty map for nil", func(t *testing.T) {
		c := &Crumb{}

		result := c.GetProperties()

		if result == nil {
			t.Error("GetProperties() should return non-nil map")
		}
		if len(result) != 0 {
			t.Errorf("GetProperties() len = %v, want 0", len(result))
		}
	})
}

func TestCrumb_ClearProperty(t *testing.T) {
	t.Run("removes existing property", func(t *testing.T) {
		c := &Crumb{
			Properties: map[string]any{"priority": int64(3)},
			UpdatedAt:  time.Now().Add(-time.Hour),
		}
		before := c.UpdatedAt

		err := c.ClearProperty("priority")

		if err != nil {
			t.Errorf("ClearProperty() error = %v", err)
		}
		if _, ok := c.Properties["priority"]; ok {
			t.Error("ClearProperty() should remove property")
		}
		if !c.UpdatedAt.After(before) {
			t.Error("ClearProperty() should update UpdatedAt")
		}
	})

	t.Run("returns error for missing property", func(t *testing.T) {
		c := &Crumb{Properties: map[string]any{}}

		err := c.ClearProperty("missing")

		if !errors.Is(err, ErrPropertyNotFound) {
			t.Errorf("ClearProperty() error = %v, want %v", err, ErrPropertyNotFound)
		}
	})

	t.Run("returns error for nil map", func(t *testing.T) {
		c := &Crumb{}

		err := c.ClearProperty("priority")

		if !errors.Is(err, ErrPropertyNotFound) {
			t.Errorf("ClearProperty() error = %v, want %v", err, ErrPropertyNotFound)
		}
	})
}
