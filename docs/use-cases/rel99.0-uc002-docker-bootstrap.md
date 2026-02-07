# Use Case: Docker Bootstrap (Docs to Working System)

## Summary

An agent builds crumbs from documentation alone inside a fresh Docker container, then uses crumbs to track its own work. This validates that the documentation is sufficient to produce a working system and that crumbs can serve as the issue tracker.

## Actor and Trigger

The actor is a coding agent (Claude) running inside a Docker container with Go toolchain and access to Claude API. The trigger is starting the container with only the `docs/` directory mounted - no source code, just documentation.

## Flow

### Phase 1: Container Setup

1. **Build Docker image**: Create an image with Go toolchain and Claude Code installed. No crumbs source code.

2. **Start container**: Mount only the `docs/` directory from the host:
   ```bash
   docker run -v ./docs:/workspace/docs crumbs-bootstrap
   ```

3. **Verify environment**: Confirm Go and Claude are available. Confirm no existing source code.

### Phase 2: Generate Release 01.0

4. **Read documentation**: Agent reads VISION.md, ARCHITECTURE.md, road-map.yaml, PRDs, and use cases from the mounted docs directory. The agent must understand the Cupboard and Table interfaces (prd-cupboard-core R2, R3), the Crumb entity structure (prd-crumbs-interface R1), and the SQLite backend requirements (prd-sqlite-backend R1-R4).

5. **Plan work items**: Agent creates a work plan for release 01.0 based on road-map.yaml. The plan covers implementing the core interfaces from prd-cupboard-core and the SQLite backend from prd-sqlite-backend.

6. **Implement release 01.0**: Work through implementation in order:
   - pkg/types: Cupboard interface (prd-cupboard-core R2), Table interface (prd-cupboard-core R3), Crumb struct (prd-crumbs-interface R1), Config struct (prd-cupboard-core R1)
   - internal/sqlite: SQLite backend (prd-sqlite-backend R1-R16), JSONL file format (prd-configuration-directories R3), directory layout (prd-configuration-directories R4)
   - cmd/crumbs: CLI tool with config loading (prd-configuration-directories R7)

7. **Validate use cases**: For each use case in release 01.0, verify success criteria:
   - rel01.0-uc001: Cupboard lifecycle works (Attach/Detach per prd-cupboard-core R4, R5)
   - rel01.0-uc002: SQLite CRUD operations work (prd-sqlite-backend R13)
   - rel01.0-uc003: Core CRUD operations work (prd-crumbs-interface R3, R6, R7, R8, R10)

8. **Build and test**: Compile crumbs and run tests:
   ```bash
   go build ./cmd/crumbs
   go test ./...
   ```

### Phase 3: Self-Hosting with Crumbs

9. **Initialize crumbs**: Create a crumbs cupboard for tracking work. The CLI uses platform-appropriate directories per prd-configuration-directories R1-R2:
   ```bash
   ./crumbs init --data-dir .crumbs
   ```

10. **Import remaining work**: Create crumbs for release 02.0+ work items using the CLI (per eng02-beads-migration command parity):
    ```bash
    ./crumbs create --type task --title "Implement Property entity and Table operations"
    ./crumbs create --type task --title "Implement property backfill on creation"
    # ... additional items from road-map.yaml
    ```

11. **Verify crumbs operations**: Test that crumbs CLI commands work correctly:
    ```bash
    ./crumbs list --json
    ./crumbs ready -n 1 --json --type task
    ./crumbs update <id> --status in_progress
    ./crumbs close <id>
    ```

12. **Continue development**: Use crumbs to track work on releases 02.0, 03.0, 04.0.

### Phase 4: Validate Self-Hosting

13. **Verify crumbs tracks crumbs**: Confirm that crumbs work items are being managed by crumbs itself. The cupboard stores data in JSONL files (prd-configuration-directories R3) with SQLite as ephemeral cache (prd-sqlite-backend R4).

14. **Complete remaining releases**: Work through releases 02.0-04.0, tracking all work in crumbs. Each crumb follows the state lifecycle (prd-crumbs-interface R2): draft → pending → ready → taken → pebble.

15. **Export container**: Package the completed crumbs implementation for use outside the container.

## Architecture Touchpoints

