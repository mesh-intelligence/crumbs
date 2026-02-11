// Tests for crumbs Table: Set, Get, Delete, Fetch, JSONL roundtrip, error paths.
// Validates: prd002-sqlite-backend (R13: Table Interface, R14.2: Crumb hydration, R15: Entity Persistence);
//
//	prd003-crumbs-interface (R1: Crumb struct, R3: Creating, R6: Retrieving, R7: Updating,
//	R8: Deleting, R9: Filter Map);
//	test-rel01.0-uc002-table-crud; test-rel01.0-uc003-crumb-lifecycle.
package sqlite

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mesh-intelligence/crumbs/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getCrumbsTable returns the crumbs table from a fresh test cupboard.
func getCrumbsTable(t *testing.T) (*Backend, types.Table) {
	t.Helper()
	b := newTestCupboard(t)
	table, err := b.GetTable(types.TableCrumbs)
	require.NoError(t, err)
	return b, table
}

func TestSetCreateWithUUID(t *testing.T) {
	_, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "Test crumb"})
	require.NoError(t, err)
	assert.NotEmpty(t, id, "Set must return a generated UUID")
	assert.Len(t, id, 36, "UUID should be 36 characters")
}

func TestSetCreateSetsDefaultState(t *testing.T) {
	_, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "Draft crumb"})
	require.NoError(t, err)

	entity, err := table.Get(id)
	require.NoError(t, err)
	crumb := entity.(*types.Crumb)
	assert.Equal(t, types.StateDraft, crumb.State, "new crumb must default to draft state")
}

func TestSetCreateSetsTimestamps(t *testing.T) {
	_, table := getCrumbsTable(t)

	before := time.Now().UTC().Add(-time.Second)
	id, err := table.Set("", &types.Crumb{Name: "Timestamp crumb"})
	require.NoError(t, err)
	after := time.Now().UTC().Add(time.Second)

	entity, err := table.Get(id)
	require.NoError(t, err)
	crumb := entity.(*types.Crumb)

	assert.True(t, crumb.CreatedAt.After(before), "CreatedAt should be after test start")
	assert.True(t, crumb.CreatedAt.Before(after), "CreatedAt should be before test end")
	assert.True(t, crumb.UpdatedAt.After(before), "UpdatedAt should be after test start")
	assert.True(t, crumb.UpdatedAt.Before(after), "UpdatedAt should be before test end")
}

func TestSetCreateTwoUniqueIDs(t *testing.T) {
	_, table := getCrumbsTable(t)

	id1, err := table.Set("", &types.Crumb{Name: "First"})
	require.NoError(t, err)
	id2, err := table.Set("", &types.Crumb{Name: "Second"})
	require.NoError(t, err)

	assert.NotEqual(t, id1, id2, "two creates must generate unique IDs")
}

func TestSetUpdateExisting(t *testing.T) {
	_, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "Original"})
	require.NoError(t, err)

	returnedID, err := table.Set(id, &types.Crumb{Name: "Updated", State: types.StateDraft})
	require.NoError(t, err)
	assert.Equal(t, id, returnedID)

	entity, err := table.Get(id)
	require.NoError(t, err)
	crumb := entity.(*types.Crumb)
	assert.Equal(t, "Updated", crumb.Name)
}

func TestSetUpdateAdvancesUpdatedAt(t *testing.T) {
	_, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "Timestamp test"})
	require.NoError(t, err)

	entity, err := table.Get(id)
	require.NoError(t, err)
	original := entity.(*types.Crumb)
	originalUpdatedAt := original.UpdatedAt

	// Small sleep to ensure timestamp difference.
	time.Sleep(10 * time.Millisecond)

	_, err = table.Set(id, &types.Crumb{Name: "Timestamp updated", State: types.StateDraft})
	require.NoError(t, err)

	entity, err = table.Get(id)
	require.NoError(t, err)
	updated := entity.(*types.Crumb)
	assert.True(t, !updated.UpdatedAt.Before(originalUpdatedAt),
		"UpdatedAt must not be before original: got %v, original %v", updated.UpdatedAt, originalUpdatedAt)
}

