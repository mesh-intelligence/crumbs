// Tests for rel01.0-uc003-crumb-lifecycle: crumb state machine
// (draft→pending→ready→taken→pebble and any→dust), timestamp behavior,
// state-based filtering, and full success/failure paths.
// Implements: test-rel01.0-uc003-crumb-lifecycle.yaml.
package integration

import (
	"testing"
	"time"

	"github.com/mesh-intelligence/crumbs/pkg/types"
)

func TestCrumbCreation(t *testing.T) {
	tests := []struct {
		name         string
		crumbName    string
		initialState string
	}{
		{"create crumb with draft state and initialized timestamps", "New crumb", types.CrumbStateDraft},
		{"create crumb with explicit ready state", "Ready crumb", types.CrumbStateReady},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := setupCupboard(t)
			tbl := mustGetTable(t, b, types.TableCrumbs)

			id := mustCreateCrumb(t, tbl, tt.crumbName, tt.initialState)
			if !isUUIDv7(id) {
				t.Errorf("expected UUID v7, got %q", id)
			}

			c := mustGetCrumb(t, tbl, id)
			if c.State != tt.initialState {
				t.Errorf("expected state %q, got %q", tt.initialState, c.State)
			}
			if c.CreatedAt.IsZero() {
				t.Error("expected CreatedAt to be set")
			}
			if c.UpdatedAt.IsZero() {
				t.Error("expected UpdatedAt to be set")
			}
		})
	}
}

func TestCrumbStateTransitions(t *testing.T) {
	tests := []struct {
		name      string
		fromState string
		toState   string
	}{
		{"transition draft to pending", types.CrumbStateDraft, types.CrumbStatePending},
		{"transition pending to ready", types.CrumbStatePending, types.CrumbStateReady},
		{"transition draft to ready", types.CrumbStateDraft, types.CrumbStateReady},
		{"transition ready to taken", types.CrumbStateReady, types.CrumbStateTaken},
		{"transition taken to pebble", types.CrumbStateTaken, types.CrumbStatePebble},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := setupCupboard(t)
			tbl := mustGetTable(t, b, types.TableCrumbs)

			id := mustCreateCrumb(t, tbl, "Test crumb", tt.fromState)
			c := mustGetCrumb(t, tbl, id)
			c.State = tt.toState
			if _, err := tbl.Set(id, c); err != nil {
				t.Fatalf("Set to %q: %v", tt.toState, err)
			}

			got := mustGetCrumb(t, tbl, id)
			if got.State != tt.toState {
				t.Errorf("expected state %q, got %q", tt.toState, got.State)
			}
		})
	}
}

func TestCrumbDustTransitions(t *testing.T) {
	tests := []struct {
		name      string
		fromState string
	}{
		{"transition draft to dust", types.CrumbStateDraft},
		{"transition pending to dust", types.CrumbStatePending},
		{"transition ready to dust", types.CrumbStateReady},
		{"transition taken to dust", types.CrumbStateTaken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := setupCupboard(t)
			tbl := mustGetTable(t, b, types.TableCrumbs)

			id := mustCreateCrumb(t, tbl, "Dust crumb", tt.fromState)
			c := mustGetCrumb(t, tbl, id)
			c.State = types.CrumbStateDust
			if _, err := tbl.Set(id, c); err != nil {
				t.Fatalf("Set to dust: %v", err)
			}

			got := mustGetCrumb(t, tbl, id)
			if got.State != types.CrumbStateDust {
				t.Errorf("expected state dust, got %q", got.State)
			}
		})
	}
}

