package types

import "time"

// Metadata represents supplementary information attached to a crumb.
// Implements: prd005-metadata-interface R1.
type Metadata struct {
	MetadataID string    `json:"metadata_id"`
	CrumbID    string    `json:"crumb_id"`
	TableName  string    `json:"table_name"`
	Content    string    `json:"content"`
	PropertyID *string   `json:"property_id"`
	CreatedAt  time.Time `json:"created_at"`
}