func TestSetInvalidDataType(t *testing.T) {
	_, table := getCrumbsTable(t)

	_, err := table.Set("", "not a crumb")
	assert.ErrorIs(t, err, types.ErrInvalidData)
}

func TestSetEmptyName(t *testing.T) {
	_, table := getCrumbsTable(t)

	_, err := table.Set("", &types.Crumb{Name: ""})
	assert.ErrorIs(t, err, types.ErrInvalidName)
}

func TestGetByID(t *testing.T) {
	_, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "Retrieve me"})
	require.NoError(t, err)

	entity, err := table.Get(id)
	require.NoError(t, err)

	crumb, ok := entity.(*types.Crumb)
	require.True(t, ok, "Get must return *types.Crumb")
	assert.Equal(t, id, crumb.CrumbID)
	assert.Equal(t, "Retrieve me", crumb.Name)
	assert.Equal(t, types.StateDraft, crumb.State)
	assert.False(t, crumb.CreatedAt.IsZero())
	assert.False(t, crumb.UpdatedAt.IsZero())
}

func TestGetNotFound(t *testing.T) {
	_, table := getCrumbsTable(t)

	_, err := table.Get("nonexistent-uuid-12345")
	assert.ErrorIs(t, err, types.ErrNotFound)
}

func TestGetEmptyID(t *testing.T) {
	_, table := getCrumbsTable(t)

	_, err := table.Get("")
	assert.ErrorIs(t, err, types.ErrInvalidID)
}

func TestDeleteByID(t *testing.T) {
	_, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "Delete me"})
	require.NoError(t, err)

	err = table.Delete(id)
	assert.NoError(t, err)

	// Subsequent Get must return not found.
	_, err = table.Get(id)
	assert.ErrorIs(t, err, types.ErrNotFound)
}

func TestDeleteNotFound(t *testing.T) {
	_, table := getCrumbsTable(t)

	err := table.Delete("nonexistent-uuid-12345")
	assert.ErrorIs(t, err, types.ErrNotFound)
}

func TestDeleteEmptyID(t *testing.T) {
	_, table := getCrumbsTable(t)

	err := table.Delete("")
	assert.ErrorIs(t, err, types.ErrInvalidID)
}

func TestDeleteRemovesFromFetch(t *testing.T) {
	_, table := getCrumbsTable(t)

	id1, err := table.Set("", &types.Crumb{Name: "Keep this"})
	require.NoError(t, err)
	id2, err := table.Set("", &types.Crumb{Name: "Delete this"})
	require.NoError(t, err)

	err = table.Delete(id2)
	require.NoError(t, err)

	results, err := table.Fetch(nil)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, id1, results[0].(*types.Crumb).CrumbID)
}

func TestFetchAllEmpty(t *testing.T) {
	_, table := getCrumbsTable(t)

	results, err := table.Fetch(nil)
	require.NoError(t, err)
	assert.NotNil(t, results, "Fetch must return empty slice, not nil")
	assert.Len(t, results, 0)
}

