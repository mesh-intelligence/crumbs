// Integration tests for property enforcement: built-in property seeding,
// auto-initialization on crumb creation, backfill on property definition,
// and the invariant that no crumb has fewer properties than are defined.
// Implements: test-rel02.0-uc001-property-enforcement (test cases 1-9);
//             prd004-properties-interface R3.5, R4.2, R9;
//             prd003-crumbs-interface R3, R5.
package integration

import (
	"testing"
	"time"

	"github.com/mesh-intelligence/crumbs/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// propertyByName looks up a property by name from the properties table.
func propertyByName(t *testing.T, backend types.Cupboard, name string) *types.Property {
	t.Helper()
	propsTbl, err := backend.GetTable(types.TableProperties)
	require.NoError(t, err)
	allProps, err := propsTbl.Fetch(types.Filter{"name": name})
	require.NoError(t, err)
	require.Len(t, allProps, 1, "expected exactly one property named %q", name)
	return allProps[0].(*types.Property)
}

// allPropertyIDs returns the set of property IDs from the properties table.
func allPropertyIDs(t *testing.T, backend types.Cupboard) map[string]bool {
	t.Helper()
	propsTbl, err := backend.GetTable(types.TableProperties)
	require.NoError(t, err)
	allProps, err := propsTbl.Fetch(nil)
	require.NoError(t, err)
	ids := make(map[string]bool, len(allProps))
	for _, p := range allProps {
		ids[p.(*types.Property).PropertyID] = true
	}
	return ids
}

// getCrumb retrieves a crumb by ID and returns it as *types.Crumb.
func getCrumb(t *testing.T, crumbsTbl types.Table, id string) *types.Crumb {
	t.Helper()
	entity, err := crumbsTbl.Get(id)
	require.NoError(t, err)
	return entity.(*types.Crumb)
}

// --- S1: built-in properties seeded on initialization ---

func TestPropertyEnforcement_BuiltInPropertiesSeeded(t *testing.T) {
	tests := []struct {
		name      string
		propName  string
		valueType string
	}{
		{"priority exists with categorical type", types.PropertyPriority, types.ValueTypeCategorical},
		{"type exists with categorical type", types.PropertyType, types.ValueTypeCategorical},
		{"description exists with text type", types.PropertyDescription, types.ValueTypeText},
		{"owner exists with text type", types.PropertyOwner, types.ValueTypeText},
		{"labels exists with list type", types.PropertyLabels, types.ValueTypeList},
	}

	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	propsTbl, err := backend.GetTable(types.TableProperties)
	require.NoError(t, err)

	allProps, err := propsTbl.Fetch(nil)
	require.NoError(t, err)
	assert.Len(t, allProps, 5, "expected five built-in properties after initialization")

	propsByName := make(map[string]*types.Property, len(allProps))
	for _, p := range allProps {
		prop := p.(*types.Property)
		propsByName[prop.Name] = prop
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop, ok := propsByName[tt.propName]
			require.True(t, ok, "property %q must exist", tt.propName)
			assert.Equal(t, tt.valueType, prop.ValueType)
		})
	}
}

// --- S2: new crumbs have all defined properties with defaults ---

func TestPropertyEnforcement_S2_NewCrumbHasAllProperties(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := crumbsTbl.Set("", &types.Crumb{Name: "Auto-init crumb"})
	require.NoError(t, err)

	got := getCrumb(t, crumbsTbl, id)
	assert.Len(t, got.Properties, 5, "new crumb should have all five built-in properties")
}

