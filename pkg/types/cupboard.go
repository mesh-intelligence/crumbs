package types

// Standard table names (prd001-cupboard-core R2.5).
const (
	TableCrumbs     = "crumbs"
	TableTrails     = "trails"
	TableProperties = "properties"
	TableMetadata   = "metadata"
	TableLinks      = "links"
	TableStashes    = "stashes"
)

// Supported backend values.
const (
	BackendSQLite = "sqlite"
)

// Supported sync strategies for SQLiteConfig.
const (
	SyncImmediate = "immediate"
	SyncOnClose   = "on_close"
	SyncBatch     = "batch"
)

// Cupboard is the core interface for storage access and lifecycle management
// (prd001-cupboard-core R2).
type Cupboard interface {
	GetTable(name string) (Table, error)
	Attach(config Config) error
	Detach() error
}

// Table provides uniform CRUD operations for all entity types
// (prd001-cupboard-core R3).
type Table interface {
	Get(id string) (any, error)
	Set(id string, data any) (string, error)
	Delete(id string) error
	Fetch(filter map[string]any) ([]any, error)
}

// Config holds the configuration for a Cupboard backend
// (prd001-cupboard-core R1).
type Config struct {
	Backend      string        `json:"backend" yaml:"backend"`
	DataDir      string        `json:"data_dir" yaml:"data_dir"`
	SQLiteConfig *SQLiteConfig `json:"sqlite_config,omitempty" yaml:"sqlite_config,omitempty"`
}

// Validate checks that the Config fields are valid (prd001-cupboard-core R1.2, R1.3).
func (c Config) Validate() error {
	if c.Backend == "" {
		return ErrBackendEmpty
	}
	if c.Backend != BackendSQLite {
		return ErrBackendUnknown
	}
	if c.DataDir == "" {
		return ErrDataDirEmpty
	}
	if c.SQLiteConfig != nil {
		return c.SQLiteConfig.Validate()
	}
	return nil
}

// SQLiteConfig holds SQLite-specific configuration options.
type SQLiteConfig struct {
	SyncStrategy  string `json:"sync_strategy,omitempty" yaml:"sync_strategy,omitempty"`
	BatchSize     int    `json:"batch_size,omitempty" yaml:"batch_size,omitempty"`
	BatchInterval int    `json:"batch_interval,omitempty" yaml:"batch_interval,omitempty"`
}

// Validate checks that the SQLiteConfig fields are valid.
func (sc SQLiteConfig) Validate() error {
	switch sc.SyncStrategy {
	case "", SyncImmediate, SyncOnClose, SyncBatch:
		// valid
	default:
		return ErrSyncStrategyUnknown
	}
	if sc.SyncStrategy == SyncBatch {
		if sc.BatchSize <= 0 {
			return ErrBatchSizeInvalid
		}
		if sc.BatchInterval <= 0 {
			return ErrBatchIntervalInvalid
		}
	}
	return nil
}

// GetSyncStrategy returns the sync strategy, defaulting to "immediate".
func (sc SQLiteConfig) GetSyncStrategy() string {
	if sc.SyncStrategy == "" {
		return SyncImmediate
	}
	return sc.SyncStrategy
}

// GetBatchSize returns the batch size.
func (sc SQLiteConfig) GetBatchSize() int {
	return sc.BatchSize
}

// GetBatchInterval returns the batch interval.
func (sc SQLiteConfig) GetBatchInterval() int {
	return sc.BatchInterval
}
