// Unit tests for JSONL loading with forward compatibility.
// Validates: prd002-sqlite-backend R4 (startup loading), R7.2 (unknown field tolerance);
//            test-rel02.0-uc002-regeneration-compatibility (test cases 1-4).
package sqlite

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mesh-intelligence/crumbs/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadJSONLUnknownFields(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		jsonl    string
		table    string
		countSQL string
		wantRows int
		checkSQL string
		checkVal string
	}{
		{
			name: "crumbs with unknown fields load successfully",
			file: "crumbs.jsonl",
			jsonl: `{"crumb_id":"aaa-001","name":"Test crumb","state":"draft","created_at":"2025-01-15T10:30:00Z","updated_at":"2025-01-15T10:30:00Z","future_field":"unknown","priority_score":42}
`,
			table:    "crumbs",
			countSQL: "SELECT COUNT(*) FROM crumbs",
			wantRows: 1,
			checkSQL: "SELECT name FROM crumbs WHERE crumb_id = 'aaa-001'",
			checkVal: "Test crumb",
		},
		{
			name: "trails with unknown fields load successfully",
			file: "trails.jsonl",
			jsonl: `{"trail_id":"bbb-001","state":"active","created_at":"2025-01-15T10:30:00Z","completed_at":null,"branch_depth":3,"agent_version":"2.0"}
`,
			table:    "trails",
			countSQL: "SELECT COUNT(*) FROM trails",
			wantRows: 1,
			checkSQL: "SELECT state FROM trails WHERE trail_id = 'bbb-001'",
			checkVal: "active",
		},
		{
			name: "properties with unknown fields load successfully",
			file: "properties.jsonl",
			jsonl: `{"property_id":"ccc-001","name":"custom_prop","description":"A test property","value_type":"text","created_at":"2025-01-15T10:30:00Z","validation_regex":"^[a-z]+$"}
`,
			table:    "properties",
			countSQL: "SELECT COUNT(*) FROM properties",
			wantRows: 1,
			checkSQL: "SELECT value_type FROM properties WHERE property_id = 'ccc-001'",
			checkVal: "text",
		},
		{
			name: "links with unknown fields load successfully",
			file: "links.jsonl",
			jsonl: `{"link_id":"ddd-001","link_type":"belongs_to","from_id":"aaa-001","to_id":"bbb-001","created_at":"2025-01-15T10:30:00Z","weight":1.5}
`,
			table:    "links",
			countSQL: "SELECT COUNT(*) FROM links",
			wantRows: 1,
			checkSQL: "SELECT link_type FROM links WHERE link_id = 'ddd-001'",
			checkVal: "belongs_to",
		},
		// Note: metadata is tested in TestLoadJSONLMultipleEntityTypesWithUnknownFields
		// because it has a foreign key on crumb_id. In isolation (no crumb row), the
		// insert is silently skipped per R4.2. The multi-entity test provides the
		// referenced crumb.
		{
			name: "stashes with unknown fields load successfully",
			file: "stashes.jsonl",
			jsonl: `{"stash_id":"fff-001","name":"test_stash","stash_type":"resource","value":"{\"uri\":\"file:///tmp\"}","version":1,"created_at":"2025-01-15T10:30:00Z","updated_at":"2025-01-15T10:30:00Z","ttl_seconds":3600}
`,
			table:    "stashes",
			countSQL: "SELECT COUNT(*) FROM stashes",
			wantRows: 1,
			checkSQL: "SELECT name FROM stashes WHERE stash_id = 'fff-001'",
			checkVal: "test_stash",
		},
		{
			name: "multiple crumbs with varying unknown fields",
			file: "crumbs.jsonl",
			jsonl: `{"crumb_id":"aaa-001","name":"Crumb one","state":"draft","created_at":"2025-01-15T10:30:00Z","updated_at":"2025-01-15T10:30:00Z","future_a":"val"}
{"crumb_id":"aaa-002","name":"Crumb two","state":"ready","created_at":"2025-01-15T10:31:00Z","updated_at":"2025-01-15T10:31:00Z","future_b":true,"future_c":99}
{"crumb_id":"aaa-003","name":"Crumb three","state":"taken","created_at":"2025-01-15T10:32:00Z","updated_at":"2025-01-15T10:32:00Z"}
`,
			table:    "crumbs",
			countSQL: "SELECT COUNT(*) FROM crumbs",
			wantRows: 3,
			checkSQL: "SELECT name FROM crumbs WHERE crumb_id = 'aaa-002'",
			checkVal: "Crumb two",
		},
		{
			name: "crumbs with nested unknown object fields",
			file: "crumbs.jsonl",
			jsonl: `{"crumb_id":"aaa-001","name":"Nested test","state":"draft","created_at":"2025-01-15T10:30:00Z","updated_at":"2025-01-15T10:30:00Z","complex_field":{"nested":"value","deep":{"level":2}}}
`,
			table:    "crumbs",
			countSQL: "SELECT COUNT(*) FROM crumbs",
			wantRows: 1,
			checkSQL: "SELECT name FROM crumbs WHERE crumb_id = 'aaa-001'",
			checkVal: "Nested test",
		},
		{
			name: "crumbs with unknown array fields",
			file: "crumbs.jsonl",
			jsonl: `{"crumb_id":"aaa-001","name":"Array test","state":"draft","created_at":"2025-01-15T10:30:00Z","updated_at":"2025-01-15T10:30:00Z","tags_v2":["alpha","beta"],"scores":[1,2,3]}
`,
			table:    "crumbs",
			countSQL: "SELECT COUNT(*) FROM crumbs",
			wantRows: 1,
			checkSQL: "SELECT name FROM crumbs WHERE crumb_id = 'aaa-001'",
			checkVal: "Array test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, dataDir := setupTestDB(t)

			// Write the test JSONL content to the file.
			err := os.WriteFile(filepath.Join(dataDir, tt.file), []byte(tt.jsonl), 0o644)
			require.NoError(t, err)

			// Load all JSONL into SQLite.
			err = loadAllJSONL(db, dataDir)
			require.NoError(t, err, "loadAllJSONL must not error on unknown fields")

			// Verify the expected row count.
			var count int
			err = db.QueryRow(tt.countSQL).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, tt.wantRows, count)

			// Verify a known field was loaded correctly.
			var val string
			err = db.QueryRow(tt.checkSQL).Scan(&val)
			require.NoError(t, err)
			assert.Equal(t, tt.checkVal, val)
		})
	}
}

