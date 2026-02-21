package engine

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreateSchema_Tables(t *testing.T) {
	db := openTestDB(t)
	if err := CreateSchema(db); err != nil {
		t.Fatalf("CreateSchema: %v", err)
	}

	want := []string{
		"crumbs",
		"trails",
		"links",
		"properties",
		"categories",
		"crumb_properties",
		"metadata",
		"stashes",
		"stash_history",
	}

	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		t.Fatalf("query tables: %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, name)
	}

	if len(got) != len(want) {
		t.Fatalf("got %d tables, want %d: %v", len(got), len(want), got)
	}
	wantSet := make(map[string]bool)
	for _, w := range want {
		wantSet[w] = true
	}
	for _, g := range got {
		if !wantSet[g] {
			t.Errorf("unexpected table %q", g)
		}
	}
}

func TestCreateSchema_Indexes(t *testing.T) {
	db := openTestDB(t)
	if err := CreateSchema(db); err != nil {
		t.Fatalf("CreateSchema: %v", err)
	}

	wantIndexes := []string{
		"idx_crumbs_state",
		"idx_trails_state",
		"idx_links_type_from",
		"idx_links_type_to",
		"idx_crumb_properties_crumb",
		"idx_crumb_properties_property",
		"idx_metadata_crumb",
		"idx_metadata_table",
		"idx_categories_property",
		"idx_stashes_name",
		"idx_stash_history_stash",
		"idx_stash_history_version",
	}

	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%' ORDER BY name`)
	if err != nil {
		t.Fatalf("query indexes: %v", err)
	}
	defer rows.Close()

	got := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got[name] = true
	}

	for _, idx := range wantIndexes {
		if !got[idx] {
			t.Errorf("missing index %q", idx)
		}
	}
	if len(got) != len(wantIndexes) {
		t.Errorf("got %d indexes, want %d", len(got), len(wantIndexes))
	}
}

func TestCreateSchema_CrumbsColumns(t *testing.T) {
	db := openTestDB(t)
	if err := CreateSchema(db); err != nil {
		t.Fatalf("CreateSchema: %v", err)
	}

	// Insert a valid crumb row to verify column types.
	_, err := db.Exec(`INSERT INTO crumbs (crumb_id, name, state, created_at, updated_at)
		VALUES ('id-1', 'Test', 'pending', '2025-01-15T10:30:00Z', '2025-01-15T10:30:00Z')`)
	if err != nil {
		t.Fatalf("insert crumb: %v", err)
	}

	var id, name, state, createdAt, updatedAt string
	err = db.QueryRow(`SELECT crumb_id, name, state, created_at, updated_at FROM crumbs WHERE crumb_id='id-1'`).
		Scan(&id, &name, &state, &createdAt, &updatedAt)
	if err != nil {
		t.Fatalf("select crumb: %v", err)
	}
	if id != "id-1" || name != "Test" || state != "pending" {
		t.Errorf("got (%q, %q, %q), want (id-1, Test, pending)", id, name, state)
	}
}

func TestCreateSchema_ForeignKeys(t *testing.T) {
	db := openTestDB(t)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable FK: %v", err)
	}
	if err := CreateSchema(db); err != nil {
		t.Fatalf("CreateSchema: %v", err)
	}

	// categories.property_id should reference properties.property_id.
	_, err := db.Exec(`INSERT INTO categories (category_id, property_id, name, ordinal) VALUES ('c1', 'nonexistent', 'test', 0)`)
	if err == nil {
		t.Error("expected FK violation for categories.property_id, got nil")
	}
}