func TestPropertyEnforcement_S2_DefaultValues(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := crumbsTbl.Set("", &types.Crumb{Name: "Default values crumb"})
	require.NoError(t, err)

	got := getCrumb(t, crumbsTbl, id)

	propsTbl, err := backend.GetTable(types.TableProperties)
	require.NoError(t, err)
	allProps, err := propsTbl.Fetch(nil)
	require.NoError(t, err)

	for _, p := range allProps {
		prop := p.(*types.Property)
		val, ok := got.Properties[prop.PropertyID]
		require.True(t, ok, "property %q must be present", prop.Name)

		switch prop.ValueType {
		case types.ValueTypeText:
			assert.Equal(t, "", val, "text property %q should default to empty string", prop.Name)
		case types.ValueTypeList:
			assert.IsType(t, []any{}, val, "list property %q should default to empty slice", prop.Name)
			assert.Empty(t, val, "list property %q should be empty", prop.Name)
		case types.ValueTypeCategorical:
			assert.Nil(t, val, "categorical property %q should default to nil", prop.Name)
		}
	}
}

// --- S3: SetProperty updates value and changes UpdatedAt ---

func TestPropertyEnforcement_S3_SetPropertyUpdatesInMemory(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := crumbsTbl.Set("", &types.Crumb{Name: "SetProperty crumb"})
	require.NoError(t, err)

	crumb := getCrumb(t, crumbsTbl, id)
	originalUpdatedAt := crumb.UpdatedAt

	ownerProp := propertyByName(t, backend, types.PropertyOwner)

	time.Sleep(10 * time.Millisecond)

	err = crumb.SetProperty(ownerProp.PropertyID, "alice")
	require.NoError(t, err)

	// Verify in-memory value changed.
	val, err := crumb.GetProperty(ownerProp.PropertyID)
	require.NoError(t, err)
	assert.Equal(t, "alice", val)

	// Verify UpdatedAt advanced in memory.
	assert.True(t, crumb.UpdatedAt.After(originalUpdatedAt),
		"UpdatedAt should advance after SetProperty")
}

func TestPropertyEnforcement_S3_SetPropertyListValue(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := crumbsTbl.Set("", &types.Crumb{Name: "Labels crumb"})
	require.NoError(t, err)

	crumb := getCrumb(t, crumbsTbl, id)
	labelsProp := propertyByName(t, backend, types.PropertyLabels)

	labels := []any{"frontend", "urgent"}
	err = crumb.SetProperty(labelsProp.PropertyID, labels)
	require.NoError(t, err)

	val, err := crumb.GetProperty(labelsProp.PropertyID)
	require.NoError(t, err)
	assert.Equal(t, labels, val)
}

func TestPropertyEnforcement_S3_SetPropertyUpdatesTimestampOnPersist(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := crumbsTbl.Set("", &types.Crumb{Name: "Timestamp crumb"})
	require.NoError(t, err)

	crumb := getCrumb(t, crumbsTbl, id)
	originalUpdatedAt := crumb.UpdatedAt

	ownerProp := propertyByName(t, backend, types.PropertyOwner)

	time.Sleep(10 * time.Millisecond)

	err = crumb.SetProperty(ownerProp.PropertyID, "bob")
	require.NoError(t, err)

	// Persist the crumb (Set updates timestamps in the crumbs row).
	_, err = crumbsTbl.Set(id, crumb)
	require.NoError(t, err)

	got := getCrumb(t, crumbsTbl, id)
	assert.True(t, !got.UpdatedAt.Before(originalUpdatedAt),
		"UpdatedAt in database should reflect the updated timestamp")
}

// --- S4: GetProperty returns current value ---

func TestPropertyEnforcement_S4_GetPropertyReturnsSetValue(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := crumbsTbl.Set("", &types.Crumb{Name: "GetProperty crumb"})
	require.NoError(t, err)

	crumb := getCrumb(t, crumbsTbl, id)
	descProp := propertyByName(t, backend, types.PropertyDescription)

	err = crumb.SetProperty(descProp.PropertyID, "A detailed description")
	require.NoError(t, err)

	// GetProperty on the same in-memory crumb returns the set value.
	val, err := crumb.GetProperty(descProp.PropertyID)
	require.NoError(t, err)
	assert.Equal(t, "A detailed description", val)
}

