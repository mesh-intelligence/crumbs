# Use Case: Regeneration Compatibility

## Summary

The `--regenerate` cycle deletes all Go code, rebuilds cupboard from documentation, and verifies that the new generation reads and operates on JSONL data created by the previous generation. This validates backward compatibility across regeneration cycles.

## Actor and Trigger

The actor is the `do-work.sh --regenerate` script. The trigger is a developer or CI pipeline running a regeneration cycle to validate that cupboard can be rebuilt from documentation alone and remain compatible with existing data.

## Flow

Each regeneration cycle produces a new generation of the cupboard binary. Generation N creates JSONL data; generation N+1 must read and modify that data without errors. The installed binary from generation N serves as the baseline: if generation N+1 diverges from the JSONL format, the data created by the installed binary will not load.

1. Generation N is built, tested, and installed. The `go install` step places the cupboard binary in `$GOBIN`. This binary becomes the reference implementation for the current JSONL format.

```bash
# End of generation N's do-work cycle
go install ./cmd/cupboard
cupboard --help
# Binary is now in $GOBIN
```

2. Create JSONL data using the generation N binary. This data represents the contract that generation N+1 must honor.

```bash
cd /tmp/regen-compat-test
cupboard init
cupboard set crumbs "" '{"Name":"Gen-N task","State":"draft"}'
cupboard set crumbs "" '{"Name":"Gen-N epic","State":"pending"}'
cupboard list crumbs
# Two crumbs in crumbs.jsonl
```

3. Run `--regenerate`. The script tags the current repo state, deletes all Go source files, reinitializes the Go module, and commits.

```bash
./scripts/do-work.sh --regenerate --cycles 1
```

Internally, `regenerate()` runs:
- `git tag generation-YYYY-MM-DD-HH-MM`
- `find . -name '*.go' -delete` (excluding .git)
- `rm -rf bin/ go.sum && rm -f go.mod`
- `go mod init github.com/mesh-intelligence/crumbs`
- `git add -A && git commit`

4. The make-work/do-work loop rebuilds cupboard from documentation. The new code (generation N+1) is written by agents reading PRDs, architecture docs, and use cases. The agents produce a fresh implementation that must conform to the documented JSONL format (prd-sqlite-backend R2).

5. Build and verify generation N+1.

```bash
go build -o cupboard-new ./cmd/cupboard
```

6. Point generation N+1 at the JSONL data created by generation N. The backend's startup sequence (prd-sqlite-backend R4) loads JSONL into SQLite. If the format has drifted, this step fails.

```bash
cd /tmp/regen-compat-test
./cupboard-new list crumbs
# Must return the two crumbs created in step 2
```

7. Verify read operations against generation N data.

```bash
./cupboard-new get crumbs $CRUMB_ID_1
# Must return "Gen-N task" with state "draft"
./cupboard-new get crumbs $CRUMB_ID_2
# Must return "Gen-N epic" with state "pending"
```

8. Verify write operations update generation N data correctly. Generation N+1 must write JSONL in a format that both itself and generation N can read.

```bash
./cupboard-new set crumbs $CRUMB_ID_1 '{"Name":"Gen-N task","State":"ready"}'
./cupboard-new list crumbs
# CRUMB_ID_1 now shows state "ready"
```

9. Verify the installed generation N binary can still read the data after generation N+1 modifies it. This confirms write compatibility in both directions.

```bash
cupboard list crumbs
# The $GOBIN cupboard (generation N) reads the same JSONL
# CRUMB_ID_1 shows state "ready" (modified by N+1)
# CRUMB_ID_2 shows state "pending" (unmodified)
```

10. Verify property operations across generations (requires rel02.0 property enforcement). If generation N defined properties and set values, generation N+1 must load those properties and values correctly.

```bash
# Using generation N binary to set properties
cupboard set properties "" '{"Name":"estimate","ValueType":"integer"}'
cupboard set crumbs $CRUMB_ID_1 '{"Name":"Gen-N task","State":"ready"}'
# Properties and crumb_properties JSONL files now have data

# Using generation N+1 binary to read
./cupboard-new get crumbs $CRUMB_ID_1
# Must include the "estimate" property with default value
```

## Architecture Touchpoints

Table 1: Components exercised by this use case