func TestFetchAllReturnsAll(t *testing.T) {
	_, table := getCrumbsTable(t)

	_, err := table.Set("", &types.Crumb{Name: "Crumb A"})
	require.NoError(t, err)
	_, err = table.Set("", &types.Crumb{Name: "Crumb B"})
	require.NoError(t, err)
	_, err = table.Set("", &types.Crumb{Name: "Crumb C"})
	require.NoError(t, err)

	results, err := table.Fetch(nil)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestFetchEmptyFilter(t *testing.T) {
	_, table := getCrumbsTable(t)

	_, err := table.Set("", &types.Crumb{Name: "Crumb A"})
	require.NoError(t, err)
	_, err = table.Set("", &types.Crumb{Name: "Crumb B"})
	require.NoError(t, err)

	results, err := table.Fetch(map[string]any{})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestFetchWithStateFilter(t *testing.T) {
	_, table := getCrumbsTable(t)

	// Create crumbs with various states. The Set with empty ID always creates as
	// draft; we then update state via Set with ID.
	id1, err := table.Set("", &types.Crumb{Name: "Draft crumb"})
	require.NoError(t, err)

	id2, err := table.Set("", &types.Crumb{Name: "Ready crumb 1"})
	require.NoError(t, err)
	_, err = table.Set(id2, &types.Crumb{Name: "Ready crumb 1", State: types.StateReady})
	require.NoError(t, err)

	id3, err := table.Set("", &types.Crumb{Name: "Ready crumb 2"})
	require.NoError(t, err)
	_, err = table.Set(id3, &types.Crumb{Name: "Ready crumb 2", State: types.StateReady})
	require.NoError(t, err)

	_ = id1

	results, err := table.Fetch(map[string]any{"states": []string{types.StateReady}})
	require.NoError(t, err)
	assert.Len(t, results, 2)
	for _, r := range results {
		assert.Equal(t, types.StateReady, r.(*types.Crumb).State)
	}
}

func TestFetchWithStateFilterNoMatches(t *testing.T) {
	_, table := getCrumbsTable(t)

	_, err := table.Set("", &types.Crumb{Name: "Draft crumb"})
	require.NoError(t, err)

	results, err := table.Fetch(map[string]any{"states": []string{types.StatePebble}})
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestFetchInvalidFilterType(t *testing.T) {
	_, table := getCrumbsTable(t)

	_, err := table.Fetch(map[string]any{"states": "not a slice"})
	assert.ErrorIs(t, err, types.ErrInvalidFilter)
}

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		to       string
		wantState string
	}{
		{"draft to pending", types.StateDraft, types.StatePending, types.StatePending},
		{"pending to ready", types.StatePending, types.StateReady, types.StateReady},
		{"ready to taken", types.StateReady, types.StateTaken, types.StateTaken},
		{"taken to pebble", types.StateTaken, types.StatePebble, types.StatePebble},
		{"draft to dust", types.StateDraft, types.StateDust, types.StateDust},
		{"pending to dust", types.StatePending, types.StateDust, types.StateDust},
		{"ready to dust", types.StateReady, types.StateDust, types.StateDust},
		{"taken to dust", types.StateTaken, types.StateDust, types.StateDust},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, table := getCrumbsTable(t)

			// Create with initial state. Set with empty id forces draft,
			// so we create then update to the desired initial state.
			id, err := table.Set("", &types.Crumb{Name: "Transition test"})
			require.NoError(t, err)

			if tt.from != types.StateDraft {
				_, err = table.Set(id, &types.Crumb{Name: "Transition test", State: tt.from})
				require.NoError(t, err)
			}

			_, err = table.Set(id, &types.Crumb{Name: "Transition test", State: tt.to})
			require.NoError(t, err)

			entity, err := table.Get(id)
			require.NoError(t, err)
			assert.Equal(t, tt.wantState, entity.(*types.Crumb).State)
		})
	}
}

func TestFullSuccessPath(t *testing.T) {
	_, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "Success path"})
	require.NoError(t, err)

	states := []string{types.StatePending, types.StateReady, types.StateTaken, types.StatePebble}
	for _, s := range states {
		_, err = table.Set(id, &types.Crumb{Name: "Success path", State: s})
		require.NoError(t, err)
	}

	entity, err := table.Get(id)
	require.NoError(t, err)
	assert.Equal(t, types.StatePebble, entity.(*types.Crumb).State)
}

func TestFullFailurePath(t *testing.T) {
	_, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "Failure path"})
	require.NoError(t, err)

	_, err = table.Set(id, &types.Crumb{Name: "Failure path", State: types.StateDust})
	require.NoError(t, err)

	entity, err := table.Get(id)
	require.NoError(t, err)
	assert.Equal(t, types.StateDust, entity.(*types.Crumb).State)
}

