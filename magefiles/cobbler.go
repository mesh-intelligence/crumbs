package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

//go:embed prompts/commits.tmpl
var commitsTmplSrc string

var commitTemplates = template.Must(template.New("commits").Parse(commitsTmplSrc))

// commitMsg renders a named commit message template with optional data.
func commitMsg(name string, data any) string {
	var buf bytes.Buffer
	if err := commitTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		panic(fmt.Sprintf("commit template %s: %v", name, err))
	}
	return buf.String()
}

// cobblerConfig holds options shared by measure and stitch targets.
type cobblerConfig struct {
	silenceAgent bool
	maxIssues    int
	promptArg    string
	branch       string
}

// registerCobblerFlags adds the shared flags to fs.
func registerCobblerFlags(fs *flag.FlagSet, cfg *cobblerConfig) {
	fs.BoolVar(&cfg.silenceAgent, "silence-agent", true, "suppress Claude output")
	fs.IntVar(&cfg.maxIssues, "max-issues", 10, "max issues to process")
	fs.StringVar(&cfg.promptArg, "prompt", "", "user prompt text")
	fs.StringVar(&cfg.branch, "branch", "", "generation branch to work on")
}

// resolveCobblerBranch sets cfg.branch from the first positional arg if unset.
func resolveCobblerBranch(cfg *cobblerConfig, fs *flag.FlagSet) {
	if cfg.branch == "" && fs.NArg() > 0 {
		cfg.branch = fs.Arg(0)
	}
}

// runClaude executes Claude with the given prompt.
// If dir is non-empty, the command runs in that directory.
func runClaude(prompt, dir string, silence bool) error {
	fmt.Println("Running Claude...")

	cmd := exec.Command(binClaude, claudeArgs...)
	cmd.Stdin = strings.NewReader(prompt)
	if dir != "" {
		cmd.Dir = dir
	}

	if silence {
		return cmd.Run()
	}

	jq := exec.Command(binJq)
	jq.Stdout = os.Stdout
	jq.Stderr = os.Stderr

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	jq.Stdin = pipe

	if err := jq.Start(); err != nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	defer func() { _ = jq.Wait() }()

	return cmd.Run()
}

// beadsCommit syncs beads state and commits the .beads/ directory.
func beadsCommit(templateName string, data any) {
	_ = exec.Command(binBd, "sync").Run()
	_ = exec.Command(binGit, "add", beadsDir).Run()
	_ = exec.Command(binGit, "commit", "-m", commitMsg(templateName, data), "--allow-empty").Run()
}
