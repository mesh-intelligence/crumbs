package types

import (
	"errors"
	"testing"
)

// mockCategoryDefiner implements CategoryDefiner for testing.
type mockCategoryDefiner struct {
	categories map[string][]*Category
	nextID     int
}

func newMockCategoryDefiner() *mockCategoryDefiner {
	return &mockCategoryDefiner{
		categories: make(map[string][]*Category),
	}
}

func (m *mockCategoryDefiner) DefineCategory(propertyID, name string, ordinal int) (*Category, error) {
	for _, cat := range m.categories[propertyID] {
		if cat.Name == name {
			return nil, ErrDuplicateName
		}
	}
	m.nextID++
	cat := &Category{
		CategoryID: "cat-" + string(rune('0'+m.nextID)),
		PropertyID: propertyID,
		Name:       name,
		Ordinal:    ordinal,
	}
	m.categories[propertyID] = append(m.categories[propertyID], cat)
	return cat, nil
}

func (m *mockCategoryDefiner) GetCategories(propertyID string) ([]*Category, error) {
	cats := m.categories[propertyID]
	if cats == nil {
		return []*Category{}, nil
	}
	result := make([]*Category, len(cats))
	copy(result, cats)
	return result, nil
}

func TestProperty_DefineCategory(t *testing.T) {
	tests := []struct {
		name      string
		property  *Property
		catName   string
		ordinal   int
		wantErr   error
		wantCatID bool
	}{
		{
			name: "categorical property creates category",
			property: &Property{
				PropertyID: "prop-1",
				Name:       "priority",
				ValueType:  ValueTypeCategorical,
			},
			catName:   "high",
			ordinal:   1,
			wantErr:   nil,
			wantCatID: true,
		},
		{
			name: "text property returns ErrInvalidValueType",
			property: &Property{
				PropertyID: "prop-2",
				Name:       "description",
				ValueType:  ValueTypeText,
			},
			catName: "value",
			ordinal: 1,
			wantErr: ErrInvalidValueType,
		},
		{
			name: "integer property returns ErrInvalidValueType",
			property: &Property{
				PropertyID: "prop-3",
				Name:       "estimate",
				ValueType:  ValueTypeInteger,
			},
			catName: "value",
			ordinal: 1,
			wantErr: ErrInvalidValueType,
		},
		{
			name: "empty name returns ErrInvalidName",
			property: &Property{
				PropertyID: "prop-4",
				Name:       "priority",
				ValueType:  ValueTypeCategorical,
			},
			catName: "",
			ordinal: 1,
			wantErr: ErrInvalidName,
		},
		{
			name: "negative ordinal is allowed",
			property: &Property{
				PropertyID: "prop-5",
				Name:       "priority",
				ValueType:  ValueTypeCategorical,
			},
			catName:   "top",
			ordinal:   -10,
			wantCatID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definer := newMockCategoryDefiner()
			cat, err := tt.property.DefineCategory(definer, tt.catName, tt.ordinal)

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

			if tt.wantCatID && cat.CategoryID == "" {
				t.Error("DefineCategory() returned category with empty CategoryID")
			}
			if cat.PropertyID != tt.property.PropertyID {
				t.Errorf("DefineCategory() PropertyID = %v, want %v", cat.PropertyID, tt.property.PropertyID)
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

func TestProperty_DefineCategory_Duplicate(t *testing.T) {
	property := &Property{
		PropertyID: "prop-dup",
		Name:       "priority",
		ValueType:  ValueTypeCategorical,
	}
	definer := newMockCategoryDefiner()

	// First definition should succeed
	_, err := property.DefineCategory(definer, "high", 1)
	if err != nil {
		t.Fatalf("First DefineCategory() unexpected error = %v", err)
	}

	// Second definition with same name should fail
	_, err = property.DefineCategory(definer, "high", 2)
	if !errors.Is(err, ErrDuplicateName) {
		t.Errorf("DefineCategory() duplicate error = %v, want ErrDuplicateName", err)
	}
}

func TestProperty_GetCategories(t *testing.T) {
	tests := []struct {
		name     string
		property *Property
		setup    func(*mockCategoryDefiner)
		wantErr  error
		wantLen  int
	}{
		{
			name: "categorical property returns empty slice",
			property: &Property{
				PropertyID: "prop-empty",
				Name:       "priority",
				ValueType:  ValueTypeCategorical,
			},
			setup:   func(m *mockCategoryDefiner) {},
			wantLen: 0,
		},
		{
			name: "categorical property returns categories",
			property: &Property{
				PropertyID: "prop-cats",
				Name:       "priority",
				ValueType:  ValueTypeCategorical,
			},
			setup: func(m *mockCategoryDefiner) {
				m.categories["prop-cats"] = []*Category{
					{CategoryID: "cat-1", PropertyID: "prop-cats", Name: "high", Ordinal: 1},
					{CategoryID: "cat-2", PropertyID: "prop-cats", Name: "low", Ordinal: 2},
				}
			},
			wantLen: 2,
		},
		{
			name: "text property returns ErrInvalidValueType",
			property: &Property{
				PropertyID: "prop-text",
				Name:       "description",
				ValueType:  ValueTypeText,
			},
			setup:   func(m *mockCategoryDefiner) {},
			wantErr: ErrInvalidValueType,
		},
		{
			name: "list property returns ErrInvalidValueType",
			property: &Property{
				PropertyID: "prop-list",
				Name:       "labels",
				ValueType:  ValueTypeList,
			},
			setup:   func(m *mockCategoryDefiner) {},
			wantErr: ErrInvalidValueType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definer := newMockCategoryDefiner()
			tt.setup(definer)

			cats, err := tt.property.GetCategories(definer)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetCategories() error = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetCategories() unexpected error = %v", err)
				return
			}

			if len(cats) != tt.wantLen {
				t.Errorf("GetCategories() returned %d categories, want %d", len(cats), tt.wantLen)
			}
		})
	}
}

func TestProperty_GetCategories_Ordering(t *testing.T) {
	property := &Property{
		PropertyID: "prop-order",
		Name:       "priority",
		ValueType:  ValueTypeCategorical,
	}
	definer := newMockCategoryDefiner()

	// Add categories in non-sorted order
	definer.categories["prop-order"] = []*Category{
		{CategoryID: "cat-3", PropertyID: "prop-order", Name: "low", Ordinal: 3},
		{CategoryID: "cat-1", PropertyID: "prop-order", Name: "high", Ordinal: 1},
		{CategoryID: "cat-2", PropertyID: "prop-order", Name: "medium", Ordinal: 2},
		{CategoryID: "cat-0", PropertyID: "prop-order", Name: "critical", Ordinal: 0},
	}

	cats, err := property.GetCategories(definer)
	if err != nil {
		t.Fatalf("GetCategories() unexpected error = %v", err)
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

func TestProperty_GetCategories_OrderingByName(t *testing.T) {
	property := &Property{
		PropertyID: "prop-name-order",
		Name:       "status",
		ValueType:  ValueTypeCategorical,
	}
	definer := newMockCategoryDefiner()

	// Add categories with same ordinal
	definer.categories["prop-name-order"] = []*Category{
		{CategoryID: "cat-z", PropertyID: "prop-name-order", Name: "zebra", Ordinal: 1},
		{CategoryID: "cat-a", PropertyID: "prop-name-order", Name: "alpha", Ordinal: 1},
		{CategoryID: "cat-b", PropertyID: "prop-name-order", Name: "beta", Ordinal: 1},
	}

	cats, err := property.GetCategories(definer)
	if err != nil {
		t.Fatalf("GetCategories() unexpected error = %v", err)
	}

	// Verify ordering by name ascending for same ordinal
	expected := []string{"alpha", "beta", "zebra"}

	for i, exp := range expected {
		if cats[i].Name != exp {
			t.Errorf("GetCategories()[%d].Name = %s, want %s", i, cats[i].Name, exp)
		}
	}
}
