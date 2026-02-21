# Testing Strategy

Tests are organized by architectural layer. Each layer has a defined scope, a fixed set of dependencies it may use, and a location in the repository. Keeping tests in their correct layer prevents integration tests from creeping into unit test files and keeps AI coding agents from needing to read the full stack to fix a single test.

## Layers

| Layer | Location | Scope | Dependencies allowed |
|-------|----------|-------|----------------------|
| Unit | `pkg/schema/` | State-transition methods on entity structs | None (no I/O, no DB) |
| Integration | `internal/persistence/` | JSONL I/O, SQLite schema, mapping, sync strategies | `pkg/api`, `pkg/schema`, `pkg/constants`, SQLite in-memory |
| End-to-end | `cmd/cupboard/` | CLI command output and exit codes | Full stack, temp directory on disk |

## Unit Tests (`pkg/schema`)

Unit tests verify that entity methods produce correct state changes in memory. They do not open files, touch a database, or call any `internal/` package.

```go
// pkg/schema/crumb_test.go
func TestCrumb_Pebble_FromTaken(t *testing.T) {
    c := &Crumb{State: constants.CrumbTaken}
    if err := c.Pebble(); err != nil {
        t.Fatalf("Pebble: %v", err)
    }
    if c.State != constants.CrumbPebble {
        t.Errorf("got %q, want %q", c.State, constants.CrumbPebble)
    }
}
```

Covered behaviors: state transitions (valid and invalid), property get/set/clear, stash acquire/release/increment, trail complete/abandon.

## Integration Tests (`internal/persistence`)

Integration tests verify that the storage layer reads and writes correctly. They use SQLite in-memory databases (`":memory:"`) and `t.TempDir()` for JSONL files—no fixtures on disk, no network.

The three sub-packages each have their own test scope:

| Sub-package | Test scope |
|-------------|-----------|
| `engine/` | Schema creation, JSONL read/write/atomic-rename, sync strategy helpers |
| `mapping/` | Hydration (SQL row → entity struct) and dehydration (entity struct → SQL params) for each entity type |
| `persistence/` (parent) | Full Table.Get/Set/Delete/Fetch round-trips using in-memory SQLite |

```go
// internal/persistence/engine/jsonl_test.go
func TestWriteJSONL_AtomicRename(t *testing.T) {
    dir := t.TempDir()
    // ...write, read back, verify no temp files remain
}
```

Integration tests may not make network calls or write to paths outside `t.TempDir()`.

## End-to-End Tests (`cmd/cupboard`)

End-to-end tests invoke the compiled binary (or call `main` via `exec.Command`) and assert on stdout, stderr, and exit code. They use a temp directory as `DataDir`.

```go
// cmd/cupboard/e2e_test.go
func TestCupboard_CreateCrumb(t *testing.T) {
    dir := t.TempDir()
    out := runCupboard(t, dir, "crumb", "add", "--name", "Fix bug")
    // assert output contains a UUID
}
```

End-to-end tests verify observable behavior only—they do not inspect internal state (JSONL file contents, SQLite rows). If you need to assert on internal state, that belongs in an integration test.

## AI Coding Agent Scope

When an AI agent is given a task, its test scope is determined by the file it is editing:

| File being edited | Tests to run |
|-------------------|-------------|
| `pkg/schema/*.go` | `go test ./pkg/schema/...` |
| `pkg/constants/*.go` | `go test ./...` (constants affect all layers) |
| `internal/persistence/engine/*.go` | `go test ./internal/persistence/engine/...` |
| `internal/persistence/mapping/*.go` | `go test ./internal/persistence/mapping/...` |
| `cmd/cupboard/*.go` | `go test ./cmd/cupboard/...` |

An agent editing `pkg/schema` should not need to read or run `internal/persistence` tests to validate its change. This is the primary value of the layer separation.
