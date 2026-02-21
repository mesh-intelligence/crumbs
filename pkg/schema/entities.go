package schema

import "time"

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
