package schema

import (
	"time"

	"github.com/petar-djukic/crumbs/pkg/constants"
)

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
	if s.StashType == constants.StashLock || s.StashType == constants.StashCounter {
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
	if s.StashType != constants.StashCounter {
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
	if s.StashType != constants.StashLock {
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
	if s.StashType != constants.StashLock {
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
