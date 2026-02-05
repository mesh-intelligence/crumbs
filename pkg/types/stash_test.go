package types

import (
	"errors"
	"testing"
)

func TestStash_SetValue(t *testing.T) {
	t.Run("sets value on context stash", func(t *testing.T) {
		s := &Stash{StashType: StashTypeContext, Version: 1}

		err := s.SetValue(map[string]any{"timeout": 30})

		if err != nil {
			t.Errorf("SetValue() error = %v", err)
		}
		if s.Version != 2 {
			t.Errorf("SetValue() version = %v, want 2", s.Version)
		}
		val, ok := s.Value.(map[string]any)
		if !ok || val["timeout"] != 30 {
			t.Errorf("SetValue() value = %v, want map with timeout=30", s.Value)
		}
	})

	t.Run("returns error for lock type", func(t *testing.T) {
		s := &Stash{StashType: StashTypeLock, Version: 1}

		err := s.SetValue("test")

		if !errors.Is(err, ErrInvalidStashType) {
			t.Errorf("SetValue() error = %v, want %v", err, ErrInvalidStashType)
		}
		if s.Version != 1 {
			t.Error("SetValue() should not increment version on error")
		}
	})

	t.Run("works for resource type", func(t *testing.T) {
		s := &Stash{StashType: StashTypeResource, Version: 1}

		err := s.SetValue(map[string]any{"uri": "https://example.com"})

		if err != nil {
			t.Errorf("SetValue() error = %v", err)
		}
	})

	t.Run("works for artifact type", func(t *testing.T) {
		s := &Stash{StashType: StashTypeArtifact, Version: 1}

		err := s.SetValue(map[string]any{"path": "/tmp/output.txt"})

		if err != nil {
			t.Errorf("SetValue() error = %v", err)
		}
	})

	t.Run("works for counter type", func(t *testing.T) {
		s := &Stash{StashType: StashTypeCounter, Version: 1}

		err := s.SetValue(map[string]any{"value": int64(100)})

		if err != nil {
			t.Errorf("SetValue() error = %v", err)
		}
	})
}

func TestStash_GetValue(t *testing.T) {
	t.Run("returns value", func(t *testing.T) {
		s := &Stash{Value: map[string]any{"key": "val"}}

		val := s.GetValue()

		if val == nil {
			t.Error("GetValue() should return value")
		}
	})

	t.Run("returns nil for empty", func(t *testing.T) {
		s := &Stash{}

		val := s.GetValue()

		if val != nil {
			t.Errorf("GetValue() = %v, want nil", val)
		}
	})
}

func TestStash_Increment(t *testing.T) {
	t.Run("increments from zero", func(t *testing.T) {
		s := &Stash{StashType: StashTypeCounter, Version: 1}

		newVal, err := s.Increment(5)

		if err != nil {
			t.Errorf("Increment() error = %v", err)
		}
		if newVal != 5 {
			t.Errorf("Increment() = %v, want 5", newVal)
		}
		if s.Version != 2 {
			t.Errorf("Increment() version = %v, want 2", s.Version)
		}
	})

	t.Run("increments from existing value", func(t *testing.T) {
		s := &Stash{
			StashType: StashTypeCounter,
			Version:   1,
			Value:     map[string]any{"value": int64(10)},
		}

		newVal, err := s.Increment(3)

		if err != nil {
			t.Errorf("Increment() error = %v", err)
		}
		if newVal != 13 {
			t.Errorf("Increment() = %v, want 13", newVal)
		}
	})

	t.Run("handles float64 value from JSON", func(t *testing.T) {
		s := &Stash{
			StashType: StashTypeCounter,
			Version:   1,
			Value:     map[string]any{"value": float64(10)},
		}

		newVal, err := s.Increment(5)

		if err != nil {
			t.Errorf("Increment() error = %v", err)
		}
		if newVal != 15 {
			t.Errorf("Increment() = %v, want 15", newVal)
		}
	})

	t.Run("decrements with negative delta", func(t *testing.T) {
		s := &Stash{
			StashType: StashTypeCounter,
			Version:   1,
			Value:     map[string]any{"value": int64(10)},
		}

		newVal, err := s.Increment(-3)

		if err != nil {
			t.Errorf("Increment() error = %v", err)
		}
		if newVal != 7 {
			t.Errorf("Increment() = %v, want 7", newVal)
		}
	})

	t.Run("returns error for non-counter type", func(t *testing.T) {
		s := &Stash{StashType: StashTypeContext, Version: 1}

		_, err := s.Increment(1)

		if !errors.Is(err, ErrInvalidStashType) {
			t.Errorf("Increment() error = %v, want %v", err, ErrInvalidStashType)
		}
	})
}

