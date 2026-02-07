# Use Case: Issue-Tracking CLI

## Summary

A user or agent runs the seven issue-tracking commands (`create`, `ready`, `update`, `close`, `show`, `list`, `comments`) against a fresh cupboard and verifies correct output in both human-readable and JSON formats.

## Actor and Trigger

The actor is a developer or coding agent using the cupboard CLI for task tracking. The trigger is invoking any of the issue-tracking commands from the terminal.

## Flow

This use case exercises the cupboard CLI commands from Phase 1 of rel02.1-uc003-self-hosting. Each command uses the Cupboard library via `GetTable("crumbs")` (prd-cupboard-core R2) and operates through the Table interface (prd-cupboard-core R3).

1. Initialize a fresh cupboard with SQLite backend.

```bash
mkdir -p /tmp/cupboard-demo
cat > /tmp/cupboard-demo/config.yaml <<EOF
backend: sqlite
data_dir: /tmp/cupboard-demo/data
EOF
export CUPBOARD_CONFIG=/tmp/cupboard-demo/config.yaml
```

2. Create a task with required fields using `cupboard create`.

```bash
cupboard create --type task --title "Implement feature" --description "Add the widget"
# Expected output (human-readable):
# Created crumb <uuid>
```

The command calls `Table.Set("", crumb)` with an empty ID to generate a UUID v7 (prd-cupboard-core R8, prd-crumbs-interface R3). The crumb is created with state `draft` and properties initialized to type-based defaults (prd-crumbs-interface R3.2, prd-properties-interface R3.5).

3. Create a task with JSON output to capture the ID.

```bash
cupboard create --type task --title "Write tests" --description "Unit and integration" --json
# Expected output (JSON):
# {"crumb_id":"<uuid>","name":"Write tests","state":"open","type":"task",...}
```

4. Create an epic with labels.

```bash
cupboard create --type epic --title "Storage layer" --description "Core storage" --labels "code,infra"
# Expected: Created crumb <uuid>
```

The `--type` and `--labels` flags set categorical and list properties (prd-properties-interface R3, R9).

5. List all crumbs using `cupboard list`.

```bash
cupboard list --json
# Expected output (JSON array):
# [{"crumb_id":"...","name":"Implement feature",...},{"crumb_id":"...","name":"Write tests",...},...]
```

The command calls `Table.Fetch(nil)` to retrieve all crumbs (prd-crumbs-interface R10).

6. Show a specific crumb using `cupboard show`.

```bash
cupboard show <crumb_id>
# Expected output (human-readable):
# ID: <crumb_id>
# Title: Write tests
# State: open
# Type: task
# Description: Unit and integration
# Created: 2024-...
```

The command calls `Table.Get(id)` and type-asserts to `*Crumb` (prd-cupboard-core R3.2, prd-crumbs-interface R6).

7. Find available work using `cupboard ready`.

```bash
cupboard ready --json --type task
# Expected output (JSON array of open tasks):
# [{"crumb_id":"...","name":"Implement feature","state":"open","type":"task"},...]
```

The command calls `Table.Fetch` with a filter for state and type (prd-crumbs-interface R9, R10).

8. Claim a task using `cupboard update`.

```bash
cupboard update <crumb_id> --status in_progress
# Expected: Updated crumb <crumb_id>
```

The command retrieves the crumb via `Table.Get`, calls `crumb.SetState("taken")` (prd-crumbs-interface R4), then persists with `Table.Set`.

9. Verify the status change.

```bash
cupboard show <crumb_id>
# Expected: State: in_progress
```

10. Add a comment to track work using `cupboard comments add`.

```bash
cupboard comments add <crumb_id> "tokens: 34256"
# Expected: Added comment to <crumb_id>
```

Comments are stored as metadata entries associated with the crumb (prd-sqlite-backend R2.8).

11. Verify the comment appears in the crumb details.

```bash
cupboard show <crumb_id>
# Expected output includes:
# Comments:
#   - tokens: 34256
```

12. Close the task using `cupboard close`.

```bash
cupboard close <crumb_id>
# Expected: Closed crumb <crumb_id>
```

