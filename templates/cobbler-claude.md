# Cobbler

Cobbler is a command-line tool that orchestrates AI coding agents to plan and execute software development work. Like the elves in "The Elves and the Shoemaker," cobbler works through your tasks while you focus on higher-level decisions.

## Project Overview

Cobbler connects to a crumbs cupboard (work item storage) and coordinates AI agents to:

1. **Plan work** - Analyze project state and propose new tasks
2. **Execute work** - Pick up tasks and implement them (docs or code)
3. **Track progress** - Log metrics, tokens, and completion status

## Architecture

```text
cobbler/
├── cmd/cobbler/       # CLI entry point
├── pkg/
│   ├── planner/       # Work planning (analyzes docs, proposes issues)
│   ├── executor/      # Work execution (implements tasks via agents)
│   ├── agent/         # Agent abstraction (Claude, other LLMs)
│   └── crumbs/        # Crumbs client (cupboard integration)
├── internal/
│   ├── prompt/        # Prompt templates and builders
│   └── config/        # Configuration loading
└── docs/              # Documentation
```

## Commands

| Command | Description |
|---------|-------------|
| `cobbler plan` | Analyze project state and propose new work |
| `cobbler stitch` | Pick up available work and implement it |
| `cobbler stitch --docs` | Work on documentation tasks only |
| `cobbler stitch --code` | Work on code tasks only |
| `cobbler stitch <id>` | Work on a specific task |
| `cobbler inspect` | Show current project and work state |
| `cobbler mend` | Fix failing tests or linter errors |

## Design Principles

1. **Cupboard integration**: Cobbler reads from and writes to a crumbs cupboard. It does not have its own task storage.

2. **Agent abstraction**: The agent interface allows different LLM backends. Start with Claude via the Anthropic API.

3. **Prompt-driven**: Planning and execution use structured prompts. Store prompt templates in `internal/prompt/`.

4. **Idempotent operations**: Running `cobbler stitch` twice on the same task should not cause issues.

5. **Observable**: Log what the agent is doing so users can follow along. Support `--verbose` for detailed output.

## Configuration

Cobbler reads configuration from (in order):

1. `.cobbler.yaml` in the current directory
2. `~/.cobbler/config.yaml`
3. Environment variables (`COBBLER_*`)

```yaml
# .cobbler.yaml
cupboard:
  backend: sqlite
  datadir: .crumbs

agent:
  provider: anthropic
  model: claude-sonnet-4-20250514

planner:
  docs_glob: "docs/**/*.md"
  code_glob: "**/*.go"
```

## Implementation Notes

### Planner (`pkg/planner/`)

The planner reads project documentation and proposes work:

1. Read VISION.md, ARCHITECTURE.md, ROADMAP.md
2. Read existing PRDs and use cases
3. Query cupboard for open/closed crumbs
4. Identify gaps between roadmap and current state
5. Propose epics and child tasks
6. Create crumbs in cupboard after user approval

### Executor (`pkg/executor/`)

The executor picks up work and implements it:

1. Query cupboard for ready crumbs
2. Claim a crumb (set state to "taken")
3. Read related docs (PRDs, architecture)
4. Build a prompt for the agent
5. Execute agent loop until task complete
6. Run quality gates (tests, linters)
7. Close the crumb with metrics

### Agent (`pkg/agent/`)

Abstract interface for LLM interaction:

```go
type Agent interface {
    // Run executes the agent with the given prompt and tools
    Run(ctx context.Context, prompt string, tools []Tool) (*Result, error)
}

type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, input map[string]any) (any, error)
}
```

Tools to implement:

- `read_file` - Read file contents
- `write_file` - Write file contents
- `edit_file` - Edit specific lines
- `run_command` - Execute shell commands
- `search_code` - Search codebase with patterns

## Development Workflow

1. Use `go build -o bin/cobbler ./cmd/cobbler` to build
2. Run tests with `go test ./...`
3. Lint with `golangci-lint run`

## Dependencies

- `github.com/petar-djukic/crumbs` - Cupboard client for work item storage
- `github.com/anthropics/anthropic-sdk-go` - Claude API client
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration

## What Cobbler Is NOT

- **Not a task tracker**: Use crumbs for that. Cobbler consumes crumbs.
- **Not an IDE plugin**: Cobbler is a standalone CLI tool.
- **Not a CI system**: Cobbler runs locally, triggered by developers.
- **Not autonomous**: Cobbler proposes work; humans approve before execution.
