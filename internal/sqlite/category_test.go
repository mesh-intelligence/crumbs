package sqlite

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mesh-intelligence/crumbs/pkg/types"
)

func TestBackend_DefineCategory(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "crumbs-category-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and attach backend
	backend := NewBackend()
	config := types.Config{
		Backend: "sqlite",
		DataDir: tmpDir,
	}
	if err := backend.Attach(config); err != nil {
		t.Fatalf("Attach failed: %v", err)
	}
	defer backend.Detach()

	// First create a categorical property
	propsTable, err := backend.GetTable(types.PropertiesTable)
	if err != nil {
		t.Fatalf("GetTable(properties) failed: %v", err)
	}

	prop := &types.Property{
		Name:        "test_category_prop",
		Description: "Test categorical property",
		ValueType:   types.ValueTypeCategorical,
	}
	propID, err := propsTable.Set("", prop)
	if err != nil {
		t.Fatalf("Set property failed: %v", err)
	}
	prop.PropertyID = propID

	tests := []struct {
		name    string
		catName string
		ordinal int
		wantErr error
	}{
		{
			name:    "create first category",
			catName: "high",
			ordinal: 1,
		},
		{
			name:    "create second category",
			catName: "medium",
			ordinal: 2,
		},
		{
			name:    "create category with negative ordinal",
			catName: "critical",
			ordinal: -1,
		},
		{
			name:    "duplicate name fails",
			catName: "high",
			ordinal: 10,
			wantErr: types.ErrDuplicateName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := backend.DefineCategory(propID, tt.catName, tt.ordinal)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("DefineCategory() error = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("DefineCategory() unexpected error = %v", err)
				return
			}

			if cat.CategoryID == "" {
				t.Error("DefineCategory() returned category with empty CategoryID")
			}
			if cat.PropertyID != propID {
				t.Errorf("DefineCategory() PropertyID = %v, want %v", cat.PropertyID, propID)
			}
			if cat.Name != tt.catName {
				t.Errorf("DefineCategory() Name = %v, want %v", cat.Name, tt.catName)
			}
			if cat.Ordinal != tt.ordinal {
				t.Errorf("DefineCategory() Ordinal = %v, want %v", cat.Ordinal, tt.ordinal)
			}
		})
	}
}

func TestBackend_GetCategories(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "crumbs-category-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and attach backend
	backend := NewBackend()
	config := types.Config{
		Backend: "sqlite",
		DataDir: tmpDir,
	}
	if err := backend.Attach(config); err != nil {
		t.Fatalf("Attach failed: %v", err)
	}
	defer backend.Detach()

	// Create a categorical property
	propsTable, err := backend.GetTable(types.PropertiesTable)
	if err != nil {
		t.Fatalf("GetTable(properties) failed: %v", err)
	}

	prop := &types.Property{
		Name:        "ordering_test",
		Description: "Test ordering property",
		ValueType:   types.ValueTypeCategorical,
	}
	propID, err := propsTable.Set("", prop)
	if err != nil {
		t.Fatalf("Set property failed: %v", err)
	}

	// Define categories in non-sorted order
	categories := []struct {
		name    string
		ordinal int
	}{
		{"low", 3},
		{"high", 1},
		{"medium", 2},
		{"critical", 0},
	}
	for _, cat := range categories {
		_, err := backend.DefineCategory(propID, cat.name, cat.ordinal)
		if err != nil {
			t.Fatalf("DefineCategory(%s) failed: %v", cat.name, err)
		}
	}

	// Get categories and verify ordering
	cats, err := backend.GetCategories(propID)
	if err != nil {
		t.Fatalf("GetCategories() failed: %v", err)
	}

	if len(cats) != 4 {
		t.Fatalf("GetCategories() returned %d categories, want 4", len(cats))
	}

	// Verify ordering by ordinal ascending
	expected := []struct {
		name    string
		ordinal int
	}{
		{"critical", 0},
		{"high", 1},
		{"medium", 2},
		{"low", 3},
	}

	for i, exp := range expected {
		if cats[i].Name != exp.name || cats[i].Ordinal != exp.ordinal {
			t.Errorf("GetCategories()[%d] = {Name: %s, Ordinal: %d}, want {Name: %s, Ordinal: %d}",
				i, cats[i].Name, cats[i].Ordinal, exp.name, exp.ordinal)
		}
	}
}

func TestBackend_GetCategories_EmptySlice(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "crumbs-category-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and attach backend
	backend := NewBackend()
	config := types.Config{
		Backend: "sqlite",
		DataDir: tmpDir,
	}
	if err := backend.Attach(config); err != nil {
		t.Fatalf("Attach failed: %v", err)
	}
	defer backend.Detach()

	// Create a categorical property without any categories
	propsTable, err := backend.GetTable(types.PropertiesTable)
	if err != nil {
		t.Fatalf("GetTable(properties) failed: %v", err)
	}

	prop := &types.Property{
		Name:        "empty_cat",
		Description: "Empty categorical property",
		ValueType:   types.ValueTypeCategorical,
	}
	propID, err := propsTable.Set("", prop)
	if err != nil {
		t.Fatalf("Set property failed: %v", err)
	}

	// Get categories - should return empty slice, not nil
	cats, err := backend.GetCategories(propID)
	if err != nil {
		t.Fatalf("GetCategories() failed: %v", err)
	}

	if cats == nil {
		t.Error("GetCategories() returned nil, want empty slice")
	}
	if len(cats) != 0 {
		t.Errorf("GetCategories() returned %d categories, want 0", len(cats))
	}
}