func TestLoadJSONLMixedKnownAndUnknownFields(t *testing.T) {
	t.Run("known fields preserved when unknown fields present", func(t *testing.T) {
		db, dataDir := setupTestDB(t)

		// Write a crumb with many unknown fields alongside all known fields.
		jsonl := `{"crumb_id":"mix-001","name":"Mixed fields","state":"pending","created_at":"2025-06-01T09:00:00Z","updated_at":"2025-06-01T09:15:00Z","future_x":"hello","future_y":42,"future_z":{"nested":true}}` + "\n"
		err := os.WriteFile(filepath.Join(dataDir, "crumbs.jsonl"), []byte(jsonl), 0o644)
		require.NoError(t, err)

		err = loadAllJSONL(db, dataDir)
		require.NoError(t, err)

		// Verify all known fields are correct.
		var id, name, state, createdAt, updatedAt string
		err = db.QueryRow("SELECT crumb_id, name, state, created_at, updated_at FROM crumbs").Scan(
			&id, &name, &state, &createdAt, &updatedAt,
		)
		require.NoError(t, err)
		assert.Equal(t, "mix-001", id)
		assert.Equal(t, "Mixed fields", name)
		assert.Equal(t, "pending", state)
		assert.Equal(t, "2025-06-01T09:00:00Z", createdAt)
		assert.Equal(t, "2025-06-01T09:15:00Z", updatedAt)
	})
}

