# Use Case: Core CRUD Operations

## Summary

A developer creates a SQLite-backed cupboard, adds crumbs with various states, queries and filters crumbs, and cleans up the database. This tracer bullet validates the core CRUD operations across the Cupboard and Table interfaces without property enforcement.

## Actor and Trigger

The actor is a developer or automated test harness. The trigger is the need to validate that the crumbs system correctly handles the full lifecycle of crumbs: creation, retrieval, state transitions, filtering, and deletion.

## Flow

1. **Create the database**: Construct a Cupboard and call `Attach(config)` with a SQLite backend configuration specifying a DataDir. The backend creates the directory, initializes empty JSONL files, and creates the SQLite schema.

2. **Get crumbs table**: Call `cupboard.GetTable("crumbs")` to obtain a Table for crumbs.

3. **Add first crumb**: Construct a Crumb and call `table.Set("", crumb)`. The operation generates a UUID v7, sets state to "draft", and initializes CreatedAt and UpdatedAt timestamps.

```go
crumb := &Crumb{Name: "Implement login feature"}
id, _ := table.Set("", crumb)
```

4. **Retrieve the crumb**: Call `table.Get(crumbID)` to retrieve the crumb. Verify all fields are populated correctly.

```go
entity, _ := table.Get(id)
crumb := entity.(*Crumb)
```

5. **Change crumb state**: Use entity method SetState to transition from "draft" to "ready", then persist with Table.Set. Verify UpdatedAt changes.

```go
crumb.SetState("ready")
_, _ = table.Set(crumb.CrumbID, crumb)
```

6. **Add second crumb**: Construct another Crumb and call `table.Set("", crumb)`. Verify it is created with state "draft".

```go
crumb2 := &Crumb{Name: "Fix authentication bug"}
id2, _ := table.Set("", crumb2)
```

7. **Fetch all crumbs**: Call `table.Fetch(nil)` or `table.Fetch(map[string]any{})`. Verify both crumbs are returned.

```go
entities, _ := table.Fetch(map[string]any{})
```

8. **Fetch with filter**: Call `table.Fetch(map[string]any{"states": []string{"ready"}})`. Verify only the first crumb (state "ready") is returned.

```go
entities, _ := table.Fetch(map[string]any{"states": []string{"ready"}})
```

9. **Dust a crumb**: Call `crumb.Dust()` then `table.Set` to mark as failed/abandoned. Verify the crumb's state becomes "dust" and UpdatedAt changes.

```go
entity, _ := table.Get(id2)
crumb2 := entity.(*Crumb)
crumb2.Dust()
_, _ = table.Set(crumb2.CrumbID, crumb2)
```

10. **Fetch excludes dust**: Call `table.Fetch(map[string]any{"states": []string{"draft", "ready"}})`. Verify the dust crumb is not returned.

11. **Delete dust crumb**: Call `table.Delete(crumbID)`. Verify the crumb is permanently removed.

```go
_ = table.Delete(id2)
```

12. **Detach the cupboard**: Call `cupboard.Detach()`. Verify all resources are released and subsequent operations return ErrCupboardDetached.

13. **Delete the database**: Remove the DataDir to clean up. Verify the directory and all JSONL files are gone.

## Architecture Touchpoints

This use case exercises the following interfaces and components:

| Interface | Operations Used |
|-----------|-----------------|
| Cupboard | Attach, Detach, GetTable |
| Table | Get, Set, Delete, Fetch |
| Crumb entity | SetState, Dust |

We validate:

- SQLite backend initialization and JSONL file creation (prd-sqlite-backend R1, R4)
- Crumb creation with UUID v7 and timestamp initialization (prd-crumbs-interface R3)
- State transitions and dust behavior (prd-crumbs-interface R4, R5)
- Filter-based queries (prd-crumbs-interface R9, R10)
- Delete cascade behavior (prd-crumbs-interface R8)
- Cupboard lifecycle and ErrCupboardDetached (prd-cupboard-core R4, R5, R6)

## Success Criteria

The demo succeeds when:

- [ ] Attach creates DataDir with all JSONL files and cupboard.db
- [ ] GetTable("crumbs") returns a Table without error
- [ ] Set("", crumb) returns generated UUID v7, crumb has state "draft" and timestamps set
- [ ] Get returns the crumb with all fields populated
- [ ] State transitions via SetState update the state field and UpdatedAt
- [ ] Fetch returns all crumbs when no filter is applied
- [ ] Fetch correctly filters by state
- [ ] Dust changes state to "dust" without deleting data
- [ ] Delete removes crumb permanently
- [ ] Detach prevents further operations with ErrCupboardDetached
- [ ] DataDir removal cleans up all files

Observable demo script:

```bash
# Run the demo binary or test
go test -v ./internal/sqlite -run TestCoreCRUDOperations

# Or run a CLI demo
crumbs demo crud --datadir /tmp/crumbs-demo
```

## Out of Scope

This use case does not cover:

- Property operations (Define, SetProperty, GetProperties, ClearProperty) - see rel02.0-uc001
- Trail operations (Complete, Abandon, belongs_to links) - see rel03.0-uc001
- Metadata operations (schema registration, structured data)
- Link operations beyond what Delete removes
- Concurrent access patterns
- Error recovery scenarios (corrupt JSONL, I/O failures)

## Dependencies

- prd-cupboard-core must be implemented (Cupboard interface: Attach, Detach, GetTable)
- prd-crumbs-interface must be implemented (Crumb entity and methods)
- prd-sqlite-backend must be implemented (JSONL persistence, schema)

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| JSONL file corruption on crash | Atomic write (temp file + rename) per prd-sqlite-backend R5.2 |
| State transition validation | Test invalid transitions return appropriate errors |
| Filter edge cases (empty, invalid) | Explicit test cases for edge cases |
