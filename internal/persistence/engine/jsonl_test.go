package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONL_ValidLines(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.jsonl")
	data := `{"id":"1","name":"first"}
{"id":"2","name":"second"}
`
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	records, warnings, err := ReadJSONL(p)
	if err != nil {
		t.Fatalf("ReadJSONL: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("got %d warnings, want 0: %v", len(warnings), warnings)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	if records[0]["id"] != "1" {
		t.Errorf("records[0][id] = %v, want 1", records[0]["id"])
	}
}

func TestReadJSONL_EmptyLines(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.jsonl")
	data := `{"id":"1"}

{"id":"2"}

`
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	records, warnings, err := ReadJSONL(p)
	if err != nil {
		t.Fatalf("ReadJSONL: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("got %d warnings, want 0", len(warnings))
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
}

func TestReadJSONL_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.jsonl")
	data := `{"id":"1"}
not json at all
{"id":"2"}
{invalid
`
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	records, warnings, err := ReadJSONL(p)
	if err != nil {
		t.Fatalf("ReadJSONL: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2 (skipping malformed)", len(records))
	}
	if len(warnings) != 2 {
		t.Fatalf("got %d warnings, want 2", len(warnings))
	}
}

func TestReadJSONL_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.jsonl")
	if err := os.WriteFile(p, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	records, warnings, err := ReadJSONL(p)
	if err != nil {
		t.Fatalf("ReadJSONL: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("got %d records from empty file", len(records))
	}
	if len(warnings) != 0 {
		t.Errorf("got %d warnings from empty file", len(warnings))
	}
}

func TestWriteJSONL_AtomicRename(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "out.jsonl")

	records := []map[string]any{
		{"id": "1", "name": "first"},
		{"id": "2", "name": "second"},
	}
	if err := WriteJSONL(p, records); err != nil {
		t.Fatalf("WriteJSONL: %v", err)
	}

	// Read it back and verify.
	got, warnings, err := ReadJSONL(p)
	if err != nil {
		t.Fatalf("ReadJSONL: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: %v", warnings)
	}
	if len(got) != 2 {
		t.Fatalf("got %d records, want 2", len(got))
	}
	if got[0]["id"] != "1" || got[1]["id"] != "2" {
		t.Errorf("roundtrip mismatch: %v", got)
	}
}

func TestWriteJSONL_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "out.jsonl")

	// Write initial data.
	if err := WriteJSONL(p, []map[string]any{{"id": "old"}}); err != nil {
		t.Fatal(err)
	}

	// Overwrite with new data.
	if err := WriteJSONL(p, []map[string]any{{"id": "new"}}); err != nil {
		t.Fatal(err)
	}

	got, _, err := ReadJSONL(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0]["id"] != "new" {
		t.Errorf("expected overwritten data, got %v", got)
	}
}

func TestWriteJSONL_EmptyRecords(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "out.jsonl")

	if err := WriteJSONL(p, nil); err != nil {
		t.Fatalf("WriteJSONL nil: %v", err)
	}

	got, _, err := ReadJSONL(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 records, got %d", len(got))
	}
}

func TestAppendJSONL(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "history.jsonl")

	if err := AppendJSONL(p, map[string]any{"version": float64(1)}); err != nil {
		t.Fatalf("first append: %v", err)
	}
	if err := AppendJSONL(p, map[string]any{"version": float64(2)}); err != nil {
		t.Fatalf("second append: %v", err)
	}

	records, _, err := ReadJSONL(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	if records[0]["version"] != float64(1) || records[1]["version"] != float64(2) {
		t.Errorf("append order: %v", records)
	}
}

func TestEnsureJSONLFiles(t *testing.T) {
	dir := t.TempDir()
	if err := EnsureJSONLFiles(dir); err != nil {
		t.Fatalf("EnsureJSONLFiles: %v", err)
	}

	for _, name := range JSONLFiles {
		p := filepath.Join(dir, name)
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("missing file %s: %v", name, err)
			continue
		}
		if info.Size() != 0 {
			t.Errorf("%s should be empty (0 bytes), got %d", name, info.Size())
		}
	}
}

func TestEnsureJSONLFiles_PreservesExisting(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "crumbs.jsonl")
	if err := os.WriteFile(p, []byte(`{"id":"keep"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureJSONLFiles(dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"id":"keep"}`+"\n" {
		t.Errorf("existing file was overwritten: %q", data)
	}
}

func TestReadJSONLTyped(t *testing.T) {
	type Record struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	dir := t.TempDir()
	p := filepath.Join(dir, "test.jsonl")
	data := `{"id":"1","name":"alpha"}
{"id":"2","name":"beta"}
`
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	records, warnings, err := ReadJSONLTyped[Record](p)
	if err != nil {
		t.Fatalf("ReadJSONLTyped: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: %v", warnings)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	if records[0].ID != "1" || records[1].Name != "beta" {
		t.Errorf("unexpected records: %+v", records)
	}
}

func TestWriteJSONLTyped(t *testing.T) {
	type Record struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	dir := t.TempDir()
	p := filepath.Join(dir, "out.jsonl")

	records := []Record{
		{ID: "1", Name: "alpha"},
		{ID: "2", Name: "beta"},
	}
	if err := WriteJSONLTyped(p, records); err != nil {
		t.Fatalf("WriteJSONLTyped: %v", err)
	}

	got, warnings, err := ReadJSONLTyped[Record](p)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: %v", warnings)
	}
	if len(got) != 2 || got[0].ID != "1" || got[1].Name != "beta" {
		t.Errorf("roundtrip mismatch: %+v", got)
	}
}
