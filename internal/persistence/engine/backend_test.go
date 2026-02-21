package engine

import (
	"testing"

	"github.com/petar-djukic/crumbs/pkg/api"
	"github.com/petar-djukic/crumbs/pkg/constants"
)

func TestNewBackend(t *testing.T) {
	b := NewBackend()
	if b == nil {
		t.Fatal("NewBackend returned nil")
	}
	if b.Attached() {
		t.Error("new backend should not be attached")
	}
	if b.DB() != nil {
		t.Error("new backend should have nil DB")
	}
	if b.tables == nil {
		t.Error("tables map should be initialized")
	}
}

func TestBackend_SyncConfig_Default(t *testing.T) {
	b := NewBackend()
	cfg := b.SyncConfig()
	if cfg.SyncStrategy != "" {
		t.Errorf("got strategy %q, want empty", cfg.SyncStrategy)
	}
}

func TestBackend_SyncConfig_WithConfig(t *testing.T) {
	b := NewBackend()
	b.config = api.Config{
		SQLiteConfig: &api.SQLiteConfig{
			SyncStrategy:  constants.SyncBatch,
			BatchSize:     200,
			BatchInterval: 15,
		},
	}
	cfg := b.SyncConfig()
	if cfg.SyncStrategy != constants.SyncBatch {
		t.Errorf("got strategy %q, want %q", cfg.SyncStrategy, constants.SyncBatch)
	}
	if cfg.BatchSize != 200 {
		t.Errorf("got batch size %d, want 200", cfg.BatchSize)
	}
}
