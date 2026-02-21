package types

import "time"

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

// Property represents a property definition (prd004-properties-interface).
type Property struct {
	PropertyID  string    `json:"property_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ValueType   string    `json:"value_type"`
	CreatedAt   time.Time `json:"created_at"`
}

// Category represents a categorical value for a property
// (prd004-properties-interface).
type Category struct {
	CategoryID string `json:"category_id"`
	PropertyID string `json:"property_id"`
	Name       string `json:"name"`
	Ordinal    int    `json:"ordinal"`
}

// Metadata represents a supplementary data entry attached to a crumb
// (prd005-metadata-interface).
type Metadata struct {
	MetadataID string    `json:"metadata_id"`
	CrumbID    string    `json:"crumb_id"`
	TableName  string    `json:"table_name"`
	Content    string    `json:"content"`
	PropertyID *string   `json:"property_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Link represents a typed relationship between entities
// (prd002-sqlite-backend).
type Link struct {
	LinkID    string    `json:"link_id"`
	LinkType  string    `json:"link_type"`
	FromID    string    `json:"from_id"`
	ToID      string    `json:"to_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Stash represents shared state for trails (prd008-stash-interface).
type Stash struct {
	StashID       string    `json:"stash_id"`
	Name          string    `json:"name"`
	StashType     string    `json:"stash_type"`
	Value         any       `json:"value"`
	Version       int64     `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	LastOperation string    `json:"last_operation"`
	ChangedBy     *string   `json:"changed_by,omitempty"`
}

// SetValue updates the stash value. Increments Version and records the
// operation. Returns ErrInvalidStashType if the stash is a lock or counter
// (use Acquire/Release or Increment instead).
func (s *Stash) SetValue(value any) error {
	if s.StashType == StashLock || s.StashType == StashCounter {
		return ErrInvalidStashType
	}
	s.Value = value
	s.Version++
	s.LastOperation = "set"
	return nil
}

// GetValue returns the current stash value.
func (s *Stash) GetValue() any {
	return s.Value
}

// Increment atomically adds delta to a counter stash value and returns
// the new value. Returns ErrInvalidStashType if the stash is not a counter.
func (s *Stash) Increment(delta int64) (int64, error) {
	if s.StashType != StashCounter {
		return 0, ErrInvalidStashType
	}
	current, ok := s.Value.(int64)
	if !ok {
		current = 0
	}
	current += delta
	s.Value = current
	s.Version++
	s.LastOperation = "increment"
	return current, nil
}

// Acquire obtains the lock for mutual exclusion.
// Returns ErrInvalidStashType if the stash is not a lock,
// ErrInvalidHolder if holder is empty, or ErrLockHeld if the
// lock is already held by a different holder.
func (s *Stash) Acquire(holder string) error {
	if s.StashType != StashLock {
		return ErrInvalidStashType
	}
	if holder == "" {
		return ErrInvalidHolder
	}
	if s.Value != nil && s.Value != "" {
		current, ok := s.Value.(string)
		if ok && current != "" && current != holder {
			return ErrLockHeld
		}
	}
	s.Value = holder
	s.Version++
	s.LastOperation = "acquire"
	s.ChangedBy = &holder
	return nil
}

// Release releases the lock. Returns ErrInvalidStashType if the stash
// is not a lock, ErrInvalidHolder if holder is empty, or
// ErrNotLockHolder if the caller is not the current lock holder.
func (s *Stash) Release(holder string) error {
	if s.StashType != StashLock {
		return ErrInvalidStashType
	}
	if holder == "" {
		return ErrInvalidHolder
	}
	current, ok := s.Value.(string)
	if !ok || current != holder {
		return ErrNotLockHolder
	}
	s.Value = ""
	s.Version++
	s.LastOperation = "release"
	s.ChangedBy = &holder
	return nil
}

// StashHistoryEntry records a single mutation to a stash
// (prd008-stash-interface).
type StashHistoryEntry struct {
	HistoryID string    `json:"history_id"`
	StashID   string    `json:"stash_id"`
	Version   int64     `json:"version"`
	Value     any       `json:"value"`
	Operation string    `json:"operation"`
	ChangedBy *string   `json:"changed_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Schema represents a metadata schema registration
// (prd005-metadata-interface).
type Schema struct {
	SchemaName  string `json:"schema_name"`
	Description string `json:"description"`
	ContentType string `json:"content_type"`
}
