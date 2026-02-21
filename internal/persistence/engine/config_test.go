package engine

import (
	"testing"

	"github.com/petar-djukic/crumbs/pkg/api"
	"github.com/petar-djukic/crumbs/pkg/constants"
)

func TestResolveSyncConfig_Nil(t *testing.T) {
	cfg := ResolveSyncConfig(nil)
	if cfg.SyncStrategy != constants.SyncImmediate {
		t.Errorf("got strategy %q, want %q", cfg.SyncStrategy, constants.SyncImmediate)
	}
	if cfg.BatchSize != DefaultBatchSize {
		t.Errorf("got batch size %d, want %d", cfg.BatchSize, DefaultBatchSize)
	}
	if cfg.BatchInterval != DefaultBatchInterval {
		t.Errorf("got batch interval %d, want %d", cfg.BatchInterval, DefaultBatchInterval)
	}
}

func TestResolveSyncConfig_EmptyStrategy(t *testing.T) {
	cfg := ResolveSyncConfig(&api.SQLiteConfig{})
	if cfg.SyncStrategy != constants.SyncImmediate {
		t.Errorf("got strategy %q, want %q", cfg.SyncStrategy, constants.SyncImmediate)
	}
}

func TestResolveSyncConfig_BatchDefaults(t *testing.T) {
	cfg := ResolveSyncConfig(&api.SQLiteConfig{SyncStrategy: constants.SyncBatch})
	if cfg.BatchSize != DefaultBatchSize {
		t.Errorf("got batch size %d, want %d", cfg.BatchSize, DefaultBatchSize)
	}
	if cfg.BatchInterval != DefaultBatchInterval {
		t.Errorf("got batch interval %d, want %d", cfg.BatchInterval, DefaultBatchInterval)
	}
}

func TestResolveSyncConfig_BatchCustom(t *testing.T) {
	cfg := ResolveSyncConfig(&api.SQLiteConfig{
		SyncStrategy:  constants.SyncBatch,
		BatchSize:     50,
		BatchInterval: 10,
	})
	if cfg.BatchSize != 50 {
		t.Errorf("got batch size %d, want 50", cfg.BatchSize)
	}
	if cfg.BatchInterval != 10 {
		t.Errorf("got batch interval %d, want 10", cfg.BatchInterval)
	}
}

func TestResolveSyncConfig_OnClose(t *testing.T) {
	cfg := ResolveSyncConfig(&api.SQLiteConfig{SyncStrategy: constants.SyncOnClose})
	if cfg.SyncStrategy != constants.SyncOnClose {
		t.Errorf("got strategy %q, want %q", cfg.SyncStrategy, constants.SyncOnClose)
	}
}

func TestIsImmediate(t *testing.T) {
	tests := []struct {
		strategy string
		want     bool
	}{
		{"", true},
		{constants.SyncImmediate, true},
		{constants.SyncOnClose, false},
		{constants.SyncBatch, false},
	}
	for _, tt := range tests {
		got := IsImmediate(api.SQLiteConfig{SyncStrategy: tt.strategy})
		if got != tt.want {
			t.Errorf("IsImmediate(%q) = %v, want %v", tt.strategy, got, tt.want)
		}
	}
}

func TestIsOnClose(t *testing.T) {
	if !IsOnClose(api.SQLiteConfig{SyncStrategy: constants.SyncOnClose}) {
		t.Error("expected true for on_close")
	}
	if IsOnClose(api.SQLiteConfig{SyncStrategy: constants.SyncImmediate}) {
		t.Error("expected false for immediate")
	}
}

func TestIsBatch(t *testing.T) {
	if !IsBatch(api.SQLiteConfig{SyncStrategy: constants.SyncBatch}) {
		t.Error("expected true for batch")
	}
	if IsBatch(api.SQLiteConfig{SyncStrategy: constants.SyncImmediate}) {
		t.Error("expected false for immediate")
	}
}
