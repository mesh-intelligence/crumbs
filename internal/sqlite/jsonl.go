// Implements: prd002-sqlite-backend (R2: JSONL File Format, R5.2: atomic write);
//             docs/ARCHITECTURE § SQLite Backend.
package sqlite

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// loadJSONL reads a JSONL file and decodes each non-empty line into a value
// of type T. Malformed lines are skipped per prd002-sqlite-backend R4.2.
func loadJSONL[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	var result []T
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var v T
		if err := json.Unmarshal(line, &v); err != nil {
			// Skip malformed lines per R4.2.
			continue
		}
		result = append(result, v)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return result, nil
}

// persistJSONL writes a slice of values to a JSONL file using the atomic write
// pattern: write to a temp file, fsync, then rename (prd002-sqlite-backend R5.2).
func persistJSONL[T any](path string, items []T) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".jsonl-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	w := bufio.NewWriter(tmp)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, item := range items {
		if err := enc.Encode(item); err != nil {
			tmp.Close()
			os.Remove(tmpName)
			return fmt.Errorf("encoding JSONL record: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("flushing JSONL file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("syncing JSONL file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming JSONL file: %w", err)
	}
	return nil
}
