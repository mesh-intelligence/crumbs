// Package api defines the public interfaces for Crumbs.
// This package contains only contracts — no domain logic, no persistence.
//
// Implements: prd001-cupboard-core R2, R3
package api

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
