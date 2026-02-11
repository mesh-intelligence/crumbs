package sqlite

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mesh-intelligence/crumbs/pkg/types"
)

// jsonlFiles lists every JSONL file the backend manages.
// Implements: prd002-sqlite-backend R1.2.
var jsonlFiles = []string{
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

// ensureJSONLFiles creates empty JSONL files that do not yet exist.
// Implements: prd002-sqlite-backend R1.4, prd010 R5.1.
func ensureJSONLFiles(dataDir string) error {
	for _, name := range jsonlFiles {
		p := filepath.Join(dataDir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			f, err := os.Create(p)
			if err != nil {
				return fmt.Errorf("creating %s: %w", name, err)
			}
			f.Close()
		}
	}
	return nil
}

// writeJSONLAtomic replaces the contents of a JSONL file using the atomic
// temp-file pattern: write to .tmp, fsync, rename.
// Implements: prd002-sqlite-backend R5.2.
func writeJSONLAtomic(path string, records []json.RawMessage) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	w := bufio.NewWriter(f)
	for _, rec := range records {
		if _, err := w.Write(rec); err != nil {
			f.Close()
			os.Remove(tmp)
			return fmt.Errorf("writing record: %w", err)
		}
		if err := w.WriteByte('\n'); err != nil {
			f.Close()
			os.Remove(tmp)
			return fmt.Errorf("writing newline: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("flushing: %w", err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("syncing: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming: %w", err)
	}
	return nil
}

// readJSONLLines reads all non-empty lines from a JSONL file as raw JSON.
// Malformed lines are skipped with a logged warning (returned in warnings).
// Implements: prd002-sqlite-backend R2.1, R4.2, prd010 R3.2, R5.2.
func readJSONLLines(path string) ([]json.RawMessage, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	var lines []json.RawMessage
	var warnings []string
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !json.Valid([]byte(line)) {
			warnings = append(warnings, fmt.Sprintf("%s:%d: malformed JSON, skipping", filepath.Base(path), lineNum))
			continue
		}
		lines = append(lines, json.RawMessage(line))
	}
	if err := scanner.Err(); err != nil {
		return nil, warnings, fmt.Errorf("reading %s: %w", path, err)
	}
	return lines, warnings, nil
}

// crumbProperty is the JSONL shape for crumb_properties.jsonl.
type crumbProperty struct {
	CrumbID    string `json:"crumb_id"`
	PropertyID string `json:"property_id"`
	Value      string `json:"value"` // JSON-encoded value.
}

// Hydration: JSONL JSON to entity structs.

func hydrateCrumb(data json.RawMessage) (*types.Crumb, error) {
	var raw struct {
		CrumbID   string `json:"crumb_id"`
		Name      string `json:"name"`
		State     string `json:"state"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshaling crumb: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing crumb created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, raw.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing crumb updated_at: %w", err)
	}
	return &types.Crumb{
		CrumbID:    raw.CrumbID,
		Name:       raw.Name,
		State:      raw.State,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Properties: make(map[string]any),
	}, nil
}

func hydrateTrail(data json.RawMessage) (*types.Trail, error) {
	var raw struct {
		TrailID     string  `json:"trail_id"`
		State       string  `json:"state"`
		CreatedAt   string  `json:"created_at"`
		CompletedAt *string `json:"completed_at"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshaling trail: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing trail created_at: %w", err)
	}
	t := &types.Trail{
		TrailID:   raw.TrailID,
		State:     raw.State,
		CreatedAt: createdAt,
	}
	if raw.CompletedAt != nil && *raw.CompletedAt != "" {
		ct, err := time.Parse(time.RFC3339, *raw.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("parsing trail completed_at: %w", err)
		}
		t.CompletedAt = &ct
	}
	return t, nil
}

func hydrateProperty(data json.RawMessage) (*types.Property, error) {
	var raw struct {
		PropertyID  string `json:"property_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ValueType   string `json:"value_type"`
		CreatedAt   string `json:"created_at"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshaling property: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing property created_at: %w", err)
	}
	return &types.Property{
		PropertyID:  raw.PropertyID,
		Name:        raw.Name,
		Description: raw.Description,
		ValueType:   raw.ValueType,
		CreatedAt:   createdAt,
	}, nil
}

func hydrateCategory(data json.RawMessage) (*types.Category, error) {
	var raw struct {
		CategoryID string `json:"category_id"`
		PropertyID string `json:"property_id"`
		Name       string `json:"name"`
		Ordinal    int    `json:"ordinal"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshaling category: %w", err)
	}
	return &types.Category{
		CategoryID: raw.CategoryID,
		PropertyID: raw.PropertyID,
		Name:       raw.Name,
		Ordinal:    raw.Ordinal,
	}, nil
}

func hydrateMetadata(data json.RawMessage) (*types.Metadata, error) {
	var raw struct {
		MetadataID string  `json:"metadata_id"`
		TableName  string  `json:"table_name"`
		CrumbID    string  `json:"crumb_id"`
		PropertyID *string `json:"property_id"`
		Content    string  `json:"content"`
		CreatedAt  string  `json:"created_at"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshaling metadata: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing metadata created_at: %w", err)
	}
	return &types.Metadata{
		MetadataID: raw.MetadataID,
		TableName:  raw.TableName,
		CrumbID:    raw.CrumbID,
		PropertyID: raw.PropertyID,
		Content:    raw.Content,
		CreatedAt:  createdAt,
	}, nil
}

func hydrateLink(data json.RawMessage) (*types.Link, error) {
	var raw struct {
		LinkID    string `json:"link_id"`
		LinkType  string `json:"link_type"`
		FromID    string `json:"from_id"`
		ToID      string `json:"to_id"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshaling link: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing link created_at: %w", err)
	}
	return &types.Link{
		LinkID:    raw.LinkID,
		LinkType:  raw.LinkType,
		FromID:    raw.FromID,
		ToID:      raw.ToID,
		CreatedAt: createdAt,
	}, nil
}

func hydrateStash(data json.RawMessage) (*types.Stash, error) {
	var raw struct {
		StashID       string          `json:"stash_id"`
		Name          string          `json:"name"`
		StashType     string          `json:"stash_type"`
		Value         json.RawMessage `json:"value"`
		Version       int64           `json:"version"`
		CreatedAt     string          `json:"created_at"`
		UpdatedAt     string          `json:"updated_at"`
		LastOperation string          `json:"last_operation"`
		ChangedBy     *string         `json:"changed_by"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshaling stash: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing stash created_at: %w", err)
	}
	// UpdatedAt is stored in JSONL but not in the Stash struct; skip parsing.
	var value any
	if len(raw.Value) > 0 && string(raw.Value) != "null" {
		if err := json.Unmarshal(raw.Value, &value); err != nil {
			return nil, fmt.Errorf("parsing stash value: %w", err)
		}
	}
	return &types.Stash{
		StashID:       raw.StashID,
		Name:          raw.Name,
		StashType:     raw.StashType,
		Value:         value,
		Version:       raw.Version,
		CreatedAt:     createdAt,
		LastOperation: raw.LastOperation,
		ChangedBy:     raw.ChangedBy,
	}, nil
}

func hydrateStashHistory(data json.RawMessage) (*types.StashHistoryEntry, error) {
	var raw struct {
		HistoryID string          `json:"history_id"`
		StashID   string          `json:"stash_id"`
		Version   int64           `json:"version"`
		Value     json.RawMessage `json:"value"`
		Operation string          `json:"operation"`
		ChangedBy *string         `json:"changed_by"`
		CreatedAt string          `json:"created_at"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshaling stash history: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing stash_history created_at: %w", err)
	}
	var value any
	if len(raw.Value) > 0 && string(raw.Value) != "null" {
		if err := json.Unmarshal(raw.Value, &value); err != nil {
			return nil, fmt.Errorf("parsing stash_history value: %w", err)
		}
	}
	return &types.StashHistoryEntry{
		HistoryID: raw.HistoryID,
		StashID:   raw.StashID,
		Version:   raw.Version,
		Value:     value,
		Operation: raw.Operation,
		ChangedBy: raw.ChangedBy,
		CreatedAt: createdAt,
	}, nil
}

func hydrateCrumbProperty(data json.RawMessage) (*crumbProperty, error) {
	var cp crumbProperty
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("unmarshaling crumb_property: %w", err)
	}
	return &cp, nil
}

// Persistence: entity structs to JSONL JSON.

func dehydrateCrumb(c *types.Crumb) (json.RawMessage, error) {
	raw := struct {
		CrumbID   string `json:"crumb_id"`
		Name      string `json:"name"`
		State     string `json:"state"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		CrumbID:   c.CrumbID,
		Name:      c.Name,
		State:     c.State,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
	}
	return json.Marshal(raw)
}

func dehydrateTrail(t *types.Trail) (json.RawMessage, error) {
	raw := struct {
		TrailID     string  `json:"trail_id"`
		State       string  `json:"state"`
		CreatedAt   string  `json:"created_at"`
		CompletedAt *string `json:"completed_at"`
	}{
		TrailID:   t.TrailID,
		State:     t.State,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
	}
	if t.CompletedAt != nil {
		s := t.CompletedAt.Format(time.RFC3339)
		raw.CompletedAt = &s
	}
	return json.Marshal(raw)
}

func dehydrateProperty(p *types.Property) (json.RawMessage, error) {
	raw := struct {
		PropertyID  string `json:"property_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ValueType   string `json:"value_type"`
		CreatedAt   string `json:"created_at"`
	}{
		PropertyID:  p.PropertyID,
		Name:        p.Name,
		Description: p.Description,
		ValueType:   p.ValueType,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
	}
	return json.Marshal(raw)
}

func dehydrateCategory(c *types.Category) (json.RawMessage, error) {
	return json.Marshal(c)
}

func dehydrateMetadata(m *types.Metadata) (json.RawMessage, error) {
	raw := struct {
		MetadataID string  `json:"metadata_id"`
		TableName  string  `json:"table_name"`
		CrumbID    string  `json:"crumb_id"`
		PropertyID *string `json:"property_id"`
		Content    string  `json:"content"`
		CreatedAt  string  `json:"created_at"`
	}{
		MetadataID: m.MetadataID,
		TableName:  m.TableName,
		CrumbID:    m.CrumbID,
		PropertyID: m.PropertyID,
		Content:    m.Content,
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
	}
	return json.Marshal(raw)
}

func dehydrateLink(l *types.Link) (json.RawMessage, error) {
	raw := struct {
		LinkID    string `json:"link_id"`
		LinkType  string `json:"link_type"`
		FromID    string `json:"from_id"`
		ToID      string `json:"to_id"`
		CreatedAt string `json:"created_at"`
	}{
		LinkID:    l.LinkID,
		LinkType:  l.LinkType,
		FromID:    l.FromID,
		ToID:      l.ToID,
		CreatedAt: l.CreatedAt.Format(time.RFC3339),
	}
	return json.Marshal(raw)
}

func dehydrateStash(s *types.Stash) (json.RawMessage, error) {
	valueJSON, err := json.Marshal(s.Value)
	if err != nil {
		return nil, fmt.Errorf("marshaling stash value: %w", err)
	}
	raw := struct {
		StashID       string          `json:"stash_id"`
		Name          string          `json:"name"`
		StashType     string          `json:"stash_type"`
		Value         json.RawMessage `json:"value"`
		Version       int64           `json:"version"`
		CreatedAt     string          `json:"created_at"`
		UpdatedAt     string          `json:"updated_at"`
		LastOperation string          `json:"last_operation"`
		ChangedBy     *string         `json:"changed_by"`
	}{
		StashID:       s.StashID,
		Name:          s.Name,
		StashType:     s.StashType,
		Value:         valueJSON,
		Version:       s.Version,
		CreatedAt:     s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     time.Now().Format(time.RFC3339),
		LastOperation: s.LastOperation,
		ChangedBy:     s.ChangedBy,
	}
	return json.Marshal(raw)
}

func dehydrateStashHistory(h *types.StashHistoryEntry) (json.RawMessage, error) {
	valueJSON, err := json.Marshal(h.Value)
	if err != nil {
		return nil, fmt.Errorf("marshaling stash history value: %w", err)
	}
	raw := struct {
		HistoryID string          `json:"history_id"`
		StashID   string          `json:"stash_id"`
		Version   int64           `json:"version"`
		Value     json.RawMessage `json:"value"`
		Operation string          `json:"operation"`
		ChangedBy *string         `json:"changed_by"`
		CreatedAt string          `json:"created_at"`
	}{
		HistoryID: h.HistoryID,
		StashID:   h.StashID,
		Version:   h.Version,
		Value:     valueJSON,
		Operation: h.Operation,
		ChangedBy: h.ChangedBy,
		CreatedAt: h.CreatedAt.Format(time.RFC3339),
	}
	return json.Marshal(raw)
}

func dehydrateCrumbProperty(cp *crumbProperty) (json.RawMessage, error) {
	return json.Marshal(cp)
}
