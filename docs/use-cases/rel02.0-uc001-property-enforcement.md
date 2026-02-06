# Use Case: Property Enforcement

## Summary

A developer defines custom properties, verifies built-in properties are seeded, and validates that all crumbs automatically have values for all defined properties. This tracer bullet validates the property enforcement mechanism: auto-initialization on crumb creation and backfill on property definition.

## Actor and Trigger

The actor is a developer or automated test harness. The trigger is the need to validate that the property system enforces the invariant: every crumb has a value for every defined property, with no gaps or partial state.

## Flow

1. **Create cupboard with seeded properties**: Construct a Cupboard and call `Attach(config)` with a SQLite backend. The backend seeds built-in properties (priority, type, description, owner, labels) during initialization.

2. **Verify built-in properties**: Call `cupboard.GetTable("properties")` and use `table.Fetch(nil)` to list all properties. Confirm the five built-in properties exist with correct value types.

```go
propsTable, _ := cupboard.GetTable("properties")
entities, _ := propsTable.Fetch(nil)
// Should return 5 properties: priority, type, description, owner, labels
```

3. **Add a crumb**: Get the crumbs table, construct a Crumb, and call `table.Set("", crumb)`. The operation creates the crumb and auto-initializes all five built-in properties with their default values.

```go
crumbsTable, _ := cupboard.GetTable("crumbs")
crumb := &Crumb{Name: "Implement feature X"}
id, _ := crumbsTable.Set("", crumb)
```

4. **Verify property initialization**: Call `crumb.GetProperties()` to get all properties. Confirm it returns all five properties with default values (empty strings for text, empty lists for list types, null for categorical).

```go
entity, _ := crumbsTable.Get(id)
crumb := entity.(*Crumb)
props := crumb.GetProperties()  // map[string]any
// Should have entries for all 5 defined properties
```

5. **Set a property value**: Call `crumb.SetProperty(propertyID, value)` then persist with `table.Set`. Verify UpdatedAt changes.

```go
crumb.SetProperty(priorityPropID, highCategoryID)
_, _ = crumbsTable.Set(crumb.CrumbID, crumb)
```

6. **Get a property value**: Call `crumb.GetProperty(propertyID)`. Verify it returns the value set in the previous step.

```go
entity, _ := crumbsTable.Get(id)
crumb := entity.(*Crumb)
value, _ := crumb.GetProperty(priorityPropID)
// Should return highCategoryID
```

7. **Define a custom property**: Construct a Property and call `propsTable.Set("", prop)`. The operation creates the property and backfills all existing crumbs with the default value (0 for integer).

```go
prop := &Property{
    Name:        "estimate",
    Description: "Story point estimate",
    ValueType:   "integer",
}
propID, _ := propsTable.Set("", prop)
```

8. **Verify backfill occurred**: Retrieve the crumb and call `crumb.GetProperties()`. Confirm the crumb now has six properties including "estimate" with value 0.

```go
entity, _ := crumbsTable.Get(id)
crumb := entity.(*Crumb)
props := crumb.GetProperties()
// Should have 6 entries including estimate with value 0
```

9. **Add another crumb after property definition**: Create another crumb. Verify the new crumb has all six properties (five built-in plus estimate) auto-initialized.

```go
crumb2 := &Crumb{Name: "Another task"}
id2, _ := crumbsTable.Set("", crumb2)
entity2, _ := crumbsTable.Get(id2)
crumb2 = entity2.(*Crumb)
props2 := crumb2.GetProperties()
// Should have 6 entries
```

10. **Update custom property**: Call `crumb.SetProperty(propertyID, value)` then persist to set the estimate to 5.

```go
crumb.SetProperty(propID, int64(5))
_, _ = crumbsTable.Set(crumb.CrumbID, crumb)
```

11. **Clear property to default**: Call `crumb.ClearProperty(propertyID)` then persist. Verify the estimate resets to 0 (the default), not null or unset.

```go
crumb.ClearProperty(propID)
_, _ = crumbsTable.Set(crumb.CrumbID, crumb)
value, _ := crumb.GetProperty(propID)
// Should return 0
```

12. **Verify invariant holds**: Retrieve both crumbs and call `GetProperties()`. Confirm each has exactly six properties with no gaps.

13. **Detach the cupboard**: Call `cupboard.Detach()`.

## Architecture Touchpoints

This use case exercises the following interfaces and components:

| Interface | Operations Used |
|-----------|-----------------|
| Cupboard | Attach, Detach, GetTable |
| Table (crumbs) | Get, Set, Fetch |
| Table (properties) | Get, Set, Fetch |
| Crumb entity | SetProperty, GetProperty, GetProperties, ClearProperty |
| Property entity | (struct fields only in this use case) |

We validate:

- Built-in property seeding (prd-properties-interface R9, prd-sqlite-backend R9)
- Crumb creation with property auto-initialization (prd-crumbs-interface R3)
- Property definition with backfill to existing crumbs (prd-properties-interface R4)
- SetProperty updates value and timestamp (prd-crumbs-interface R5)
- GetProperty retrieves current value (prd-crumbs-interface R5)
- ClearProperty resets to default, not null (prd-crumbs-interface R5)
- Invariant: every crumb has every property (no partial state)

## Success Criteria

The demo succeeds when:

- [ ] properties table Fetch returns five built-in properties after initialization
- [ ] Newly added crumbs have all defined properties with default values
- [ ] SetProperty updates value and changes UpdatedAt
- [ ] GetProperty returns the current value
- [ ] Creating a Property via Table.Set backfills existing crumbs
- [ ] GetProperties() returns all properties (never partial) for any crumb
- [ ] ClearProperty resets value to default, not null or unset
- [ ] Crumbs added after property definition have the new property auto-initialized
- [ ] No crumb ever has fewer properties than are defined

Observable demo script:

```bash
# Run the demo binary or test
go test -v ./internal/sqlite -run TestPropertyEnforcement

# Or run a CLI demo
crumbs demo properties --datadir /tmp/crumbs-demo
```

## Out of Scope

This use case does not cover:

- Core CRUD operations (Get, Set, Delete, Fetch) - see rel01.0-uc003
- Category management (DefineCategory, GetCategories entity methods)
- Property deletion or renaming (not supported per prd-properties-interface)
- Trail operations - see rel03.0-uc001
- Concurrent property definition and crumb creation
- Property value validation beyond type checking

## Dependencies

- rel01.0-uc001 (Cupboard lifecycle) must pass
- rel01.0-uc003 (Core CRUD) must pass
- prd-properties-interface must be implemented (Property entity, Table operations)
- prd-crumbs-interface property methods must be implemented

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Backfill performance on large datasets | Test with 1000+ crumbs to validate acceptable latency |
| Atomic transaction failures | Verify rollback behavior when backfill fails mid-operation |
| Default value edge cases (null vs zero) | Explicit test cases for each value type's default |
| Race condition on Define + Add | Document that Define should complete before concurrent Adds |