func TestBackend_DefineCategory_JSONLPersistence(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "crumbs-category-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and attach backend
	backend := NewBackend()
	config := types.Config{
		Backend: "sqlite",
		DataDir: tmpDir,
	}
	if err := backend.Attach(config); err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	// Create a categorical property
	propsTable, err := backend.GetTable(types.PropertiesTable)
	if err != nil {
		t.Fatalf("GetTable(properties) failed: %v", err)
	}

	prop := &types.Property{
		Name:        "jsonl_test",
		Description: "Test JSONL persistence",
		ValueType:   types.ValueTypeCategorical,
	}
	propID, err := propsTable.Set("", prop)
	if err != nil {
		t.Fatalf("Set property failed: %v", err)
	}

	// Define a category
	cat, err := backend.DefineCategory(propID, "persisted", 1)
	if err != nil {
		t.Fatalf("DefineCategory() failed: %v", err)
	}

	// Detach to flush everything
	if err := backend.Detach(); err != nil {
		t.Fatalf("Detach failed: %v", err)
	}

	// Verify JSONL file exists and contains the category
	jsonlPath := filepath.Join(tmpDir, "categories.jsonl")
	content, err := os.ReadFile(jsonlPath)
	if err != nil {
		t.Fatalf("failed to read categories.jsonl: %v", err)
	}

	if len(content) == 0 {
		t.Error("categories.jsonl is empty")
	}

	// Check for expected content
	if string(content) == "" {
		t.Error("categories.jsonl content is empty")
	}

	// Reattach and verify category is loaded from JSONL
	backend2 := NewBackend()
	if err := backend2.Attach(config); err != nil {
		t.Fatalf("Attach (reopen) failed: %v", err)
	}
	defer backend2.Detach()

	cats, err := backend2.GetCategories(propID)
	if err != nil {
		t.Fatalf("GetCategories() after reopen failed: %v", err)
	}

	if len(cats) != 1 {
		t.Fatalf("GetCategories() after reopen returned %d categories, want 1", len(cats))
	}

	if cats[0].CategoryID != cat.CategoryID {
		t.Errorf("CategoryID mismatch after reopen: got %s, want %s", cats[0].CategoryID, cat.CategoryID)
	}
	if cats[0].Name != "persisted" {
		t.Errorf("Category name mismatch after reopen: got %s, want persisted", cats[0].Name)
	}
}

func TestBackend_DefineCategory_Detached(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "crumbs-category-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create backend but don't attach
	backend := NewBackend()

	_, err = backend.DefineCategory("prop-1", "cat", 1)
	if !errors.Is(err, types.ErrCupboardDetached) {
		t.Errorf("DefineCategory() on detached backend error = %v, want ErrCupboardDetached", err)
	}

	_, err = backend.GetCategories("prop-1")
	if !errors.Is(err, types.ErrCupboardDetached) {
		t.Errorf("GetCategories() on detached backend error = %v, want ErrCupboardDetached", err)
	}
}

func TestProperty_DefineCategory_Integration(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "crumbs-category-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and attach backend
	backend := NewBackend()
	config := types.Config{
		Backend: "sqlite",
		DataDir: tmpDir,
	}
	if err := backend.Attach(config); err != nil {
		t.Fatalf("Attach failed: %v", err)
	}
	defer backend.Detach()

	// Create a categorical property
	propsTable, err := backend.GetTable(types.PropertiesTable)
	if err != nil {
		t.Fatalf("GetTable(properties) failed: %v", err)
	}

	prop := &types.Property{
		Name:        "integration_test",
		Description: "Integration test property",
		ValueType:   types.ValueTypeCategorical,
	}
	propID, err := propsTable.Set("", prop)
	if err != nil {
		t.Fatalf("Set property failed: %v", err)
	}
	prop.PropertyID = propID

	// Use the entity method to define a category
	cat, err := prop.DefineCategory(backend, "integration_cat", 5)
	if err != nil {
		t.Fatalf("Property.DefineCategory() failed: %v", err)
	}

	if cat.CategoryID == "" {
		t.Error("DefineCategory() returned category with empty CategoryID")
	}

	// Use the entity method to get categories
	cats, err := prop.GetCategories(backend)
	if err != nil {
		t.Fatalf("Property.GetCategories() failed: %v", err)
	}

	if len(cats) != 1 {
		t.Fatalf("GetCategories() returned %d categories, want 1", len(cats))
	}

	if cats[0].Name != "integration_cat" {
		t.Errorf("GetCategories()[0].Name = %s, want integration_cat", cats[0].Name)
	}
}

func TestProperty_DefineCategory_NonCategorical(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "crumbs-category-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and attach backend
	backend := NewBackend()
	config := types.Config{
		Backend: "sqlite",
		DataDir: tmpDir,
	}
	if err := backend.Attach(config); err != nil {
		t.Fatalf("Attach failed: %v", err)
	}
	defer backend.Detach()

	// Create a text property
	propsTable, err := backend.GetTable(types.PropertiesTable)
	if err != nil {
		t.Fatalf("GetTable(properties) failed: %v", err)
	}

	prop := &types.Property{
		Name:        "text_prop",
		Description: "Text property",
		ValueType:   types.ValueTypeText,
	}
	propID, err := propsTable.Set("", prop)
	if err != nil {
		t.Fatalf("Set property failed: %v", err)
	}
	prop.PropertyID = propID

	// Try to define a category on text property - should fail
	_, err = prop.DefineCategory(backend, "should_fail", 1)
	if !errors.Is(err, types.ErrInvalidValueType) {
		t.Errorf("DefineCategory() on text property error = %v, want ErrInvalidValueType", err)
	}

	// Try to get categories on text property - should fail
	_, err = prop.GetCategories(backend)
	if !errors.Is(err, types.ErrInvalidValueType) {
		t.Errorf("GetCategories() on text property error = %v, want ErrInvalidValueType", err)
	}
}