func TestFetchByStateFiltering(t *testing.T) {
	_, table := getCrumbsTable(t)

	// Create crumbs with mixed states.
	_, err := table.Set("", &types.Crumb{Name: "Draft crumb"})
	require.NoError(t, err)

	id2, err := table.Set("", &types.Crumb{Name: "Pebble crumb"})
	require.NoError(t, err)
	_, err = table.Set(id2, &types.Crumb{Name: "Pebble crumb", State: types.StatePebble})
	require.NoError(t, err)

	id3, err := table.Set("", &types.Crumb{Name: "Dust crumb"})
	require.NoError(t, err)
	_, err = table.Set(id3, &types.Crumb{Name: "Dust crumb", State: types.StateDust})
	require.NoError(t, err)

	// Filter for draft only.
	results, err := table.Fetch(map[string]any{"states": []string{types.StateDraft}})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Draft crumb", results[0].(*types.Crumb).Name)

	// Filter for pebble only.
	results, err = table.Fetch(map[string]any{"states": []string{types.StatePebble}})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Pebble crumb", results[0].(*types.Crumb).Name)

	// Filter for dust only.
	results, err = table.Fetch(map[string]any{"states": []string{types.StateDust}})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Dust crumb", results[0].(*types.Crumb).Name)

	// Fetch all (no filter).
	results, err = table.Fetch(nil)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestJSONLPersistence(t *testing.T) {
	b, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "Persisted crumb"})
	require.NoError(t, err)

	// Verify crumbs.jsonl contains the crumb.
	jsonlPath := filepath.Join(b.config.DataDir, "crumbs.jsonl")
	data, err := os.ReadFile(jsonlPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Persisted crumb")
	assert.Contains(t, string(data), id)
}

func TestJSONLReflectsUpdate(t *testing.T) {
	b, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "Before update"})
	require.NoError(t, err)

	_, err = table.Set(id, &types.Crumb{Name: "After update", State: types.StateReady})
	require.NoError(t, err)

	jsonlPath := filepath.Join(b.config.DataDir, "crumbs.jsonl")
	data, err := os.ReadFile(jsonlPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "After update")
	assert.NotContains(t, string(data), "Before update")
}

func TestJSONLReflectsDelete(t *testing.T) {
	b, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "To be deleted"})
	require.NoError(t, err)
	_, err = table.Set("", &types.Crumb{Name: "To be kept"})
	require.NoError(t, err)

	err = table.Delete(id)
	require.NoError(t, err)

	jsonlPath := filepath.Join(b.config.DataDir, "crumbs.jsonl")
	data, err := os.ReadFile(jsonlPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "To be kept")
	assert.NotContains(t, string(data), "To be deleted")
}