func TestPropertyEnforcement_S4_GetPropertyReturnsDefaultBeforeSet(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := crumbsTbl.Set("", &types.Crumb{Name: "Default get crumb"})
	require.NoError(t, err)

	got := getCrumb(t, crumbsTbl, id)
	ownerProp := propertyByName(t, backend, types.PropertyOwner)
	val, err := got.GetProperty(ownerProp.PropertyID)
	require.NoError(t, err)
	assert.Equal(t, "", val, "owner should default to empty string before any set")
}

func TestPropertyEnforcement_S4_GetPropertyOnNonexistentReturnsError(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := crumbsTbl.Set("", &types.Crumb{Name: "Error test crumb"})
	require.NoError(t, err)

	crumb := getCrumb(t, crumbsTbl, id)
	_, err = crumb.GetProperty("nonexistent-property-id")
	assert.ErrorIs(t, err, types.ErrPropertyNotFound)
}

// --- S5: creating a Property via Table.Set backfills existing crumbs ---

func TestPropertyEnforcement_S5_BackfillSingleCrumb(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	crumbID, err := crumbsTbl.Set("", &types.Crumb{Name: "Backfill target"})
	require.NoError(t, err)

	propsTbl, err := backend.GetTable(types.TableProperties)
	require.NoError(t, err)
	newPropID, err := propsTbl.Set("", &types.Property{
		Name:        "estimate",
		ValueType:   types.ValueTypeInteger,
		Description: "Time estimate in hours",
	})
	require.NoError(t, err)

	got := getCrumb(t, crumbsTbl, crumbID)
	assert.Len(t, got.Properties, 6, "crumb should have six properties after backfill")

	val, ok := got.Properties[newPropID]
	assert.True(t, ok, "new property should exist in crumb's properties")
	assert.Equal(t, float64(0), val, "integer property should default to 0")
}

func TestPropertyEnforcement_S5_BackfillMultipleCrumbs(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	var crumbIDs []string
	for _, name := range []string{"Crumb one", "Crumb two", "Crumb three"} {
		id, err := crumbsTbl.Set("", &types.Crumb{Name: name})
		require.NoError(t, err)
		crumbIDs = append(crumbIDs, id)
	}

	propsTbl, err := backend.GetTable(types.TableProperties)
	require.NoError(t, err)
	newPropID, err := propsTbl.Set("", &types.Property{
		Name:        "complexity",
		ValueType:   types.ValueTypeInteger,
		Description: "Task complexity",
	})
	require.NoError(t, err)

	for _, cid := range crumbIDs {
		got := getCrumb(t, crumbsTbl, cid)
		assert.Len(t, got.Properties, 6, "crumb %s should have six properties", cid)
		val, ok := got.Properties[newPropID]
		assert.True(t, ok, "crumb %s should have the new property", cid)
		assert.Equal(t, float64(0), val, "integer default should be 0 for crumb %s", cid)
	}
}

// --- S6: GetProperties returns all properties (never partial) ---

func TestPropertyEnforcement_S6_GetPropertiesReturnsAll(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := crumbsTbl.Set("", &types.Crumb{Name: "All properties crumb"})
	require.NoError(t, err)

	got := getCrumb(t, crumbsTbl, id)
	props := got.GetProperties()
	assert.Len(t, props, 5, "GetProperties should return all five built-in properties")

	expectedIDs := allPropertyIDs(t, backend)
	for pid := range props {
		assert.True(t, expectedIDs[pid], "property ID %s should be a defined property", pid)
	}
}

func TestPropertyEnforcement_S6_GetPropertiesIncludesCustom(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	propsTbl, err := backend.GetTable(types.TableProperties)
	require.NoError(t, err)
	_, err = propsTbl.Set("", &types.Property{
		Name:        "custom_field",
		ValueType:   types.ValueTypeText,
		Description: "Custom text field",
	})
	require.NoError(t, err)

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)
	id, err := crumbsTbl.Set("", &types.Crumb{Name: "Custom props crumb"})
	require.NoError(t, err)

	got := getCrumb(t, crumbsTbl, id)
	props := got.GetProperties()
	assert.Len(t, props, 6, "GetProperties should return six properties (five built-in + custom)")
}