func TestLoadJSONLMalformedAndUnknownFieldsTogether(t *testing.T) {
	t.Run("malformed lines skipped alongside records with unknown fields", func(t *testing.T) {
		db, dataDir := setupTestDB(t)

		// Mix of: valid with unknown fields, malformed, valid without unknown fields.
		jsonl := `{"crumb_id":"ok-001","name":"Valid with extra","state":"draft","created_at":"2025-01-15T10:30:00Z","updated_at":"2025-01-15T10:30:00Z","future_field":"v1"}
not valid json at all
{"crumb_id":"ok-002","name":"Valid without extra","state":"ready","created_at":"2025-01-15T10:31:00Z","updated_at":"2025-01-15T10:31:00Z"}
`
		err := os.WriteFile(filepath.Join(dataDir, "crumbs.jsonl"), []byte(jsonl), 0o644)
		require.NoError(t, err)

		err = loadAllJSONL(db, dataDir)
		require.NoError(t, err)

		// Only valid records should be loaded (malformed skipped per R4.2).
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM crumbs").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}

func TestBackendAttachWithUnknownFieldsInJSONL(t *testing.T) {
	t.Run("backend attach succeeds with unknown fields in JSONL", func(t *testing.T) {
		dataDir := t.TempDir()

		// Pre-populate JSONL files with records containing unknown fields,
		// simulating data from a future generation (N+1).
		crumbsJSONL := `{"crumb_id":"gen-001","name":"Future crumb","state":"draft","created_at":"2025-01-15T10:30:00Z","updated_at":"2025-01-15T10:30:00Z","future_field":"from_gen_n_plus_1","priority_v2":{"level":"high","score":90}}` + "\n"
		trailsJSONL := `{"trail_id":"trail-001","state":"active","created_at":"2025-01-15T10:30:00Z","completed_at":null,"duration_estimate":"2h"}` + "\n"

		for _, name := range jsonlFiles {
			content := []byte{}
			switch name {
			case "crumbs.jsonl":
				content = []byte(crumbsJSONL)
			case "trails.jsonl":
				content = []byte(trailsJSONL)
			}
			err := os.WriteFile(filepath.Join(dataDir, name), content, 0o644)
			require.NoError(t, err)
		}

		b := NewBackend()
		config := types.Config{
			Backend: "sqlite",
			DataDir: dataDir,
		}

		err := b.Attach(config)
		require.NoError(t, err, "Attach must succeed with unknown fields in JSONL")
		defer b.Detach()

		// Verify crumb is accessible via the Table interface.
		table, err := b.GetTable(types.TableCrumbs)
		require.NoError(t, err)

		entity, err := table.Get("gen-001")
		require.NoError(t, err)

		crumb, ok := entity.(*types.Crumb)
		require.True(t, ok)
		assert.Equal(t, "gen-001", crumb.CrumbID)
		assert.Equal(t, "Future crumb", crumb.Name)
		assert.Equal(t, "draft", crumb.State)

		// Verify trail is accessible.
		trailTable, err := b.GetTable(types.TableTrails)
		require.NoError(t, err)

		trailEntity, err := trailTable.Get("trail-001")
		require.NoError(t, err)

		trail, ok := trailEntity.(*types.Trail)
		require.True(t, ok)
		assert.Equal(t, "trail-001", trail.TrailID)
		assert.Equal(t, "active", trail.State)
	})
}

func TestBackendRoundTripWithUnknownFields(t *testing.T) {
	t.Run("entities loaded from JSONL with unknown fields can be read and written back", func(t *testing.T) {
		dataDir := t.TempDir()

		// Write JSONL with unknown fields.
		crumbsJSONL := `{"crumb_id":"rt-001","name":"Round trip test","state":"draft","created_at":"2025-01-15T10:30:00Z","updated_at":"2025-01-15T10:30:00Z","future_field":"preserved?"}` + "\n"
		for _, name := range jsonlFiles {
			content := []byte{}
			if name == "crumbs.jsonl" {
				content = []byte(crumbsJSONL)
			}
			err := os.WriteFile(filepath.Join(dataDir, name), content, 0o644)
			require.NoError(t, err)
		}

		// Attach, modify the crumb, and persist.
		b := NewBackend()
		config := types.Config{Backend: "sqlite", DataDir: dataDir}

		err := b.Attach(config)
		require.NoError(t, err)

		table, err := b.GetTable(types.TableCrumbs)
		require.NoError(t, err)

		// Retrieve and modify.
		entity, err := table.Get("rt-001")
		require.NoError(t, err)
		crumb := entity.(*types.Crumb)
		assert.Equal(t, "Round trip test", crumb.Name)

		crumb.Name = "Updated name"
		_, err = table.Set(crumb.CrumbID, crumb)
		require.NoError(t, err)

		b.Detach()

		// Re-attach and verify the modification persisted.
		b2 := NewBackend()
		err = b2.Attach(config)
		require.NoError(t, err)
		defer b2.Detach()

		table2, err := b2.GetTable(types.TableCrumbs)
		require.NoError(t, err)

		entity2, err := table2.Get("rt-001")
		require.NoError(t, err)
		crumb2 := entity2.(*types.Crumb)
		assert.Equal(t, "Updated name", crumb2.Name)
		assert.Equal(t, "draft", crumb2.State)
	})
}

func TestLoadJSONLEmptyAndMissingFiles(t *testing.T) {
	t.Run("empty JSONL files load without error", func(t *testing.T) {
		db, dataDir := setupTestDB(t)

		err := loadAllJSONL(db, dataDir)
		require.NoError(t, err)

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM crumbs").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestLoadJSONLMissingKnownFields(t *testing.T) {
	t.Run("records with missing optional fields and unknown extra fields load", func(t *testing.T) {
		db, dataDir := setupTestDB(t)

		// Trail with null completed_at and unknown fields.
		jsonl := `{"trail_id":"trail-001","state":"active","created_at":"2025-01-15T10:30:00Z","completed_at":null,"future_metrics":{"latency_ms":42}}` + "\n"
		err := os.WriteFile(filepath.Join(dataDir, "trails.jsonl"), []byte(jsonl), 0o644)
		require.NoError(t, err)

		err = loadAllJSONL(db, dataDir)
		require.NoError(t, err)

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM trails").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestLoadJSONLMultipleEntityTypesWithUnknownFields(t *testing.T) {
	t.Run("all entity types tolerate unknown fields simultaneously", func(t *testing.T) {
		db, dataDir := setupTestDB(t)

		files := map[string]string{
			"crumbs.jsonl":     `{"crumb_id":"c-001","name":"C1","state":"draft","created_at":"2025-01-15T10:30:00Z","updated_at":"2025-01-15T10:30:00Z","x":"1"}` + "\n",
			"trails.jsonl":     `{"trail_id":"t-001","state":"active","created_at":"2025-01-15T10:30:00Z","completed_at":null,"x":"2"}` + "\n",
			"properties.jsonl": `{"property_id":"p-001","name":"prop1","description":"d","value_type":"text","created_at":"2025-01-15T10:30:00Z","x":"3"}` + "\n",
			"links.jsonl":      `{"link_id":"l-001","link_type":"belongs_to","from_id":"c-001","to_id":"t-001","created_at":"2025-01-15T10:30:00Z","x":"4"}` + "\n",
			"metadata.jsonl":   `{"metadata_id":"m-001","table_name":"comments","crumb_id":"c-001","property_id":null,"content":"hello","created_at":"2025-01-15T10:30:00Z","x":"5"}` + "\n",
			"stashes.jsonl":    `{"stash_id":"s-001","name":"stash1","stash_type":"resource","value":"{}","version":1,"created_at":"2025-01-15T10:30:00Z","updated_at":"2025-01-15T10:30:00Z","x":"6"}` + "\n",
		}

		for name, content := range files {
			err := os.WriteFile(filepath.Join(dataDir, name), []byte(content), 0o644)
			require.NoError(t, err)
		}

		err := loadAllJSONL(db, dataDir)
		require.NoError(t, err, "loading all entity types with unknown fields must succeed")

		// Verify each table has one row.
		tables := []struct {
			query string
			label string
		}{
			{"SELECT COUNT(*) FROM crumbs", "crumbs"},
			{"SELECT COUNT(*) FROM trails", "trails"},
			{"SELECT COUNT(*) FROM properties", "properties"},
			{"SELECT COUNT(*) FROM links", "links"},
			{"SELECT COUNT(*) FROM metadata", "metadata"},
			{"SELECT COUNT(*) FROM stashes", "stashes"},
		}

		for _, tc := range tables {
			var count int
			err := db.QueryRow(tc.query).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count, "%s should have 1 row", tc.label)
		}
	})
}

func TestInsertRecordsIgnoresUnknownFields(t *testing.T) {
	t.Run("insertRecords extracts only known columns from records", func(t *testing.T) {
		db, dataDir := setupTestDB(t)
		_ = dataDir // not needed for this test

		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		records := []json.RawMessage{
			json.RawMessage(`{"crumb_id":"ins-001","name":"Insert test","state":"draft","created_at":"2025-01-15T10:30:00Z","updated_at":"2025-01-15T10:30:00Z","unknown_a":"val","unknown_b":123}`),
		}

		columns := []string{"crumb_id", "name", "state", "created_at", "updated_at"}
		err = insertRecords(tx, "crumbs", columns, records)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		var name string
		err = db.QueryRow("SELECT name FROM crumbs WHERE crumb_id = 'ins-001'").Scan(&name)
		require.NoError(t, err)
		assert.Equal(t, "Insert test", name)
	})
}
