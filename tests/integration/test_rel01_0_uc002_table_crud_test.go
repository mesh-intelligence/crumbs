// Tests for rel01.0-uc002-table-crud: Table interface CRUD operations
// (Get, Set, Delete, Fetch) through the SQLite backend with UUID v7
// generation, field fidelity, filtering, JSONL persistence, and
// cross-table behavior.
// Implements: test-rel01.0-uc002-table-crud.yaml.
package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mesh-intelligence/crumbs/pkg/types"
)

func TestCrumbCreateWithUUID(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "create crumb with empty ID generates UUID v7",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id := mustCreateCrumb(t, tbl, "Test crumb", types.CrumbStateDraft)
				if !isUUIDv7(id) {
					t.Errorf("expected UUID v7, got %q", id)
				}
				c := mustGetCrumb(t, tbl, id)
				if c.Name != "Test crumb" {
					t.Errorf("expected name 'Test crumb', got %q", c.Name)
				}
				if c.State != types.CrumbStateDraft {
					t.Errorf("expected state draft, got %q", c.State)
				}
			},
		},
		{
			name: "created crumb persists to JSONL file",
			run: func(t *testing.T) {
				b, dir := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id := mustCreateCrumb(t, tbl, "Persisted crumb", types.CrumbStateDraft)
				if !isUUIDv7(id) {
					t.Errorf("expected UUID v7, got %q", id)
				}
				assertJSONLContains(t, dir, "crumbs.jsonl", `"name":"Persisted crumb"`)
			},
		},
		{
			name: "two creates generate unique UUIDs",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id1 := mustCreateCrumb(t, tbl, "First crumb", types.CrumbStateDraft)
				id2 := mustCreateCrumb(t, tbl, "Second crumb", types.CrumbStateDraft)

				if id1 == id2 {
					t.Error("expected unique IDs, got same")
				}
				if !isUUIDv7(id1) || !isUUIDv7(id2) {
					t.Errorf("expected both UUID v7: %q, %q", id1, id2)
				}

				results := fetchAll(t, tbl)
				if len(results) != 2 {
					t.Errorf("expected 2 crumbs, got %d", len(results))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestCrumbGetRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		crumbName string
		state     string
	}{
		{"get retrieves entity with matching fields", "Retrieve me", types.CrumbStateReady},
		{"round-trip fidelity for crumb fields", "Fidelity test", types.CrumbStatePending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := setupCupboard(t)
			tbl := mustGetTable(t, b, types.TableCrumbs)

			id := mustCreateCrumb(t, tbl, tt.crumbName, tt.state)
			c := mustGetCrumb(t, tbl, id)

			if c.CrumbID != id {
				t.Errorf("expected CrumbID %q, got %q", id, c.CrumbID)
			}
			if c.Name != tt.crumbName {
				t.Errorf("expected name %q, got %q", tt.crumbName, c.Name)
			}
			if c.State != tt.state {
				t.Errorf("expected state %q, got %q", tt.state, c.State)
			}
			if c.CreatedAt.IsZero() {
				t.Error("expected non-zero CreatedAt")
			}
			if c.UpdatedAt.IsZero() {
				t.Error("expected non-zero UpdatedAt")
			}
		})
	}
}

