// Package constants centralizes all domain rules as named constants.
// Prevents magic strings across the codebase.
//
// Implements: prd001-cupboard-core R2.5, prd002-sqlite-backend,
// prd003-crumbs-interface R2.1, prd004-properties-interface,
// prd006-trails-interface, prd008-stash-interface R3
package constants

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

// Crumb states (prd003-crumbs-interface R2.1).
const (
	CrumbDraft   = "draft"
	CrumbPending = "pending"
	CrumbReady   = "ready"
	CrumbTaken   = "taken"
	CrumbPebble  = "pebble"
	CrumbDust    = "dust"
)

// Trail states (prd006-trails-interface).
const (
	TrailDraft     = "draft"
	TrailPending   = "pending"
	TrailActive    = "active"
	TrailCompleted = "completed"
	TrailAbandoned = "abandoned"
)

// Stash types (prd008-stash-interface R3).
const (
	StashResource = "resource"
	StashArtifact = "artifact"
	StashContext  = "context"
	StashCounter  = "counter"
	StashLock     = "lock"
)

// Link types (prd002-sqlite-backend).
const (
	LinkBelongsTo    = "belongs_to"
	LinkChildOf      = "child_of"
	LinkBranchesFrom = "branches_from"
	LinkScopedTo     = "scoped_to"
)

// Property value types (prd004-properties-interface).
const (
	ValueTypeCategorical = "categorical"
	ValueTypeText        = "text"
	ValueTypeInteger     = "integer"
	ValueTypeBoolean     = "boolean"
	ValueTypeTimestamp   = "timestamp"
	ValueTypeList        = "list"
)
