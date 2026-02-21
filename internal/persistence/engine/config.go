package engine

import (
	"github.com/petar-djukic/crumbs/pkg/api"
	"github.com/petar-djukic/crumbs/pkg/constants"
)

// DefaultBatchSize is the default number of writes before a batch flush (R16.5).
const DefaultBatchSize = 100

// DefaultBatchInterval is the default seconds between batch flushes (R16.5).
const DefaultBatchInterval = 5

// ResolveSyncConfig returns a fully-populated SQLiteConfig with defaults
// applied for any unset fields (R16.1, R16.2).
func ResolveSyncConfig(cfg *api.SQLiteConfig) api.SQLiteConfig {
	if cfg == nil {
		return api.SQLiteConfig{
			SyncStrategy:  constants.SyncImmediate,
			BatchSize:     DefaultBatchSize,
			BatchInterval: DefaultBatchInterval,
		}
	}
	resolved := *cfg
	if resolved.SyncStrategy == "" {
		resolved.SyncStrategy = constants.SyncImmediate
	}
	if resolved.SyncStrategy == constants.SyncBatch {
		if resolved.BatchSize <= 0 {
			resolved.BatchSize = DefaultBatchSize
		}
		if resolved.BatchInterval <= 0 {
			resolved.BatchInterval = DefaultBatchInterval
		}
	}
	return resolved
}

// IsImmediate reports whether the sync strategy requires JSONL writes
// after every SQLite commit (R16.2).
func IsImmediate(cfg api.SQLiteConfig) bool {
	s := cfg.GetSyncStrategy()
	return s == "" || s == constants.SyncImmediate
}

// IsOnClose reports whether JSONL writes are deferred until Detach (R16.3).
func IsOnClose(cfg api.SQLiteConfig) bool {
	return cfg.GetSyncStrategy() == constants.SyncOnClose
}

// IsBatch reports whether JSONL writes are batched by count or interval (R16.4).
func IsBatch(cfg api.SQLiteConfig) bool {
	return cfg.GetSyncStrategy() == constants.SyncBatch
}
