# Use Case: CRUD Operations Demo

## Summary

A developer creates a SQLite-backed cupboard, adds crumbs with various states, manages custom properties, and cleans up the database. This tracer bullet validates the core CRUD operations and property enforcement across the Cupboard, CrumbTable, and PropertyTable interfaces.

## Actor and Trigger

The actor is a developer or automated test harness. The trigger is the need to validate that the crumbs system correctly handles the full lifecycle of crumbs and properties, including the enforcement that all crumbs have values for all defined properties.

## Flow
<!-- Isnt there a make functon on GO that does somethong like this? We need an interface to Cubboard and we call the Open method on it.-->

1. **Create the database**: Call `OpenCupboard` with a SQLite backend configuration specifying a DataDir. The backend creates the directory, initializes empty JSON files, creates the SQLite schema, and seeds built-in properties (priority, type, description, owner, labels, dependencies).

2. **Verify built-in properties**: Call `Properties().List()` to confirm the six built-in properties exist with correct value types and categories.

3. **Add first crumb**: Call `Crumbs().Add("Implement login feature")`. The operation generates a UUID v7, sets state to "draft", initializes CreatedAt and UpdatedAt, and auto-initializes all six built-in properties with their default values.

4. **Verify property initialization**: Call `Crumbs().GetProperties(crumbID)` on the new crumb. Confirm it returns all six properties with default values (empty strings for text, empty lists for list types, null for categorical).

5. **Change crumb state**: Call `Crumbs().Get(crumbID)` to retrieve the crumb, then use a state-change operation (or SetProperty if state is a property) to transition from "draft" to "ready".

6. **Add second crumb**: Call `Crumbs().Add("Fix authentication bug")`. Verify it also has all properties initialized.

7. **Set property values**: Call `Crumbs().SetProperty(crumbID, priorityPropertyID, highCategoryID)` to set priority on the first crumb. Verify UpdatedAt changes.

8. **Define a custom property**: Call `Properties().Define("estimate", "Story point estimate", "integer")`. The operation creates the property and backfills all existing crumbs (both crumbs from steps 3 and 6) with the default value (0 for integer).

9. **Verify backfill**: Call `Crumbs().GetProperties(crumbID)` on both crumbs. Confirm each now has seven properties including "estimate" with value 0.

10. **Update custom property**: Call `Crumbs().SetProperty(crumbID, estimatePropertyID, 5)` on the first crumb to set the estimate to 5.

11. **Clear property to default**: Call `Crumbs().ClearProperty(crumbID, estimatePropertyID)` on the first crumb. Verify the estimate resets to 0 (not deleted).

12. **Archive a crumb**: Call `Crumbs().Archive(secondCrumbID)`. Verify the crumb's state becomes "archived" and UpdatedAt changes.

13. **Fetch with filter**: Call `Crumbs().Fetch({"states": ["ready"]})`. Verify only the first crumb (state "ready") is returned, not the archived one.

14. **Purge archived crumb**: Call `Crumbs().Purge(secondCrumbID)`. Verify the crumb, its property values, and any links are removed.

15. **Close the cupboard**: Call `Close()`. Verify all resources are released and subsequent operations return ErrCupboardClosed.

16. **Delete the database**: Remove the DataDir to clean up. Verify the directory and all JSON files are gone.

## Architecture Touchpoints

This use case exercises the following interfaces and components:

| Interface | Operations Used |
|-----------|-----------------|
| Cupboard | OpenCupboard, Close |
| CrumbTable | Add, Get, Archive, Purge, Fetch, SetProperty, GetProperties, ClearProperty |
| PropertyTable | Define, List |

The use case validates:

- SQLite backend initialization and JSON file creation (prd-sqlite-backend R1, R4)
- Built-in property seeding (prd-properties-interface R8, prd-sqlite-backend R9)
- Crumb creation with property auto-initialization (prd-crumbs-interface R3.7)
- Property definition with backfill to existing crumbs (prd-properties-interface R4.9)
- ClearProperty resets to default (prd-crumbs-interface R12.2)
- State transitions and archive behavior (prd-crumbs-interface R5)
- Filter-based queries (prd-crumbs-interface R7, R8)
- Purge cascade behavior (prd-crumbs-interface R6)
- Cupboard lifecycle and ErrCupboardClosed (prd-cupboard-core R4, R5)

## Success Criteria

The demo succeeds when:

- [ ] OpenCupboard creates DataDir with all JSON files and cupboard.db
- [ ] Properties().List() returns six built-in properties after initialization
- [ ] Newly added crumbs have all defined properties with default values
- [ ] Properties().Define() backfills existing crumbs with the new property
- [ ] GetProperties() returns all properties (never partial) for any crumb
- [ ] ClearProperty() resets value to default, not null or unset
- [ ] Archive() changes state to "archived" without deleting data
- [ ] Fetch() correctly filters by state
- [ ] Purge() removes crumb and all associated property values
- [ ] Close() prevents further operations with ErrCupboardClosed
- [ ] DataDir removal cleans up all files

Observable demo script:

```bash
# Run the demo binary or test
go test -v ./internal/sqlite -run TestCRUDOperationsDemo

# Or run a CLI demo
crumbs demo --datadir /tmp/crumbs-demo
```

## Out of Scope

This use case does not cover:

- Trail operations (Start, Complete, Abandon, GetCrumbs)
- Metadata operations (Register, Add, Get, Search)
- Link operations beyond what Purge removes
- Concurrent access patterns
- Error recovery scenarios (corrupt JSON, I/O failures)
- Dolt or DynamoDB backends
- Category management beyond built-in categories

## Dependencies

- prd-cupboard-core must be implemented (OpenCupboard, Close)
- prd-crumbs-interface must be implemented (CrumbTable operations)
- prd-properties-interface must be implemented (PropertyTable operations)
- prd-sqlite-backend must be implemented (JSON persistence, schema)

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Backfill performance on large datasets | Test with 1000+ crumbs to validate acceptable latency |
| Atomic transaction failures | Verify rollback behavior when backfill fails mid-operation |
| Default value edge cases (null vs zero) | Explicit test cases for each value type's default |