func TestCrumbTimestampTracking(t *testing.T) {
	t.Run("UpdatedAt advances on state transition", func(t *testing.T) {
		b, _ := setupCupboard(t)
		tbl := mustGetTable(t, b, types.TableCrumbs)

		id := mustCreateCrumb(t, tbl, "Timestamp crumb", types.CrumbStateDraft)
		original := mustGetCrumb(t, tbl, id)
		origCreatedAt := original.CreatedAt
		origUpdatedAt := original.UpdatedAt

		// Sleep to ensure timestamp difference.
		time.Sleep(1100 * time.Millisecond)

		c := mustGetCrumb(t, tbl, id)
		c.State = types.CrumbStateReady
		if _, err := tbl.Set(id, c); err != nil {
			t.Fatalf("Set: %v", err)
		}

		got := mustGetCrumb(t, tbl, id)
		if got.State != types.CrumbStateReady {
			t.Errorf("expected state ready, got %q", got.State)
		}
		if !got.UpdatedAt.After(origUpdatedAt) {
			t.Errorf("expected UpdatedAt to advance: original=%v, new=%v", origUpdatedAt, got.UpdatedAt)
		}
		if !got.CreatedAt.Equal(origCreatedAt) {
			t.Errorf("expected CreatedAt unchanged: original=%v, new=%v", origCreatedAt, got.CreatedAt)
		}
	})
}

func TestCrumbFetchByState(t *testing.T) {
	tests := []struct {
		name         string
		setupStates  []string
		filterState  string
		wantCount    int
		wantNames    []string
		excludeNames []string
	}{
		{
			name:         "fetch by single state returns matching crumbs only",
			setupStates:  []string{types.CrumbStateDraft, types.CrumbStateReady, types.CrumbStateTaken},
			filterState:  types.CrumbStateReady,
			wantCount:    1,
			wantNames:    []string{"Ready crumb"},
			excludeNames: []string{"Draft crumb", "Taken crumb"},
		},
		{
			name:         "filter excludes pebble crumbs from non-terminal state list",
			setupStates:  []string{types.CrumbStateDraft, types.CrumbStatePebble},
			filterState:  types.CrumbStateDraft,
			wantCount:    1,
			wantNames:    []string{"Draft crumb"},
			excludeNames: []string{"Pebble crumb"},
		},
		{
			name:         "filter excludes dust crumbs from non-terminal state list",
			setupStates:  []string{types.CrumbStateDraft, types.CrumbStateDust},
			filterState:  types.CrumbStateDraft,
			wantCount:    1,
			wantNames:    []string{"Draft crumb"},
			excludeNames: []string{"Dust crumb"},
		},
		{
			name:         "filter for pebble state returns only pebble crumbs",
			setupStates:  []string{types.CrumbStateDraft, types.CrumbStatePebble, types.CrumbStateDust},
			filterState:  types.CrumbStatePebble,
			wantCount:    1,
			wantNames:    []string{"Pebble crumb"},
			excludeNames: []string{"Draft crumb", "Dust crumb"},
		},
		{
			name:         "filter for dust state returns only dust crumbs",
			setupStates:  []string{types.CrumbStateDraft, types.CrumbStatePebble, types.CrumbStateDust},
			filterState:  types.CrumbStateDust,
			wantCount:    1,
			wantNames:    []string{"Dust crumb"},
			excludeNames: []string{"Draft crumb", "Pebble crumb"},
		},
	}

	stateToName := map[string]string{
		types.CrumbStateDraft:   "Draft crumb",
		types.CrumbStateReady:   "Ready crumb",
		types.CrumbStateTaken:   "Taken crumb",
		types.CrumbStatePebble:  "Pebble crumb",
		types.CrumbStateDust:    "Dust crumb",
		types.CrumbStatePending: "Pending crumb",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := setupCupboard(t)
			tbl := mustGetTable(t, b, types.TableCrumbs)

			for _, state := range tt.setupStates {
				name := stateToName[state]
				mustCreateCrumb(t, tbl, name, state)
			}

			results := fetchByStates(t, tbl, []string{tt.filterState})
			if len(results) != tt.wantCount {
				t.Errorf("expected %d results, got %d", tt.wantCount, len(results))
			}

			resultNames := make(map[string]bool)
			for _, r := range results {
				c := r.(*types.Crumb)
				resultNames[c.Name] = true
			}

			for _, name := range tt.wantNames {
				if !resultNames[name] {
					t.Errorf("expected %q in results", name)
				}
			}
			for _, name := range tt.excludeNames {
				if resultNames[name] {
					t.Errorf("did not expect %q in results", name)
				}
			}
		})
	}
}

