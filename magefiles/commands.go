package main

// Binary names.
const (
	binGit    = "git"
	binBd     = "bd"
	binClaude = "claude"
	binJq     = "jq"
	binGo     = "go"
	binLint   = "golangci-lint"
)

// Paths and prefixes.
const (
	beadsDir   = ".beads/"
	modulePath = "github.com/mesh-intelligence/crumbs"
	genPrefix  = "generation-"
)

// claudeArgs are the CLI arguments for automated Claude execution.
var claudeArgs = []string{
	"--dangerously-skip-permissions",
	"-p",
	"--verbose",
	"--output-format", "stream-json",
}
