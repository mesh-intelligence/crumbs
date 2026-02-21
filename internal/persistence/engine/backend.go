package engine

import (
	"database/sql"
	"sync"

	"github.com/petar-djukic/crumbs/pkg/api"

	// Register the pure-Go SQLite driver.
	_ "modernc.org/sqlite"
)

// Backend is the SQLite storage engine for Crumbs.
// It holds an open database connection, sync configuration,
// and per-table accessors. All public methods are safe for
// concurrent use via the embedded RWMutex.
//
// Implements: prd002-sqlite-backend R3, R8, R11
type Backend struct {
	mu       sync.RWMutex
	attached bool
	config   api.Config
	db       *sql.DB
	tables   map[string]api.Table
}

// NewBackend returns a Backend ready for Attach.
// No database connection is opened until Attach is called.
func NewBackend() *Backend {
	return &Backend{
		tables: make(map[string]api.Table),
	}
}

// DB returns the underlying *sql.DB. Returns nil when detached.
func (b *Backend) DB() *sql.DB {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.db
}

// Attached reports whether the backend is currently attached.
func (b *Backend) Attached() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.attached
}

// Config returns the configuration used at attach time.
func (b *Backend) Config() api.Config {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.config
}

// SyncConfig returns the resolved SQLiteConfig, providing defaults
// when the caller did not supply one.
func (b *Backend) SyncConfig() api.SQLiteConfig {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.config.SQLiteConfig != nil {
		return *b.config.SQLiteConfig
	}
	return api.SQLiteConfig{}
}
