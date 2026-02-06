# Use Case: Trail-Based Exploration

## Summary

A developer creates a trail to explore an implementation approach, adds crumbs to the trail via belongs_to links, and either completes or abandons the trail. This tracer bullet validates the trail lifecycle and the atomic cleanup behavior when abandoning exploratory work.

## Actor and Trigger

The actor is a coding agent or developer exploring alternative implementation approaches. The trigger is uncertainty about the best approach to a task, requiring exploratory work that may be discarded.

## Flow

1. **Create cupboard and get tables**: Construct a Cupboard and call `Attach(config)`. Get the trails, crumbs, and links tables.

```go
cupboard := NewCupboard()
cupboard.Attach(config)
trailsTable, _ := cupboard.GetTable("trails")
crumbsTable, _ := cupboard.GetTable("crumbs")
linksTable, _ := cupboard.GetTable("links")
```

2. **Create a parent crumb (optional)**: Create a task that the trail will explore approaches for. This is optional; trails can exist without a parent.

```go
parentCrumb := &Crumb{Name: "Implement caching layer"}
parentID, _ := crumbsTable.Set("", parentCrumb)
```

3. **Create a trail**: Construct a Trail and call `trailsTable.Set("", trail)`. The trail starts in "active" state.

```go
trail := &Trail{ParentCrumbID: &parentID}
trailID, _ := trailsTable.Set("", trail)
```

4. **Verify trail state**: Retrieve the trail and confirm it exists with state "active" and the correct parent.

```go
entity, _ := trailsTable.Get(trailID)
trail := entity.(*Trail)
// trail.State should be "active"
// trail.ParentCrumbID should point to parentID
```

5. **Add exploration crumbs**: Create crumbs for the exploration and associate them with the trail via belongs_to links.

```go
crumbA := &Crumb{Name: "Try approach A: in-memory cache"}
crumbAID, _ := crumbsTable.Set("", crumbA)

// Associate crumb with trail via belongs_to link
linkA := &Link{LinkType: "belongs_to", FromID: crumbAID, ToID: trailID}
linksTable.Set("", linkA)
```

6. **Add more crumbs to trail**: Repeat step 5 for additional exploration.

```go
crumbB := &Crumb{Name: "Try approach B: Redis cache"}
crumbBID, _ := crumbsTable.Set("", crumbB)

linkB := &Link{LinkType: "belongs_to", FromID: crumbBID, ToID: trailID}
linksTable.Set("", linkB)
```

7. **List crumbs on trail**: Query the crumbs table with a trail_id filter to get crumbs belonging to this trail.

```go
filter := map[string]any{"trail_id": trailID}
entities, _ := crumbsTable.Fetch(filter)
// Should return both exploration crumbs
```

8. **Remove a crumb from trail**: Delete the belongs_to link to disassociate the crumb without deleting it.

```go
// Find the link for crumbB and delete it
linksTable.Delete(linkBID)
// crumbB is now no longer on the trail but still exists
```

9. **Abandon the trail**: Call the Abandon entity method, then persist with Table.Set. When persisted, the backend:
   - Deletes all crumbs that belong to this trail (via belongs_to links)
   - Leaves crumbs that were removed (step 8) intact

```go
entity, _ := trailsTable.Get(trailID)
trail := entity.(*Trail)
trail.Abandon()                         // updates state in memory
trailsTable.Set(trail.TrailID, trail)   // backend cascades: deletes crumbA
```

10. **Verify cleanup**: Query for the deleted crumb. Verify it returns ErrNotFound. The removed crumb (approach B) should still exist.

```go
_, err := crumbsTable.Get(crumbAID)  // should be ErrNotFound
_, err = crumbsTable.Get(crumbBID)   // should succeed (was removed before abandon)
```

11. **Create another trail and complete it**: Create a new trail, add crumbs, then complete it using the Complete entity method and Table.Set. When persisted, the backend:
    - Removes belongs_to links (crumbs become permanent)
    - Crumbs remain in the database

```go
trail2 := &Trail{}
trail2ID, _ := trailsTable.Set("", trail2)

crumbC := &Crumb{Name: "Successful approach"}
crumbCID, _ := crumbsTable.Set("", crumbC)
linkC := &Link{LinkType: "belongs_to", FromID: crumbCID, ToID: trail2ID}
linksTable.Set("", linkC)

entity, _ := trailsTable.Get(trail2ID)
trail2 := entity.(*Trail)
trail2.Complete()                         // updates state in memory
trailsTable.Set(trail2.TrailID, trail2)   // backend cascades: removes links
```

12. **Verify completion**: Crumbs from the completed trail should still exist and be queryable.

```go
entity, _ := crumbsTable.Get(crumbCID)  // should succeed
crumbC := entity.(*Crumb)               // crumb is now permanent
```

13. **Filter crumbs by trail**: After completion, crumbs no longer have belongs_to links, so trail_id filter returns empty.

```go
filter := map[string]any{"trail_id": trail2ID}
entities, _ := crumbsTable.Fetch(filter)
// Should return empty slice (links were removed on complete)
```

14. **Detach the cupboard**: Call `cupboard.Detach()`.

## Architecture Touchpoints

This use case exercises the following interfaces and components:

| Interface | Operations Used |
|-----------|-----------------|
| Cupboard | Attach, Detach, GetTable |
| Table (crumbs) | Get, Set, Delete, Fetch (with trail_id filter) |
| Table (trails) | Get, Set |
| Table (links) | Get, Set, Delete, Fetch |
| Trail entity | Complete, Abandon |

We validate:

- Trail creation via Table.Set (prd-trails-interface R3)
- Crumb-trail membership via belongs_to links (prd-trails-interface R7)
- Trail completion makes crumbs permanent by removing links (prd-trails-interface R5)
- Trail abandonment deletes associated crumbs atomically (prd-trails-interface R6)
- Trail_id filter in crumbs Fetch (prd-crumbs-interface R9)
- belongs_to link management (prd-sqlite-backend)

## Success Criteria

The demo succeeds when:

- [ ] Trail created via Table.Set starts in "active" state
- [ ] belongs_to links associate crumbs with trails
- [ ] Fetch with trail_id filter returns crumbs on that trail
- [ ] Deleting belongs_to link disassociates without deleting the crumb
- [ ] trail.Abandon() then Table.Set deletes trail's crumbs atomically
- [ ] Crumbs removed from trail before abandon survive
- [ ] trail.Complete() then Table.Set removes belongs_to links
- [ ] Completed trail's crumbs remain queryable
- [ ] Trail state is "completed" or "abandoned" after respective operations
- [ ] No orphan crumbs or links after abandonment

Observable demo script:

```bash
# Run the demo binary or test
go test -v ./internal/sqlite -run TestTrailExploration

# Or run a CLI demo
crumbs demo trails --datadir /tmp/crumbs-demo
```

## Out of Scope

This use case does not cover:

- Nested trails (trails containing trails)
- Trail merging or splitting
- Trail history or versioning
- Stash operations - see rel03.0-uc002 (if created)
- Concurrent trail operations
- Trail templates or predefined workflows

## Dependencies

- rel01.0-uc001 (Cupboard lifecycle) must pass
- rel01.0-uc003 (Core CRUD) must pass
- prd-trails-interface must be implemented
- prd-sqlite-backend (links table) must be implemented

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Orphan crumbs after failed abandonment | Use transaction; rollback on failure |
| belongs_to links not cleaned up | Verify link cleanup in tests |
| Performance with large trails | Test with 100+ crumbs on a single trail |
| Race condition on concurrent trail ops | Document single-writer expectation for trails |