func TestCrumbFetchAllStates(t *testing.T) {
	t.Run("list without filter returns all crumbs", func(t *testing.T) {
		b, _ := setupCupboard(t)
		tbl := mustGetTable(t, b, types.TableCrumbs)

		mustCreateCrumb(t, tbl, "Draft crumb", types.CrumbStateDraft)
		mustCreateCrumb(t, tbl, "Ready crumb", types.CrumbStateReady)
		mustCreateCrumb(t, tbl, "Taken crumb", types.CrumbStateTaken)

		results := fetchAll(t, tbl)
		if len(results) != 3 {
			t.Errorf("expected 3 crumbs, got %d", len(results))
		}

		names := make(map[string]bool)
		for _, r := range results {
			c := r.(*types.Crumb)
			names[c.Name] = true
		}
		for _, want := range []string{"Draft crumb", "Ready crumb", "Taken crumb"} {
			if !names[want] {
				t.Errorf("expected %q in results", want)
			}
		}
	})
}

func TestFullSuccessPath(t *testing.T) {
	t.Run("full success path draft to pebble", func(t *testing.T) {
		b, _ := setupCupboard(t)
		tbl := mustGetTable(t, b, types.TableCrumbs)

		id := mustCreateCrumb(t, tbl, "Success path crumb", types.CrumbStateDraft)

		// Verify initial state.
		c := mustGetCrumb(t, tbl, id)
		if c.State != types.CrumbStateDraft {
			t.Fatalf("expected initial state draft, got %q", c.State)
		}

		// Transition through states: draft → pending → ready → taken → pebble.
		transitions := []string{
			types.CrumbStatePending,
			types.CrumbStateReady,
			types.CrumbStateTaken,
			types.CrumbStatePebble,
		}

		for _, state := range transitions {
			c = mustGetCrumb(t, tbl, id)
			c.State = state
			if _, err := tbl.Set(id, c); err != nil {
				t.Fatalf("Set to %q: %v", state, err)
			}
		}

		got := mustGetCrumb(t, tbl, id)
		if got.State != types.CrumbStatePebble {
			t.Errorf("expected state pebble, got %q", got.State)
		}
	})
}

func TestFullFailurePath(t *testing.T) {
	t.Run("full failure path draft to dust", func(t *testing.T) {
		b, _ := setupCupboard(t)
		tbl := mustGetTable(t, b, types.TableCrumbs)

		id := mustCreateCrumb(t, tbl, "Failure path crumb", types.CrumbStateDraft)

		// Verify initial state.
		c := mustGetCrumb(t, tbl, id)
		if c.State != types.CrumbStateDraft {
			t.Fatalf("expected initial state draft, got %q", c.State)
		}

		// Transition directly to dust.
		c.State = types.CrumbStateDust
		if _, err := tbl.Set(id, c); err != nil {
			t.Fatalf("Set to dust: %v", err)
		}

		got := mustGetCrumb(t, tbl, id)
		if got.State != types.CrumbStateDust {
			t.Errorf("expected state dust, got %q", got.State)
		}
	})
}

func TestMixedTerminalStates(t *testing.T) {
	t.Run("mixed terminal states in same table", func(t *testing.T) {
		b, _ := setupCupboard(t)
		tbl := mustGetTable(t, b, types.TableCrumbs)

		mustCreateCrumb(t, tbl, "Draft crumb", types.CrumbStateDraft)
		mustCreateCrumb(t, tbl, "Pebble crumb", types.CrumbStatePebble)
		mustCreateCrumb(t, tbl, "Dust crumb", types.CrumbStateDust)

		all := fetchAll(t, tbl)
		if len(all) != 3 {
			t.Errorf("expected 3 crumbs total, got %d", len(all))
		}

		drafts := fetchByStates(t, tbl, []string{types.CrumbStateDraft})
		if len(drafts) != 1 {
			t.Errorf("expected 1 draft, got %d", len(drafts))
		}

		pebbles := fetchByStates(t, tbl, []string{types.CrumbStatePebble})
		if len(pebbles) != 1 {
			t.Errorf("expected 1 pebble, got %d", len(pebbles))
		}

		dusts := fetchByStates(t, tbl, []string{types.CrumbStateDust})
		if len(dusts) != 1 {
			t.Errorf("expected 1 dust, got %d", len(dusts))
		}
	})
}
