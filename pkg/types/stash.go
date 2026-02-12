package types

import "time"

// Stash represents shared state scoped to a trail or global.
// Implements: prd008-stash-interface (R1: struct, R2: types, R4-R6: entity methods, R7: history).
type Stash struct {
	StashID       string    `json:"stash_id"`
	Name          string    `json:"name"`
	StashType     string    `json:"stash_type"`
	Value         any       `json:"value"`
	Version       int64     `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	LastOperation string    `json:"last_operation"`
	ChangedBy     *string   `json:"changed_by"`
}

// Stash type constants per prd008-stash-interface R2.1.
const (
	StashTypeResource = "resource"
	StashTypeArtifact = "artifact"
	StashTypeContext  = "context"
	StashTypeCounter  = "counter"
	StashTypeLock     = "lock"
)

// Stash operation constants per prd008-stash-interface R7.3.
const (
	StashOpCreate    = "create"
	StashOpSet       = "set"
	StashOpIncrement = "increment"
	StashOpAcquire   = "acquire"
	StashOpRelease   = "release"
)

// StashHistoryEntry records a single mutation of a stash.
// Backends use this struct when returning history data.
// See prd008-stash-interface R7.2.
type StashHistoryEntry struct {
	HistoryID string    `json:"history_id"`
	StashID   string    `json:"stash_id"`
	Version   int64     `json:"version"`
	Value     any       `json:"value"`
	Operation string    `json:"operation"`
	ChangedBy *string   `json:"changed_by"`
	CreatedAt time.Time `json:"created_at"`
}

// SetValue updates the stash value.
// Returns ErrInvalidStashType if called on a lock-type stash.
// See prd008-stash-interface R4.2.
func (s *Stash) SetValue(value any) error {
	if s.StashType == StashTypeLock {
		return ErrInvalidStashType
	}
	s.Value = value
	s.Version++
	s.LastOperation = StashOpSet
	return nil
}

// GetValue retrieves the current value.
// See prd008-stash-interface R4.3.
func (s *Stash) GetValue() any {
	return s.Value
}

// Increment atomically adds delta to a counter-type stash.
// Returns the new counter value.
// Returns ErrInvalidStashType if the stash is not a counter.
// See prd008-stash-interface R5.2.
func (s *Stash) Increment(delta int64) (int64, error) {
	if s.StashType != StashTypeCounter {
		return 0, ErrInvalidStashType
	}
	var current int64
	if v, ok := s.Value.(map[string]any); ok {
		if raw, exists := v["value"]; exists {
			switch n := raw.(type) {
			case int64:
				current = n
			case float64:
				current = int64(n)
			}
		}
	}
	current += delta
	s.Value = map[string]any{"value": current}
	s.Version++
	s.LastOperation = StashOpIncrement
	return current, nil
}

// Acquire obtains the lock for the given holder.
// Returns ErrInvalidStashType if the stash is not a lock.
// Returns ErrInvalidHolder if holder is empty.
// Returns ErrLockHeld if the lock is held by another holder.
// Reentrant: acquiring a lock already held by the same holder succeeds.
// See prd008-stash-interface R6.2.
func (s *Stash) Acquire(holder string) error {
	if s.StashType != StashTypeLock {
		return ErrInvalidStashType
	}
	if holder == "" {
		return ErrInvalidHolder
	}
	if s.Value != nil {
		if v, ok := s.Value.(map[string]any); ok {
			if h, exists := v["holder"]; exists {
				if h == holder {
					return nil
				}
				return ErrLockHeld
			}
		}
	}
	s.Value = map[string]any{
		"holder":      holder,
		"acquired_at": time.Now().Format(time.RFC3339),
	}
	s.Version++
	s.LastOperation = StashOpAcquire
	return nil
}

// Release releases the lock held by the given holder.
// Returns ErrInvalidStashType if the stash is not a lock.
// Returns ErrNotLockHolder if the lock is not held by the specified holder.
// See prd008-stash-interface R6.3.
func (s *Stash) Release(holder string) error {
	if s.StashType != StashTypeLock {
		return ErrInvalidStashType
	}
	if s.Value == nil {
		return ErrNotLockHolder
	}
	if v, ok := s.Value.(map[string]any); ok {
		if h, exists := v["holder"]; exists && h == holder {
			s.Value = nil
			s.Version++
			s.LastOperation = StashOpRelease
			return nil
		}
	}
	return ErrNotLockHolder
}