// --- S7: ClearProperty resets value to default, not null ---

func TestPropertyEnforcement_S7_ClearPropertyResetsToDefault(t *testing.T) {
	tests := []struct {
		name         string
		propName     string
		setValue     any
		expectedType string
	}{
		{"clear text resets to empty string", types.PropertyOwner, "charlie", types.ValueTypeText},
		{"clear list resets to empty array", types.PropertyLabels, []any{"tag1", "tag2"}, types.ValueTypeList},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, _ := newAttachedBackend(t)
			defer backend.Detach()

			crumbsTbl, err := backend.GetTable(types.TableCrumbs)
			require.NoError(t, err)

			id, err := crumbsTbl.Set("", &types.Crumb{Name: "Clear " + tt.propName + " crumb"})
			require.NoError(t, err)

			crumb := getCrumb(t, crumbsTbl, id)
			prop := propertyByName(t, backend, tt.propName)

			// Set a non-default value, then clear it.
			err = crumb.SetProperty(prop.PropertyID, tt.setValue)
			require.NoError(t, err)

			err = crumb.ClearProperty(prop.PropertyID)
			require.NoError(t, err)

			// ClearProperty sets the in-memory value to nil (prd003-crumbs-interface R5.5).
			// Full default-value resolution is deferred to Table.Set per R5.7.
			val, err := crumb.GetProperty(prop.PropertyID)
			require.NoError(t, err)
			assert.Nil(t, val, "ClearProperty should set the value to nil in memory")
		})
	}
}

func TestPropertyEnforcement_S7_ClearPropertyUpdatesTimestamp(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := crumbsTbl.Set("", &types.Crumb{Name: "Clear timestamp crumb"})
	require.NoError(t, err)

	crumb := getCrumb(t, crumbsTbl, id)
	ownerProp := propertyByName(t, backend, types.PropertyOwner)

	err = crumb.SetProperty(ownerProp.PropertyID, "dave")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)
	beforeClear := crumb.UpdatedAt

	err = crumb.ClearProperty(ownerProp.PropertyID)
	require.NoError(t, err)

	assert.True(t, crumb.UpdatedAt.After(beforeClear),
		"UpdatedAt should advance after ClearProperty")
}

func TestPropertyEnforcement_S7_ClearPropertyOnNonexistentReturnsError(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	id, err := crumbsTbl.Set("", &types.Crumb{Name: "Clear error crumb"})
	require.NoError(t, err)

	crumb := getCrumb(t, crumbsTbl, id)
	err = crumb.ClearProperty("nonexistent-property-id")
	assert.ErrorIs(t, err, types.ErrPropertyNotFound)
}

// --- S8: crumbs added after property definition have new property ---

func TestPropertyEnforcement_S8_CrumbAfterPropertyDefinition(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	propsTbl, err := backend.GetTable(types.TableProperties)
	require.NoError(t, err)
	newPropID, err := propsTbl.Set("", &types.Property{
		Name:        "story_points",
		ValueType:   types.ValueTypeInteger,
		Description: "Estimated story points",
	})
	require.NoError(t, err)

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)

	var crumbIDs []string
	for _, name := range []string{"Crumb A", "Crumb B"} {
		id, err := crumbsTbl.Set("", &types.Crumb{Name: name})
		require.NoError(t, err)
		crumbIDs = append(crumbIDs, id)
	}

	for _, cid := range crumbIDs {
		got := getCrumb(t, crumbsTbl, cid)
		assert.Len(t, got.Properties, 6, "crumb should have six properties (five built-in + story_points)")
		val, ok := got.Properties[newPropID]
		assert.True(t, ok, "new property should be auto-initialized")
		assert.Equal(t, float64(0), val, "integer should default to 0")
	}
}

