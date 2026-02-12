package types

import "time"

// Property defines a named, typed attribute that can be set on crumbs.
// Implements: prd004-properties-interface (R1: struct, R3: value types).
type Property struct {
	PropertyID  string    `json:"property_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ValueType   string    `json:"value_type"`
	CreatedAt   time.Time `json:"created_at"`
}

// Category defines a valid value for a categorical property.
// Implements: prd004-properties-interface R2.
type Category struct {
	CategoryID string `json:"category_id"`
	PropertyID string `json:"property_id"`
	Name       string `json:"name"`
	Ordinal    int    `json:"ordinal"`
}

// Value type constants per prd004-properties-interface R3.1.
const (
	ValueTypeCategorical = "categorical"
	ValueTypeText        = "text"
	ValueTypeInteger     = "integer"
	ValueTypeBoolean     = "boolean"
	ValueTypeTimestamp   = "timestamp"
	ValueTypeList        = "list"
)