func TestStash_Acquire(t *testing.T) {
	t.Run("acquires unlocked lock", func(t *testing.T) {
		s := &Stash{StashType: StashTypeLock, Version: 1}

		err := s.Acquire("worker-1")

		if err != nil {
			t.Errorf("Acquire() error = %v", err)
		}
		if s.Version != 2 {
			t.Errorf("Acquire() version = %v, want 2", s.Version)
		}
		lockData, ok := s.Value.(map[string]any)
		if !ok {
			t.Fatal("Acquire() value should be a map")
		}
		if lockData["holder"] != "worker-1" {
			t.Errorf("Acquire() holder = %v, want worker-1", lockData["holder"])
		}
	})

	t.Run("reentrant acquire succeeds", func(t *testing.T) {
		s := &Stash{
			StashType: StashTypeLock,
			Version:   2,
			Value:     map[string]any{"holder": "worker-1", "acquired_at": "2024-01-01T00:00:00Z"},
		}

		err := s.Acquire("worker-1")

		if err != nil {
			t.Errorf("Acquire() reentrant should succeed, got %v", err)
		}
	})

	t.Run("returns error when held by another", func(t *testing.T) {
		s := &Stash{
			StashType: StashTypeLock,
			Version:   2,
			Value:     map[string]any{"holder": "worker-1"},
		}

		err := s.Acquire("worker-2")

		if !errors.Is(err, ErrLockHeld) {
			t.Errorf("Acquire() error = %v, want %v", err, ErrLockHeld)
		}
	})

	t.Run("returns error for empty holder", func(t *testing.T) {
		s := &Stash{StashType: StashTypeLock, Version: 1}

		err := s.Acquire("")

		if !errors.Is(err, ErrInvalidHolder) {
			t.Errorf("Acquire() error = %v, want %v", err, ErrInvalidHolder)
		}
	})

	t.Run("returns error for non-lock type", func(t *testing.T) {
		s := &Stash{StashType: StashTypeCounter, Version: 1}

		err := s.Acquire("worker-1")

		if !errors.Is(err, ErrInvalidStashType) {
			t.Errorf("Acquire() error = %v, want %v", err, ErrInvalidStashType)
		}
	})
}

func TestStash_Release(t *testing.T) {
	t.Run("releases held lock", func(t *testing.T) {
		s := &Stash{
			StashType: StashTypeLock,
			Version:   2,
			Value:     map[string]any{"holder": "worker-1"},
		}

		err := s.Release("worker-1")

		if err != nil {
			t.Errorf("Release() error = %v", err)
		}
		if s.Value != nil {
			t.Errorf("Release() value = %v, want nil", s.Value)
		}
		if s.Version != 3 {
			t.Errorf("Release() version = %v, want 3", s.Version)
		}
	})

	t.Run("returns error for wrong holder", func(t *testing.T) {
		s := &Stash{
			StashType: StashTypeLock,
			Version:   2,
			Value:     map[string]any{"holder": "worker-1"},
		}

		err := s.Release("worker-2")

		if !errors.Is(err, ErrNotLockHolder) {
			t.Errorf("Release() error = %v, want %v", err, ErrNotLockHolder)
		}
	})

	t.Run("returns error for unlocked lock", func(t *testing.T) {
		s := &Stash{StashType: StashTypeLock, Version: 1}

		err := s.Release("worker-1")

		if !errors.Is(err, ErrNotLockHolder) {
			t.Errorf("Release() error = %v, want %v", err, ErrNotLockHolder)
		}
	})

	t.Run("returns error for non-lock type", func(t *testing.T) {
		s := &Stash{StashType: StashTypeCounter, Version: 1}

		err := s.Release("worker-1")

		if !errors.Is(err, ErrInvalidStashType) {
			t.Errorf("Release() error = %v, want %v", err, ErrInvalidStashType)
		}
	})
}