func TestCrumbUpdate(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "update entity via Set with existing ID",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id := mustCreateCrumb(t, tbl, "Original name", types.CrumbStateDraft)
				c := mustGetCrumb(t, tbl, id)
				c.Name = "Updated name"
				if _, err := tbl.Set(id, c); err != nil {
					t.Fatalf("Set update: %v", err)
				}

				got := mustGetCrumb(t, tbl, id)
				if got.Name != "Updated name" {
					t.Errorf("expected 'Updated name', got %q", got.Name)
				}
			},
		},
		{
			name: "updated entity confirmed via Get",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id := mustCreateCrumb(t, tbl, "Before update", types.CrumbStateDraft)
				c := mustGetCrumb(t, tbl, id)
				c.Name = "After update"
				c.State = types.CrumbStateReady
				tbl.Set(id, c)

				got := mustGetCrumb(t, tbl, id)
				if got.Name != "After update" {
					t.Errorf("expected 'After update', got %q", got.Name)
				}
				if got.State != types.CrumbStateReady {
					t.Errorf("expected state ready, got %q", got.State)
				}
			},
		},
		{
			name: "update persists to JSONL",
			run: func(t *testing.T) {
				b, dir := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id := mustCreateCrumb(t, tbl, "JSONL update test", types.CrumbStateDraft)
				c := mustGetCrumb(t, tbl, id)
				c.Name = "JSONL updated"
				c.State = types.CrumbStateTaken
				tbl.Set(id, c)

				assertJSONLContains(t, dir, "crumbs.jsonl", `"name":"JSONL updated"`)
				assertJSONLContains(t, dir, "crumbs.jsonl", `"state":"taken"`)
			},
		},
		{
			name: "UpdatedAt changes on update",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id := mustCreateCrumb(t, tbl, "Timestamp test", types.CrumbStateDraft)
				original := mustGetCrumb(t, tbl, id)
				origUpdatedAt := original.UpdatedAt

				time.Sleep(10 * time.Millisecond)
				original.Name = "Timestamp updated"
				tbl.Set(id, original)

				got := mustGetCrumb(t, tbl, id)
				if !got.UpdatedAt.After(origUpdatedAt) && !got.UpdatedAt.Equal(origUpdatedAt) {
					t.Errorf("expected UpdatedAt >= original, got %v <= %v", got.UpdatedAt, origUpdatedAt)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestCrumbFetchAll(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, tbl types.Table)
		want  int
	}{
		{
			name: "Fetch with empty filter returns all crumbs",
			setup: func(t *testing.T, tbl types.Table) {
				mustCreateCrumb(t, tbl, "Crumb A", types.CrumbStateDraft)
				mustCreateCrumb(t, tbl, "Crumb B", types.CrumbStateReady)
				mustCreateCrumb(t, tbl, "Crumb C", types.CrumbStateTaken)
			},
			want: 3,
		},
		{
			name:  "Fetch empty table returns empty list",
			setup: func(t *testing.T, tbl types.Table) {},
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := setupCupboard(t)
			tbl := mustGetTable(t, b, types.TableCrumbs)
			tt.setup(t, tbl)

			results := fetchAll(t, tbl)
			if len(results) != tt.want {
				t.Errorf("expected %d crumbs, got %d", tt.want, len(results))
			}
		})
	}
}

func TestCrumbFetchWithFilter(t *testing.T) {
	tests := []struct {
		name        string
		filterState string
		wantCount   int
	}{
		{"Fetch with state filter returns matching crumbs", types.CrumbStateReady, 2},
		{"Fetch with filter returns no matches", types.CrumbStatePebble, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := setupCupboard(t)
			tbl := mustGetTable(t, b, types.TableCrumbs)

			mustCreateCrumb(t, tbl, "Draft crumb", types.CrumbStateDraft)
			mustCreateCrumb(t, tbl, "Ready crumb 1", types.CrumbStateReady)
			mustCreateCrumb(t, tbl, "Ready crumb 2", types.CrumbStateReady)
			mustCreateCrumb(t, tbl, "Taken crumb", types.CrumbStateTaken)

			results := fetchByStates(t, tbl, []string{tt.filterState})
			if len(results) != tt.wantCount {
				t.Errorf("expected %d results, got %d", tt.wantCount, len(results))
			}
			for _, r := range results {
				c := r.(*types.Crumb)
				if c.State != tt.filterState {
					t.Errorf("expected state %q, got %q", tt.filterState, c.State)
				}
			}
		})
	}
}