| Component | Role | PRD Reference |
|-----------|------|---------------|
| Docker | Isolated build environment | - |
| docs/ directory | Source of truth for implementation | - |
| crumbs CLI | Work tracking and self-hosting | prd-configuration-directories R7 |
| Cupboard interface | Core storage abstraction with Attach/Detach lifecycle | prd-cupboard-core R2, R4, R5 |
| Table interface | Uniform CRUD operations (Get, Set, Delete, Fetch) | prd-cupboard-core R3 |
| Config struct | Backend selection and data directory | prd-cupboard-core R1, prd-configuration-directories R8 |
| SQLite backend | Storage implementation with JSONL source of truth | prd-sqlite-backend R1-R16 |
| Crumb entity | Work item with state lifecycle and properties | prd-crumbs-interface R1-R11 |
| JSONL files | Line-delimited JSON for data persistence | prd-configuration-directories R3-R4 |

We validate:

- Documentation completeness: Can an agent build crumbs from docs alone?
- PRD accuracy: Do the PRDs (prd-cupboard-core, prd-crumbs-interface, prd-sqlite-backend, prd-configuration-directories) specify enough detail for implementation?
- Interface contracts: Do Cupboard.GetTable and Table operations work as specified in prd-cupboard-core R2-R3?
- Use case validity: Do the success criteria correctly define "done"?
- Self-hosting: Can crumbs track its own development using the state lifecycle from prd-crumbs-interface R2?

## Success Criteria

The demo succeeds when:

- [ ] Docker container starts with only docs/ mounted (no source code)
- [ ] Agent reads and understands all documentation (VISION, ARCHITECTURE, PRDs)
- [ ] Agent implements release 01.0 from documentation (prd-cupboard-core, prd-crumbs-interface, prd-sqlite-backend, prd-configuration-directories)
- [ ] Cupboard interface works: Attach initializes storage (prd-cupboard-core R4), Detach cleans up (prd-cupboard-core R5)
- [ ] Table interface works: Get, Set, Delete, Fetch operations (prd-cupboard-core R3)
- [ ] Crumb entity works: state transitions, property methods (prd-crumbs-interface R4, R5)
- [ ] SQLite backend works: JSONL persistence, startup loading (prd-sqlite-backend R4, R5)
- [ ] All release 01.0 tests pass
- [ ] All release 01.0 use case success criteria are met
- [ ] Crumbs CLI initializes and accepts work items (prd-configuration-directories R1, R2)
- [ ] Remaining releases (02.0+) are tracked in crumbs
- [ ] Final crumbs implementation matches documentation

Observable demo:

```bash
# Build bootstrap image
docker build -t crumbs-bootstrap -f Dockerfile.bootstrap .

# Run bootstrap (mounts only docs/)
docker run -it \
  -v $(pwd)/docs:/workspace/docs:ro \
  -v crumbs-output:/workspace/output \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  crumbs-bootstrap

# Inside container, agent runs:
# 1. Read docs (VISION, ARCHITECTURE, PRDs)
# 2. Implement release 01.0 per prd-cupboard-core, prd-crumbs-interface, prd-sqlite-backend
# 3. Initialize crumbs, import remaining work
# 4. Complete remaining releases using crumbs

# Extract output
docker cp crumbs-bootstrap:/workspace/output ./generated-crumbs
```

## Out of Scope

This use case does not cover:

- Multi-agent coordination (single agent bootstrap)
- Continuous integration setup
- Production deployment
- Performance optimization
- Documentation generation (docs are input, not output)

## Dependencies

Container and tooling dependencies:

- Docker installed on host
- Claude API access
- Go toolchain in image

PRD dependencies (must be complete before this use case):

| PRD | Required sections |
|-----|-------------------|
| prd-cupboard-core | R1 (Config), R2 (Cupboard interface), R3 (Table interface), R4 (Attach), R5 (Detach), R7 (error types), R8 (UUID v7) |
| prd-crumbs-interface | R1 (Crumb struct), R2 (states), R3 (creation), R4 (state transitions), R5 (property methods), R6-R10 (operations) |
| prd-sqlite-backend | R1 (directory layout), R2 (JSONL format), R3 (schema), R4 (startup), R5 (write operations), R11-R15 (implementation) |
| prd-configuration-directories | R1 (CLI config directory), R2 (data directory), R3 (JSONL format), R4 (file layout), R7 (CLI loading) |

Documentation dependencies:

- VISION.md (goals and boundaries)
- ARCHITECTURE.md (components and interfaces)
- road-map.yaml (release schedule)

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Documentation gaps block implementation | Agent files issues for missing specs; human reviews and updates PRDs |
| PRD ambiguity in interface contracts | Cross-reference prd-cupboard-core, prd-crumbs-interface, and prd-sqlite-backend |
| API rate limits slow progress | Use appropriate rate limiting; run overnight if needed |
| Build failures | Agent debugs and fixes; documents any doc corrections needed |
| Container state loss | Mount output volume; commit work frequently |
