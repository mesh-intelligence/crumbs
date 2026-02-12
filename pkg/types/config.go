// Package types defines shared interfaces, entity structs, constants,
// and sentinel errors for the Crumbs storage system.
// Implements: prd001-cupboard-core (Config R1, Errors R7);
//             docs/ARCHITECTURE § Main Interfaces.
package types

import "errors"

// Config selects a backend and provides backend-specific parameters.
// See prd001-cupboard-core R1.
type Config struct {
	Backend string `json:"backend" yaml:"backend"`
	DataDir string `json:"data_dir" yaml:"data_dir"`
}

// Config validation errors per prd001-cupboard-core R1.4.
var (
	ErrBackendEmpty         = errors.New("backend must not be empty")
	ErrBackendUnknown       = errors.New("unknown backend")
	ErrSyncStrategyUnknown  = errors.New("unknown sync strategy")
	ErrBatchSizeInvalid     = errors.New("batch size must be positive")
	ErrBatchIntervalInvalid = errors.New("batch interval must be positive")
)
