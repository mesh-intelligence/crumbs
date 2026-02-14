# magefiles

Build tooling for the crumbs project, implemented as [Mage](https://magefile.org) targets. Mage requires all target files to live in a single flat directory (no subdirectories for Go source), so we organize by concern: one file per subsystem.

All orchestration logic (cobbler, generator, beads, stats, container execution) is provided by the shared [mage-claude-orchestrator](https://github.com/mesh-intelligence/mage-claude-orchestrator) library. The magefiles in this directory are thin wrappers that load `configuration.yaml` and delegate to the library.

## Files

| File | Purpose |
|------|---------|
| orchestrator.go | Namespace types (`Cobbler`, `Generator`, `Beads`), config loading, one-liner target wrappers |
| build.go | Top-level targets: `Build`, `Clean`, `Init`, `Reset`, `Install` |
| test.go | Test targets: `Cobbler`, `Generator`, `Resume`, `Unit`, `Integration`, `All` |
| stats.go | `Stats` target: delegates to library for Go LOC and documentation word counts |
| lint.go | `Lint` target: runs golangci-lint |

## Directories

| Directory | Contents |
|-----------|----------|
| prompts/ | Go templates (`measure.tmpl`, `stitch.tmpl`) rendered as Claude prompts |
| seeds/ | Go templates for seed files created during generator start/reset |

## Other Files

| File | Purpose |
|------|---------|
| test-plan.yaml | Test plan for generator lifecycle and isolation tests |

## Architecture

Mage targets are grouped into namespaces using Go struct types. The `Generator`, `Cobbler`, `Beads`, and `Test` structs each use `mg.Namespace` to create `generator:*`, `cobbler:*`, `beads:*`, and `test:*` targets. Top-level targets (`Build`, `Reset`, `Init`) live in build.go as package-level functions.

On startup, `orchestrator.go` loads `configuration.yaml` into a `baseCfg` variable. Each target creates a new `orchestrator.Orchestrator` from this config and calls the corresponding library method. There are no CLI flags; all settings come from the configuration file.

Test targets in `test.go` create orchestrators with config overrides (e.g. `SilenceAgent`, `MaxIssues`, `Cycles`) to control test behavior. Verification helpers (git branch/tag checks, beads list queries) are defined inline using `exec.Command`.

For the generation workflow, see [eng02-generation-workflow.md](../docs/engineering/eng02-generation-workflow.md).
