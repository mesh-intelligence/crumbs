package api

import "errors"

// Config validation errors (prd001-cupboard-core R1.4).
var (
	ErrBackendEmpty         = errors.New("backend must not be empty")
	ErrBackendUnknown       = errors.New("unknown backend")
	ErrDataDirEmpty         = errors.New("data dir must not be empty")
	ErrSyncStrategyUnknown  = errors.New("unknown sync strategy")
	ErrBatchSizeInvalid     = errors.New("batch size must be positive")
	ErrBatchIntervalInvalid = errors.New("batch interval must be positive")
)

// Cupboard lifecycle errors (prd001-cupboard-core R7.1).
var (
	ErrCupboardDetached = errors.New("cupboard is detached")
	ErrAlreadyAttached  = errors.New("cupboard is already attached")
	ErrTableNotFound    = errors.New("table not found")
)

// Table operation errors (prd001-cupboard-core R7.2).
var (
	ErrNotFound    = errors.New("entity not found")
	ErrInvalidID   = errors.New("invalid entity ID")
	ErrInvalidData = errors.New("invalid entity data")
)

// Filter errors.
var (
	ErrInvalidFilter = errors.New("invalid filter value type")
)