| Component | Role | Reference |
|-----------|------|-----------|
| `scripts/do-work.sh --regenerate` | Orchestrates the regeneration cycle: tag, delete, reinit, rebuild | - |
| SQLite backend startup | Loads JSONL files into SQLite; format must match across generations | prd-sqlite-backend R4 |
| JSONL file format | The contract between generations; format stability is the invariant | prd-sqlite-backend R2 |
| JSONL write path | Writes must produce JSONL that older and newer generations can parse | prd-sqlite-backend R5 |
| Unknown field handling | Unknown fields in JSONL are ignored, enabling forward compatibility | prd-sqlite-backend R7.2 |
| Cupboard interface | Attach/Detach, GetTable work identically across generations | prd-cupboard-core R4, R5 |
| Property system | Property definitions and values persist across regeneration | prd-properties-interface R4 |

We validate:

- JSONL format stability across independently-built generations
- Startup sequence loads data from a different generation without errors
- Read operations return correct data from a previous generation's JSONL
- Write operations produce JSONL compatible with both generations
- Unknown fields do not cause failures (forward compatibility)
- Property data survives regeneration

## Success / Demo Criteria

The demo succeeds when all of the following hold.

Table 2: Verification steps

| Step | Action | Expected |
|------|--------|----------|
| 1 | Build and install generation N | `cupboard --help` works from `$GOBIN` |
| 2 | Create crumbs with generation N | `crumbs.jsonl` contains the created entries |
| 3 | Run `--regenerate` | Go files deleted, module reinitialized, clean commit created, tag exists |
| 4 | Build generation N+1 from fresh code | `go build` succeeds, new binary produced |
| 5 | Generation N+1 reads generation N data | `list crumbs` returns all crumbs created in step 2 |
| 6 | Generation N+1 modifies generation N data | `set crumbs` updates state, JSONL reflects change |
| 7 | Generation N reads data modified by N+1 | Installed `cupboard` returns updated values |
| 8 | Property data survives | Properties and values created by N are readable by N+1 |

Observable demo:

```bash
# Full cycle
go install ./cmd/cupboard
cupboard init --datadir /tmp/regen-test
cupboard set crumbs "" '{"Name":"Compat test","State":"draft"}'
CRUMB_ID=$(cupboard list crumbs --json | jq -r '.[0].crumb_id')

./scripts/do-work.sh --regenerate --cycles 1

go build -o /tmp/cupboard-new ./cmd/cupboard
/tmp/cupboard-new list crumbs --datadir /tmp/regen-test
# Expect: "Compat test" crumb present

/tmp/cupboard-new set crumbs "$CRUMB_ID" '{"Name":"Compat test","State":"ready"}' --datadir /tmp/regen-test
cupboard get crumbs "$CRUMB_ID" --datadir /tmp/regen-test
# Expect: state is "ready"
```

## Out of Scope

- Schema migrations or versioned JSONL formats (we rely on format stability and unknown-field tolerance)
- Testing more than two generations in sequence (N and N+1 only)
- Cross-platform compatibility (both generations run on the same OS/architecture)
- Performance comparison between generations
- Validating that agents produce correct code during regeneration (that is the make-work/do-work loop's responsibility)

## Dependencies

### Use Case Dependencies

| Use case | Why |
|----------|-----|
| rel01.1-uc001-go-install | Generation N must be installable via `go install` |
| rel01.1-uc002-jsonl-git-roundtrip | JSONL must be the source of truth, surviving database deletion |
| rel02.0-uc001-property-enforcement | Property data must persist across generations |

### PRD Dependencies

| PRD | Requirements | Purpose |
|-----|-------------|---------|
| prd-sqlite-backend | R2, R4, R5, R7 | JSONL format, startup loading, write persistence, error handling |
| prd-cupboard-core | R3, R4, R5 | Table interface, Attach/Detach lifecycle |
| prd-properties-interface | R4, R9 | Property persistence and seeding |

### Script Dependencies

The `--regenerate` option in `do-work.sh` must be implemented (already done).

## Risks and Mitigations

Table 3: Risks

| Risk | Mitigation |
|------|------------|
| JSONL format drift between generations | PRD specifies the format (prd-sqlite-backend R2); agents read the PRD during regeneration |
| New fields added by N+1 break generation N | Unknown fields are ignored per prd-sqlite-backend R7.2 (forward compatibility) |
| Generation N+1 fails to build | Build verification is step 5 in the success criteria; failure is caught before data tests |
| Property schema changes break cross-generation reads | Property definitions live in JSONL; the format is stable per prd-properties-interface |
| Regeneration deletes test data | Test data lives in a separate directory (`/tmp/regen-test`), not in the repo |