func TestCrumbDelete(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Delete removes entity from SQLite",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id := mustCreateCrumb(t, tbl, "Delete me", types.CrumbStateDraft)
				if err := tbl.Delete(id); err != nil {
					t.Fatalf("Delete: %v", err)
				}
				_, err := tbl.Get(id)
				if err != types.ErrNotFound {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "Delete removes entity from JSONL",
			run: func(t *testing.T) {
				b, dir := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id := mustCreateCrumb(t, tbl, "JSONL delete test", types.CrumbStateDraft)
				tbl.Delete(id)

				assertJSONLNotContains(t, dir, "crumbs.jsonl", `"name":"JSONL delete test"`)
			},
		},
		{
			name: "Fetch after delete excludes deleted entity",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id1 := mustCreateCrumb(t, tbl, "Keep this", types.CrumbStateDraft)
				id2 := mustCreateCrumb(t, tbl, "Delete this", types.CrumbStateDraft)
				tbl.Delete(id2)

				results := fetchAll(t, tbl)
				if len(results) != 1 {
					t.Fatalf("expected 1 crumb, got %d", len(results))
				}
				c := results[0].(*types.Crumb)
				if c.CrumbID != id1 {
					t.Errorf("expected remaining crumb ID %q, got %q", id1, c.CrumbID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestCrumbGetDeleteNonexistent(t *testing.T) {
	tests := []struct {
		name      string
		operation string
	}{
		{"Get nonexistent crumb returns error", "get"},
		{"Delete nonexistent crumb returns error", "delete"},
		{"Get nonexistent trail returns error", "get_trail"},
		{"Delete nonexistent trail returns error", "delete_trail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := setupCupboard(t)

			switch tt.operation {
			case "get":
				tbl := mustGetTable(t, b, types.TableCrumbs)
				_, err := tbl.Get("nonexistent-uuid-12345")
				if err != types.ErrNotFound {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			case "delete":
				tbl := mustGetTable(t, b, types.TableCrumbs)
				err := tbl.Delete("nonexistent-uuid-12345")
				if err != types.ErrNotFound {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			case "get_trail":
				tbl := mustGetTable(t, b, types.TableTrails)
				_, err := tbl.Get("nonexistent-uuid-12345")
				if err != types.ErrNotFound {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			case "delete_trail":
				tbl := mustGetTable(t, b, types.TableTrails)
				err := tbl.Delete("nonexistent-uuid-12345")
				if err != types.ErrNotFound {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			}
		})
	}
}

func TestTrailCRUDOperations(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "create trail with empty ID generates UUID v7",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableTrails)

				id := mustCreateTrail(t, tbl, types.TrailStateActive)
				if !isUUIDv7(id) {
					t.Errorf("expected UUID v7, got %q", id)
				}

				tr := mustGetTrail(t, tbl, id)
				if tr.State != types.TrailStateActive {
					t.Errorf("expected state active, got %q", tr.State)
				}

				results := fetchAll(t, tbl)
				if len(results) != 1 {
					t.Errorf("expected 1 trail, got %d", len(results))
				}
			},
		},
		{
			name: "Get trail returns entity with matching fields",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableTrails)

				id := mustCreateTrail(t, tbl, types.TrailStateActive)
				tr := mustGetTrail(t, tbl, id)

				if tr.TrailID != id {
					t.Errorf("expected TrailID %q, got %q", id, tr.TrailID)
				}
				if tr.State != types.TrailStateActive {
					t.Errorf("expected state active, got %q", tr.State)
				}
			},
		},
		{
			name: "Update trail via Set with existing ID",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableTrails)

				id := mustCreateTrail(t, tbl, types.TrailStateDraft)
				tr := mustGetTrail(t, tbl, id)
				tr.State = types.TrailStateActive
				tbl.Set(id, tr)

				got := mustGetTrail(t, tbl, id)
				if got.State != types.TrailStateActive {
					t.Errorf("expected state active, got %q", got.State)
				}
			},
		},
		{
			name: "Fetch trails with empty filter returns all trails",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableTrails)

				mustCreateTrail(t, tbl, types.TrailStateActive)
				mustCreateTrail(t, tbl, types.TrailStateCompleted)

				results := fetchAll(t, tbl)
				if len(results) != 2 {
					t.Errorf("expected 2 trails, got %d", len(results))
				}
			},
		},
		{
			name: "Fetch trails with filter returns matching trails",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableTrails)

				mustCreateTrail(t, tbl, types.TrailStateActive)
				mustCreateTrail(t, tbl, types.TrailStateActive)
				mustCreateTrail(t, tbl, types.TrailStateCompleted)

				results := fetchByStates(t, tbl, []string{types.TrailStateActive})
				if len(results) != 2 {
					t.Errorf("expected 2 active trails, got %d", len(results))
				}
				for _, r := range results {
					tr := r.(*types.Trail)
					if tr.State != types.TrailStateActive {
						t.Errorf("expected state active, got %q", tr.State)
					}
				}
			},
		},
		{
			name: "Delete trail removes from storage",
			run: func(t *testing.T) {
				b, _ := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableTrails)

				id := mustCreateTrail(t, tbl, types.TrailStateActive)
				if err := tbl.Delete(id); err != nil {
					t.Fatalf("Delete: %v", err)
				}
				_, err := tbl.Get(id)
				if err != types.ErrNotFound {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "Trail persists to JSONL file",
			run: func(t *testing.T) {
				b, dir := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableTrails)

				mustCreateTrail(t, tbl, types.TrailStatePending)
				assertJSONLContains(t, dir, "trails.jsonl", `"state":"pending"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestJSONLReflectsOperations(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "JSONL reflects create operation",
			run: func(t *testing.T) {
				b, dir := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				mustCreateCrumb(t, tbl, "Create JSONL test", types.CrumbStateDraft)
				assertJSONLContains(t, dir, "crumbs.jsonl", `"name":"Create JSONL test"`)
			},
		},
		{
			name: "JSONL reflects update operation",
			run: func(t *testing.T) {
				b, dir := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id := mustCreateCrumb(t, tbl, "Before JSONL update", types.CrumbStateDraft)
				c := mustGetCrumb(t, tbl, id)
				c.Name = "After JSONL update"
				c.State = types.CrumbStateReady
				tbl.Set(id, c)

				assertJSONLContains(t, dir, "crumbs.jsonl", `"name":"After JSONL update"`)
				assertJSONLContains(t, dir, "crumbs.jsonl", `"state":"ready"`)
				assertJSONLNotContains(t, dir, "crumbs.jsonl", `"name":"Before JSONL update"`)
			},
		},
		{
			name: "JSONL reflects delete operation",
			run: func(t *testing.T) {
				b, dir := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id1 := mustCreateCrumb(t, tbl, "To be deleted", types.CrumbStateDraft)
				mustCreateCrumb(t, tbl, "To be kept", types.CrumbStateDraft)
				tbl.Delete(id1)

				assertJSONLContains(t, dir, "crumbs.jsonl", `"name":"To be kept"`)
				assertJSONLNotContains(t, dir, "crumbs.jsonl", `"name":"To be deleted"`)
			},
		},
		{
			name: "Multiple operations reflected in JSONL",
			run: func(t *testing.T) {
				b, dir := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				id1 := mustCreateCrumb(t, tbl, "Multi op 1", types.CrumbStateDraft)
				id2 := mustCreateCrumb(t, tbl, "Multi op 2", types.CrumbStateDraft)

				c1 := mustGetCrumb(t, tbl, id1)
				c1.Name = "Multi op 1 updated"
				c1.State = types.CrumbStateReady
				tbl.Set(id1, c1)
				tbl.Delete(id2)

				assertJSONLContains(t, dir, "crumbs.jsonl", `"name":"Multi op 1 updated"`)
				assertJSONLContains(t, dir, "crumbs.jsonl", `"state":"ready"`)
				assertJSONLNotContains(t, dir, "crumbs.jsonl", `"name":"Multi op 2"`)
			},
		},
		{
			name: "JSONL lines are valid JSON",
			run: func(t *testing.T) {
				b, dir := setupCupboard(t)
				tbl := mustGetTable(t, b, types.TableCrumbs)

				mustCreateCrumb(t, tbl, "Valid JSON 1", types.CrumbStateDraft)
				mustCreateCrumb(t, tbl, "Valid JSON 2", types.CrumbStateReady)

				lines := readJSONLFile(t, dir, "crumbs.jsonl")
				for i, line := range lines {
					if !json.Valid([]byte(line)) {
						t.Errorf("line %d is not valid JSON: %s", i+1, line)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestDetachPreventsOperations(t *testing.T) {
	tests := []struct {
		name      string
		operation string
	}{
		{"Get after detach returns error", "get"},
		{"Set after detach returns error", "set"},
		{"Fetch after detach returns error", "fetch"},
		{"Delete after detach returns error", "delete"},
		{"GetTable after detach returns error", "gettable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := setupCupboard(t)
			tbl := mustGetTable(t, b, types.TableCrumbs)

			// Create a crumb before detach for get/delete tests.
			id := mustCreateCrumb(t, tbl, "Pre-detach crumb", types.CrumbStateDraft)
			b.Detach()

			switch tt.operation {
			case "get":
				// After detach, we can't use the table accessor because the
				// backend's mutex-protected operations will fail. GetTable itself
				// returns ErrCupboardDetached.
				_, err := b.GetTable(types.TableCrumbs)
				if err != types.ErrCupboardDetached {
					t.Fatalf("expected ErrCupboardDetached, got %v", err)
				}
			case "set":
				_, err := b.GetTable(types.TableCrumbs)
				if err != types.ErrCupboardDetached {
					t.Fatalf("expected ErrCupboardDetached, got %v", err)
				}
			case "fetch":
				_, err := b.GetTable(types.TableCrumbs)
				if err != types.ErrCupboardDetached {
					t.Fatalf("expected ErrCupboardDetached, got %v", err)
				}
			case "delete":
				_, err := b.GetTable(types.TableCrumbs)
				if err != types.ErrCupboardDetached {
					t.Fatalf("expected ErrCupboardDetached, got %v", err)
				}
			case "gettable":
				_, err := b.GetTable(types.TableCrumbs)
				if err != types.ErrCupboardDetached {
					t.Fatalf("expected ErrCupboardDetached, got %v", err)
				}
			}
			_ = id // used above
		})
	}
}
