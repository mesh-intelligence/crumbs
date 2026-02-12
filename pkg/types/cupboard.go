package types

import "errors"

// Cupboard defines the core contract for storage access and lifecycle
// management. Backends implement this interface.
// See prd001-cupboard-core R2.
type Cupboard interface {
	GetTable(name string) (Table, error)
	Attach(config Config) error
	Detach() error
}

// Cupboard lifecycle errors per prd001-cupboard-core R7.1.
var (
	ErrCupboardDetached = errors.New("cupboard is detached")
	ErrAlreadyAttached  = errors.New("cupboard is already attached")
	ErrTableNotFound    = errors.New("table not found")
)
