// Tests for SQLite backend Cupboard lifecycle: Attach, Detach, GetTable.
// Validates: prd002-sqlite-backend (R4: Startup, R6: Shutdown, R11: Cupboard Interface, R12: Table Name Routing);
//
//	prd001-cupboard-core (R2: Cupboard Interface, R4: Attach, R5: Detach, R6: Error Handling After Detach);
//	test-rel01.0-uc001-cupboard-lifecycle; test-rel01.0-uc002-table-crud (S10: Detach tests).
package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mesh-intelligence/crumbs/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestCupboard creates a Backend attached to a temporary directory.
// It registers cleanup to Detach and remove the directory after the test.
func newTestCupboard(t *testing.T) *Backend {
	t.Helper()
	dir := t.TempDir()
	b := NewBackend()
	cfg := types.Config{
		Backend: types.BackendSQLite,
		DataDir: dir,
	}
	err := b.Attach(cfg)
	require.NoError(t, err, "Attach must succeed")
	t.Cleanup(func() {
		b.Detach()
	})
	return b
}

func TestAttachCreatesDirectoryAndFiles(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "data")
	b := NewBackend()
	cfg := types.Config{Backend: types.BackendSQLite, DataDir: dir}

	err := b.Attach(cfg)
	require.NoError(t, err)
	defer b.Detach()

	// DataDir must exist.
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// All JSONL files must exist.
	for _, name := range jsonlFiles {
		p := filepath.Join(dir, name)
		_, err := os.Stat(p)
		assert.NoError(t, err, "expected JSONL file %s to exist", name)
	}

	// cupboard.db must exist (SQLite cache).
	_, err = os.Stat(filepath.Join(dir, "cupboard.db"))
	assert.NoError(t, err, "expected cupboard.db to exist")
}

func TestAttachDoubleAttachReturnsError(t *testing.T) {
	b := newTestCupboard(t)
	cfg := types.Config{Backend: types.BackendSQLite, DataDir: t.TempDir()}

	err := b.Attach(cfg)
	assert.ErrorIs(t, err, types.ErrAlreadyAttached)
}

func TestAttachInvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config types.Config
		errIs  error
	}{
		{
			name:   "empty backend",
			config: types.Config{Backend: "", DataDir: "/tmp/x"},
			errIs:  types.ErrBackendEmpty,
		},
		{
			name:   "unknown backend",
			config: types.Config{Backend: "postgres", DataDir: "/tmp/x"},
			errIs:  types.ErrBackendUnknown,
		},
		{
			name:   "empty data dir",
			config: types.Config{Backend: types.BackendSQLite, DataDir: ""},
			errIs:  types.ErrDataDirEmpty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBackend()
			err := b.Attach(tt.config)
			assert.ErrorIs(t, err, tt.errIs)
		})
	}
}

func TestDetachIdempotent(t *testing.T) {
	b := newTestCupboard(t)

	err := b.Detach()
	assert.NoError(t, err)

	// Second detach is idempotent.
	err = b.Detach()
	assert.NoError(t, err)
}

func TestGetTableCrumbs(t *testing.T) {
	b := newTestCupboard(t)

	table, err := b.GetTable(types.TableCrumbs)
	require.NoError(t, err)
	assert.NotNil(t, table)
}

func TestGetTableUnknownReturnsError(t *testing.T) {
	b := newTestCupboard(t)

	_, err := b.GetTable("nonexistent")
	assert.ErrorIs(t, err, types.ErrTableNotFound)
}

func TestGetTableSameInstance(t *testing.T) {
	b := newTestCupboard(t)

	t1, err := b.GetTable(types.TableCrumbs)
	require.NoError(t, err)
	t2, err := b.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	// R12.4: GetTable returns the same accessor instance.
	assert.Same(t, t1, t2)
}

func TestGetTableAfterDetachReturnsError(t *testing.T) {
	b := newTestCupboard(t)
	err := b.Detach()
	require.NoError(t, err)

	_, err = b.GetTable(types.TableCrumbs)
	assert.ErrorIs(t, err, types.ErrCupboardDetached)
}

func TestPostDetachOperationsReturnError(t *testing.T) {
	b := newTestCupboard(t)

	// Get a table reference before detaching.
	table, err := b.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	err = b.Detach()
	require.NoError(t, err)

	// All Table operations must return ErrCupboardDetached.
	_, err = table.Get("some-id")
	assert.ErrorIs(t, err, types.ErrCupboardDetached, "Get after detach")

	_, err = table.Set("", &types.Crumb{Name: "test"})
	assert.ErrorIs(t, err, types.ErrCupboardDetached, "Set after detach")

	err = table.Delete("some-id")
	assert.ErrorIs(t, err, types.ErrCupboardDetached, "Delete after detach")

	_, err = table.Fetch(nil)
	assert.ErrorIs(t, err, types.ErrCupboardDetached, "Fetch after detach")
}

func TestAttachDeletesExistingDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "cupboard.db")

	// Create a stale DB file.
	err := os.WriteFile(dbPath, []byte("stale data"), 0o644)
	require.NoError(t, err)

	b := NewBackend()
	cfg := types.Config{Backend: types.BackendSQLite, DataDir: dir}
	err = b.Attach(cfg)
	require.NoError(t, err)
	defer b.Detach()

	// DB should be a fresh SQLite file, not the stale data.
	data, err := os.ReadFile(dbPath)
	require.NoError(t, err)
	assert.NotEqual(t, "stale data", string(data))
}
