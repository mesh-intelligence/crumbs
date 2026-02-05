# Roadmap

This document defines the release schedule for crumbs. Each release delivers a set of use cases that validate specific functionality. We complete releases in order, prioritizing the earliest incomplete release.

## How to Use This Document

When picking work, find the earliest release with incomplete use cases and work on those first. Release 99.0 contains use cases not yet assigned to a release; they will be scheduled as the roadmap evolves.

Read this document together with VISION.md (goals and boundaries) and ARCHITECTURE.md (components and interfaces).

## Release Schedule

### Release 00.0: Cross-Release Milestones

Milestones that span multiple releases and validate end-to-end system readiness.

| Use Case | Summary | Status |
|----------|---------|--------|
| rel00.0-uc001-self-hosting | Crumbs tracks its own development | Not started |

**Done when**: The crumbs system can track crumbs development work, replacing beads for this repository.

### Release 01.0: Core Storage with SQLite Backend

Implement the Cupboard and Table interfaces with SQLite backend. Validates core concepts and provides a working system for local use.

| Use Case | Summary | Status |
|----------|---------|--------|
| rel01.0-uc001-cupboard-lifecycle | Config, attach, detach, error handling | Not started |
| rel01.0-uc002-sqlite-crud | ORM pattern: get, modify, save entities | Not started |
| rel01.0-uc003-crud-operations | Add, Get, Archive, Purge, Fetch operations | Not started |

**Deliverables**:
- Cupboard interface with Attach, Detach, GetTable
- Table interface with Get, Set, Delete, Fetch
- Crumb entity with state transitions
- SQLite backend with JSON persistence
- Basic CLI commands

**Done when**: All three use cases pass. An agent can create, query, and manage crumbs via the Table interface.

### Release 02.0: Properties with Enforcement

Implement the property system with automatic initialization and backfill. Enables extensibility without schema changes.

| Use Case | Summary | Status |
|----------|---------|--------|
| rel02.0-uc001-property-enforcement | Define properties, auto-init, backfill | Not started |

**Deliverables**:
- Property and Category entities
- PropertyTable with Define, List operations
- Crumb property methods: SetProperty, GetProperty, ClearProperty
- Built-in property seeding (priority, type, description, owner, labels, dependencies)
- Property enforcement: auto-initialization on crumb creation, backfill on property definition

**Done when**: Every crumb always has a value for every defined property. No gaps, no partial state.

### Release 03.0: Trails and Stashes

Implement trail lifecycle for exploratory workflows and stash for shared state coordination.

| Use Case | Summary | Status |
|----------|---------|--------|
| rel03.0-uc001-trail-exploration | Create trails, add crumbs, complete/abandon | Not started |

**Deliverables**:
- Trail entity with lifecycle methods: Complete, Abandon, AddCrumb, RemoveCrumb, GetCrumbs
- Stash entity with methods: SetValue, GetValue, Increment, Acquire, Release, GetHistory
- Trail filtering in Fetch
- Atomic cleanup on trail abandonment

**Done when**: An agent can explore implementation approaches with trails, abandoning failed attempts and completing successful ones.

### Release 04.0: Metadata and Additional Backends

Implement pluggable backend architecture with Dolt for version control.

| Use Case | Summary | Status |
|----------|---------|--------|
| rel04.0-uc001-dolt-backend | Dolt backend with commits, branches, merge | Not started |

**Deliverables**:
- Metadata entity and schema registration
- Dolt backend with version control (commit, branch, checkout, merge)
- Backend abstraction validated with multiple implementations

**Done when**: The same Table interface works with both SQLite and Dolt backends. Dolt provides version-controlled task tracking.

### Release 99.0: Unscheduled

Use cases not yet assigned to a release. These will be scheduled as the roadmap evolves.

| Use Case | Summary | Status |
|----------|---------|--------|
| rel99.0-uc001-blazes-templates | Workflow templates for common patterns | Not started |
| rel99.0-uc002-docker-bootstrap | Build crumbs from docs alone in Docker | Not started |

## Prioritization Rules

1. **Complete releases in order**: Finish all use cases in release N before starting release N+1
2. **Milestones (00.0) validate readiness**: The self-hosting milestone validates that the system is production-ready
3. **Minor releases (X.1, X.2)**: Add functionality to a major release without renumbering; prioritize after the major release completes
4. **Unscheduled (99.0)**: Work on these only when all scheduled releases are complete or when explicitly promoted to a release

## Updating This Document

When a use case is completed, update its status. When adding new use cases:
- Assign to an existing release if it fits the scope
- Create a minor release (e.g., 01.1) for additions to a completed release
- Add to 99.0 if not yet scheduled