The command calls `crumb.Pebble()` (prd-crumbs-interface R4.3), which requires the crumb to be in `taken` state, then persists with `Table.Set`.

13. Verify the final state.

```bash
cupboard show <crumb_id>
# Expected: State: closed
```

14. Confirm ready excludes closed tasks.

```bash
cupboard ready --json --type task
# Expected: JSON array without the closed task
```

## Architecture Touchpoints

Table 1: Components and PRD references

| Component | Interface/Operation | PRD Reference |
|-----------|---------------------|---------------|
| cupboard CLI | `create` command | prd-crumbs-interface R3 (creation via Table.Set) |
| cupboard CLI | `ready` command | prd-crumbs-interface R9, R10 (Fetch with filter) |
| cupboard CLI | `update` command | prd-crumbs-interface R4 (SetState) |
| cupboard CLI | `close` command | prd-crumbs-interface R4.3 (Pebble) |
| cupboard CLI | `show` command | prd-cupboard-core R3.2 (Table.Get) |
| cupboard CLI | `list` command | prd-crumbs-interface R10 (Table.Fetch) |
| cupboard CLI | `comments add` command | prd-sqlite-backend R2.8 (metadata) |
| Cupboard interface | `GetTable("crumbs")` | prd-cupboard-core R2 |
| Table interface | Get, Set, Fetch | prd-cupboard-core R3 |
| Crumb entity | State transitions, Properties | prd-crumbs-interface R2, R4, R5 |
| Properties | type, priority, labels | prd-properties-interface R9 |
| SQLite backend | JSONL persistence | prd-sqlite-backend R5 |

## Success / Demo Criteria

Run the following sequence and verify observable outputs.

Table 2: Demo script

| Step | Command | Verify |
|------|---------|--------|
| 1 | `cupboard create --type task --title "Demo task" --description "For testing" --json` | JSON output includes `crumb_id`, `state: open`, `type: task` |
| 2 | `cupboard list --json` | JSON array contains the created task |
| 3 | `cupboard show <crumb_id>` | Human-readable output shows title, state, type, description |
| 4 | `cupboard ready --json --type task` | JSON array includes the task (state is open) |
| 5 | `cupboard update <crumb_id> --status in_progress` | Exit code 0, confirmation message |
| 6 | `cupboard ready --json --type task` | JSON array excludes the task (no longer open) |
| 7 | `cupboard comments add <crumb_id> "tokens: 12345"` | Exit code 0, confirmation message |
| 8 | `cupboard show <crumb_id>` | Output includes the comment text |
| 9 | `cupboard close <crumb_id>` | Exit code 0, confirmation message |
| 10 | `cupboard show <crumb_id>` | State shows as closed |

All test cases from test001-self-hosting.yaml validate these commands. The test suite provides comprehensive coverage including error cases (missing title, invalid ID, etc.).

## Out of Scope

- Script integration (do-work.sh, make-work.sh) — covered by rel02.1-uc003-self-hosting Phase 2
- Interactive agent workflow rules — covered by rel02.1-uc003-self-hosting Phase 3
- Trail operations (create, complete, abandon)
- Multi-user or concurrent access patterns
- Remote backends

## Dependencies

- Cupboard interface with Attach/Detach/GetTable (prd-cupboard-core R2, R4, R5)
- Table interface with Get/Set/Delete/Fetch (prd-cupboard-core R3)
- Crumb entity with state transitions and property methods (prd-crumbs-interface R1–R5)
- SQLite backend with JSONL persistence (prd-sqlite-backend R1–R5)
- Property definitions for type, priority, labels (prd-properties-interface R1, R9)
- Built-in properties seeded on first startup (prd-sqlite-backend R9)

## Risks and Mitigations

Table 3: Risks

| Risk | Mitigation |
|------|------------|
| CLI output format changes break scripts | Use `--json` flag for machine-readable output; document JSON schema |
| Missing validation returns cryptic errors | CLI validates inputs before calling library; provides user-friendly messages |
| Property initialization fails silently | Backend logs initialization errors; test suite verifies property defaults |
