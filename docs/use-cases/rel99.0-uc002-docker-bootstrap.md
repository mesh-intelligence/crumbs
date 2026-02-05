# Use Case: Docker Bootstrap (Docs to Working System)

## Summary

An agent builds crumbs from documentation alone inside a fresh Docker container, then transitions from beads to crumbs for tracking its own work. This validates that the documentation is sufficient to produce a working system and that crumbs can replace beads as the issue tracker.

## Actor and Trigger

The actor is a coding agent (Claude) running inside a Docker container with Go, beads CLI, and access to Claude API. The trigger is starting the container with only the `docs/` directory mounted - no source code, just documentation.

## Flow

### Phase 1: Container Setup

1. **Build Docker image**: Create an image with Go toolchain, beads CLI, and Claude Code installed. No crumbs source code.

2. **Start container**: Mount only the `docs/` directory from the host:
   ```bash
   docker run -v ./docs:/workspace/docs crumbs-bootstrap
   ```

3. **Verify environment**: Confirm Go, beads, and Claude are available. Confirm no existing source code.

### Phase 2: Generate Release 01.0

4. **Read documentation**: Agent reads VISION.md, ARCHITECTURE.md, ROADMAP.md, PRDs, and use cases from the mounted docs directory.

5. **Create work items**: Using beads, create epics and issues for release 01.0 based on ROADMAP.md:
   ```bash
   bd create --title "Release 01.0: Core Storage" --epic
   bd create --title "Implement Cupboard interface" --parent <epic-id>
   # ... additional issues
   ```

6. **Implement release 01.0**: Work through issues in order, implementing:
   - pkg/types (Cupboard, Table interfaces, entity types)
   - internal/sqlite (SQLite backend implementation)
   - cmd/crumbs (CLI tool)

7. **Validate use cases**: For each use case in release 01.0, verify success criteria:
   - rel01.0-uc001: Cupboard lifecycle works
   - rel01.0-uc002: SQLite CRUD operations work
   - rel01.0-uc003: Core CRUD operations work

8. **Build and test**: Compile crumbs and run tests:
   ```bash
   go build ./cmd/crumbs
   go test ./...
   ```

### Phase 3: Transition to Crumbs

9. **Initialize crumbs**: Create a crumbs cupboard for tracking work:
   ```bash
   ./crumbs init --datadir .crumbs
   ```

10. **Import remaining work**: Create crumbs for release 02.0+ work items:
    ```bash
    ./crumbs add "Implement PropertyTable.Define"
    ./crumbs add "Implement property backfill"
    # ... additional items from ROADMAP.md
    ```

11. **Retire beads**: Stop using beads for new work in this container. Beads data remains for history but new tracking uses crumbs.

12. **Continue development**: Use crumbs to track work on releases 02.0, 03.0, 04.0.

### Phase 4: Validate Self-Hosting

13. **Verify crumbs tracks crumbs**: Confirm that crumbs work items are being managed by crumbs itself.

14. **Complete remaining releases**: Work through releases 02.0-04.0, tracking all work in crumbs.

15. **Export container**: Package the completed crumbs implementation for use outside the container.

## Architecture Touchpoints

| Component | Role |
|-----------|------|
| Docker | Isolated build environment |
| docs/ directory | Source of truth for implementation |
| beads CLI | Initial work tracking (before crumbs exists) |
| crumbs CLI | Self-hosting work tracking (after release 01.0) |
| Cupboard, Table | Core interfaces implemented from PRDs |
| SQLite backend | Storage implementation |

We validate:

- Documentation completeness: Can an agent build crumbs from docs alone?
- PRD accuracy: Do the PRDs specify enough detail for implementation?
- Use case validity: Do the success criteria correctly define "done"?
- Self-hosting: Can crumbs track its own development?
- Beads-to-crumbs transition: Is migration straightforward?

## Success Criteria

The demo succeeds when:

- [ ] Docker container starts with only docs/ mounted (no source code)
- [ ] Agent reads and understands all documentation
- [ ] Agent creates appropriate work items in beads
- [ ] Agent implements release 01.0 from documentation
- [ ] All release 01.0 tests pass
- [ ] All release 01.0 use case success criteria are met
- [ ] Crumbs initializes and accepts work items
- [ ] Agent transitions from beads to crumbs for tracking
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
# 1. Read docs, create beads issues
# 2. Implement release 01.0
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

- Docker installed on host
- Claude API access
- beads CLI available in image
- Go toolchain in image
- All documentation complete (VISION, ARCHITECTURE, ROADMAP, PRDs, use cases)

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Documentation gaps block implementation | Agent files issues for missing specs; human reviews and updates docs |
| API rate limits slow progress | Use appropriate rate limiting; run overnight if needed |
| Build failures | Agent debugs and fixes; documents any doc corrections needed |
| Beads-to-crumbs migration issues | Keep beads available as fallback; document migration steps |
| Container state loss | Mount output volume; commit work frequently |
