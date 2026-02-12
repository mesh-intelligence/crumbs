package types

import "time"

// Link represents a directed edge in the entity graph.
// Implements: prd007-links-interface (R1: struct, R2: link types).
type Link struct {
	LinkID    string    `json:"link_id"`
	LinkType  string    `json:"link_type"`
	FromID    string    `json:"from_id"`
	ToID      string    `json:"to_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Link type constants per prd007-links-interface R2.2.
const (
	LinkTypeBelongsTo    = "belongs_to"
	LinkTypeChildOf      = "child_of"
	LinkTypeBranchesFrom = "branches_from"
	LinkTypeScopedTo     = "scoped_to"
)