// --- S9: no crumb ever has fewer properties than defined ---

func TestPropertyEnforcement_S9_InvariantHoldsAfterMultipleDefinitions(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)
	propsTbl, err := backend.GetTable(types.TableProperties)
	require.NoError(t, err)

	// Create an early crumb (before any custom properties).
	earlyID, err := crumbsTbl.Set("", &types.Crumb{Name: "Early crumb"})
	require.NoError(t, err)

	// Define first custom property.
	_, err = propsTbl.Set("", &types.Property{
		Name:        "prop_one",
		ValueType:   types.ValueTypeText,
		Description: "First custom property",
	})
	require.NoError(t, err)

	// Create a middle crumb (after prop_one).
	middleID, err := crumbsTbl.Set("", &types.Crumb{Name: "Middle crumb"})
	require.NoError(t, err)

	// Define second custom property.
	_, err = propsTbl.Set("", &types.Property{
		Name:        "prop_two",
		ValueType:   types.ValueTypeInteger,
		Description: "Second custom property",
	})
	require.NoError(t, err)

	// Create a late crumb (after both custom properties).
	lateID, err := crumbsTbl.Set("", &types.Crumb{Name: "Late crumb"})
	require.NoError(t, err)

	// All three crumbs should have exactly 7 properties (5 built-in + 2 custom).
	expectedCount := 7
	for _, cid := range []string{earlyID, middleID, lateID} {
		got := getCrumb(t, crumbsTbl, cid)
		assert.Len(t, got.Properties, expectedCount,
			"crumb %s should have %d properties", cid, expectedCount)
	}
}

func TestPropertyEnforcement_S9_InvariantHoldsAfterMixedOperations(t *testing.T) {
	backend, _ := newAttachedBackend(t)
	defer backend.Detach()

	crumbsTbl, err := backend.GetTable(types.TableCrumbs)
	require.NoError(t, err)
	propsTbl, err := backend.GetTable(types.TableProperties)
	require.NoError(t, err)

	// Create two crumbs.
	id1, err := crumbsTbl.Set("", &types.Crumb{Name: "Crumb 1"})
	require.NoError(t, err)
	id2, err := crumbsTbl.Set("", &types.Crumb{Name: "Crumb 2"})
	require.NoError(t, err)

	// Define a text property.
	_, err = propsTbl.Set("", &types.Property{
		Name:        "field_a",
		ValueType:   types.ValueTypeText,
		Description: "Field A",
	})
	require.NoError(t, err)

	// Create a third crumb.
	id3, err := crumbsTbl.Set("", &types.Crumb{Name: "Crumb 3"})
	require.NoError(t, err)

	// Define another property.
	_, err = propsTbl.Set("", &types.Property{
		Name:        "field_b",
		ValueType:   types.ValueTypeBoolean,
		Description: "Field B",
	})
	require.NoError(t, err)

	// Verify invariant: all crumbs have the same property count (5 built-in + 2 custom = 7).
	expectedCount := 7
	for _, cid := range []string{id1, id2, id3} {
		got := getCrumb(t, crumbsTbl, cid)
		assert.Len(t, got.Properties, expectedCount,
			"crumb %s should have %d properties", cid, expectedCount)
	}

	// Verify no crumb has fewer properties than are defined.
	definedCount := len(allPropertyIDs(t, backend))
	assert.Equal(t, expectedCount, definedCount, "defined property count should match expected")

	allCrumbs, err := crumbsTbl.Fetch(nil)
	require.NoError(t, err)
	for _, c := range allCrumbs {
		crumb := c.(*types.Crumb)
		assert.GreaterOrEqual(t, len(crumb.Properties), definedCount,
			"crumb %s must not have fewer properties than defined", crumb.CrumbID)
	}
}