func TestJSONLRoundtrip(t *testing.T) {
	dir := t.TempDir()

	// Phase 1: create crumbs, then detach.
	b1 := NewBackend()
	cfg := types.Config{Backend: types.BackendSQLite, DataDir: dir}
	err := b1.Attach(cfg)
	require.NoError(t, err)

	table1, err := b1.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id1, err := table1.Set("", &types.Crumb{Name: "Roundtrip crumb 1"})
	require.NoError(t, err)
	id2, err := table1.Set("", &types.Crumb{Name: "Roundtrip crumb 2"})
	require.NoError(t, err)

	// Update one crumb's state.
	_, err = table1.Set(id1, &types.Crumb{Name: "Roundtrip crumb 1", State: types.StateReady})
	require.NoError(t, err)

	err = b1.Detach()
	require.NoError(t, err)

	// Phase 2: delete cupboard.db to simulate fresh startup, re-attach.
	dbPath := filepath.Join(dir, "cupboard.db")
	os.Remove(dbPath)

	b2 := NewBackend()
	err = b2.Attach(cfg)
	require.NoError(t, err)
	defer b2.Detach()

	table2, err := b2.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	// Verify all data survived the roundtrip.
	entity1, err := table2.Get(id1)
	require.NoError(t, err)
	crumb1 := entity1.(*types.Crumb)
	assert.Equal(t, "Roundtrip crumb 1", crumb1.Name)
	assert.Equal(t, types.StateReady, crumb1.State)

	entity2, err := table2.Get(id2)
	require.NoError(t, err)
	crumb2 := entity2.(*types.Crumb)
	assert.Equal(t, "Roundtrip crumb 2", crumb2.Name)
	assert.Equal(t, types.StateDraft, crumb2.State)

	// Fetch should return both.
	results, err := table2.Fetch(nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestJSONLRoundtripAfterDBDeletion(t *testing.T) {
	dir := t.TempDir()
	cfg := types.Config{Backend: types.BackendSQLite, DataDir: dir}

	// Create data.
	b := NewBackend()
	err := b.Attach(cfg)
	require.NoError(t, err)
	table, err := b.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := table.Set("", &types.Crumb{Name: "Survive DB deletion"})
	require.NoError(t, err)
	err = b.Detach()
	require.NoError(t, err)

	// Verify JSONL file exists.
	jsonlPath := filepath.Join(dir, "crumbs.jsonl")
	_, err = os.Stat(jsonlPath)
	require.NoError(t, err)

	// Re-attach (Attach deletes and recreates cupboard.db per R4.1).
	b2 := NewBackend()
	err = b2.Attach(cfg)
	require.NoError(t, err)
	defer b2.Detach()

	table2, err := b2.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	entity, err := table2.Get(id)
	require.NoError(t, err)
	assert.Equal(t, "Survive DB deletion", entity.(*types.Crumb).Name)
}

func TestMixedTerminalStates(t *testing.T) {
	_, table := getCrumbsTable(t)

	_, err := table.Set("", &types.Crumb{Name: "Draft crumb"})
	require.NoError(t, err)

	id2, err := table.Set("", &types.Crumb{Name: "Pebble crumb"})
	require.NoError(t, err)
	_, err = table.Set(id2, &types.Crumb{Name: "Pebble crumb", State: types.StatePebble})
	require.NoError(t, err)

	id3, err := table.Set("", &types.Crumb{Name: "Dust crumb"})
	require.NoError(t, err)
	_, err = table.Set(id3, &types.Crumb{Name: "Dust crumb", State: types.StateDust})
	require.NoError(t, err)

	results, err := table.Fetch(nil)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	stateCounts := map[string]int{}
	for _, r := range results {
		stateCounts[r.(*types.Crumb).State]++
	}
	assert.Equal(t, 1, stateCounts[types.StateDraft])
	assert.Equal(t, 1, stateCounts[types.StatePebble])
	assert.Equal(t, 1, stateCounts[types.StateDust])
}

func TestFetchMultipleStatesFilter(t *testing.T) {
	_, table := getCrumbsTable(t)

	_, err := table.Set("", &types.Crumb{Name: "Draft crumb"})
	require.NoError(t, err)

	id2, err := table.Set("", &types.Crumb{Name: "Ready crumb"})
	require.NoError(t, err)
	_, err = table.Set(id2, &types.Crumb{Name: "Ready crumb", State: types.StateReady})
	require.NoError(t, err)

	id3, err := table.Set("", &types.Crumb{Name: "Pebble crumb"})
	require.NoError(t, err)
	_, err = table.Set(id3, &types.Crumb{Name: "Pebble crumb", State: types.StatePebble})
	require.NoError(t, err)

	// Filter for draft and ready.
	results, err := table.Fetch(map[string]any{"states": []string{types.StateDraft, types.StateReady}})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestRoundTripFieldFidelity(t *testing.T) {
	_, table := getCrumbsTable(t)

	id, err := table.Set("", &types.Crumb{Name: "Fidelity test"})
	require.NoError(t, err)

	entity, err := table.Get(id)
	require.NoError(t, err)
	crumb := entity.(*types.Crumb)

	assert.Equal(t, id, crumb.CrumbID)
	assert.Equal(t, "Fidelity test", crumb.Name)
	assert.Equal(t, types.StateDraft, crumb.State)
	assert.False(t, crumb.CreatedAt.IsZero(), "CreatedAt must be set")
	assert.False(t, crumb.UpdatedAt.IsZero(), "UpdatedAt must be set")
}
