// Package types defines shared public types and interfaces for the Crumbs system.
// Implements: prd001-cupboard-core (R2: Cupboard Interface, R2.5: standard table names, R7.1: lifecycle errors);
//             docs/ARCHITECTURE § Main Interface.
package types

import "errors"

// Cupboard is the contract between applications and storage backends.
// Applications call GetTable to access a named table, Attach to initialize
// a backend, and Detach to release resources.
type Cupboard interface {
	// GetTable returns a Table for the specified table name.
	// Returns ErrTableNotFound if the name is not a standard table.
	GetTable(name string) (Table, error)

	// Attach initializes the backend using the provided Config.
	// Returns ErrAlreadyAttached if the cupboard is already attached.
	Attach(config Config) error

	// Detach releases all resources held by the cupboard.
	// Subsequent operations return ErrCupboardDetached.
	Detach() error
}

// Standard table names (prd001-cupboard-core R2.5).
const (
	TableCrumbs     = "crumbs"
	TableTrails     = "trails"
	TableProperties = "properties"
	TableMetadata   = "metadata"
	TableLinks      = "links"
	TableStashes    = "stashes"
)

// Cupboard lifecycle errors (prd001-cupboard-core R7.1).
var (
	ErrCupboardDetached = errors.New("cupboard is detached")
	ErrAlreadyAttached  = errors.New("cupboard is already attached")
	ErrTableNotFound    = errors.New("table not found")
)
