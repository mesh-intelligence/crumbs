package engine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReadJSONL reads a JSONL file and decodes each non-empty line into a
// map[string]any. Malformed lines are skipped and reported via the
// returned warnings slice (R2.1, R4.2, R7.1).
func ReadJSONL(path string) ([]map[string]any, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var (
		records  []map[string]any
		warnings []string
		lineNum  int
	)
	scanner := bufio.NewScanner(f)
	// Allow up to 1 MiB per line.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal(line, &obj); err != nil {
			warnings = append(warnings, fmt.Sprintf("%s:%d: %v", filepath.Base(path), lineNum, err))
			continue
		}
		records = append(records, obj)
	}
	if err := scanner.Err(); err != nil {
		return records, warnings, fmt.Errorf("scan %s: %w", path, err)
	}
	return records, warnings, nil
}

// WriteJSONL atomically writes records to a JSONL file using the
// temp-file → fsync → rename pattern (R5.2, R16.7).
func WriteJSONL(path string, records []map[string]any) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".jsonl-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	// Clean up the temp file on any error path.
	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	w := bufio.NewWriter(tmp)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, rec := range records {
		if err := enc.Encode(rec); err != nil {
			return fmt.Errorf("encode record: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("fsync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename %s → %s: %w", tmpName, path, err)
	}
	success = true
	return nil
}

// AppendJSONL appends a single record to a JSONL file. This is used
// for append-only files like stash_history.jsonl where rewriting the
// entire file is unnecessary.
func AppendJSONL(path string, record map[string]any) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open %s for append: %w", path, err)
	}
	defer f.Close()

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}
	data = append(data, '\n')
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("fsync %s: %w", path, err)
	}
	return nil
}

// EnsureJSONLFiles creates empty JSONL files that do not already exist
// in the given directory (R1.4).
func EnsureJSONLFiles(dir string) error {
	for _, name := range JSONLFiles {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			continue
		}
		f, err := os.Create(p)
		if err != nil {
			return fmt.Errorf("create %s: %w", name, err)
		}
		f.Close()
	}
	return nil
}

// JSONLFiles lists all JSONL files managed by the backend (R1.2).
var JSONLFiles = []string{
	"crumbs.jsonl",
	"trails.jsonl",
	"links.jsonl",
	"properties.jsonl",
	"categories.jsonl",
	"crumb_properties.jsonl",
	"metadata.jsonl",
	"stashes.jsonl",
	"stash_history.jsonl",
}

// ReadJSONLTyped reads a JSONL file and decodes each line into a value
// of type T using json.Decoder. Malformed lines are skipped and
// reported via warnings.
func ReadJSONLTyped[T any](path string) ([]T, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var (
		results  []T
		warnings []string
		lineNum  int
	)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var v T
		if err := json.Unmarshal(line, &v); err != nil {
			warnings = append(warnings, fmt.Sprintf("%s:%d: %v", filepath.Base(path), lineNum, err))
			continue
		}
		results = append(results, v)
	}
	if err := scanner.Err(); err != nil {
		return results, warnings, fmt.Errorf("scan %s: %w", path, err)
	}
	return results, warnings, nil
}

// WriteJSONLTyped atomically writes typed records to a JSONL file using
// the temp-file → fsync → rename pattern.
func WriteJSONLTyped[T any](path string, records []T) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".jsonl-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	w := bufio.NewWriter(tmp)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, rec := range records {
		if err := enc.Encode(rec); err != nil {
			return fmt.Errorf("encode record: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("fsync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	success = true
	return nil
}
