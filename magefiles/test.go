package main

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Test groups test targets (all, unit, integration).
type Test mg.Namespace

// All runs all tests (unit and integration).
func (Test) All() error {
	return sh.RunV("go", "test", "-v", "./...")
}

// Unit runs only unit tests, excluding the tests/ directory.
func (Test) Unit() error {
	pkgs, err := sh.Output("go", "list", "./...")
	if err != nil {
		return err
	}
	var unitPkgs []string
	for pkg := range strings.SplitSeq(pkgs, "\n") {
		if pkg != "" && !strings.Contains(pkg, "/tests/") && !strings.HasSuffix(pkg, "/tests") {
			unitPkgs = append(unitPkgs, pkg)
		}
	}
	if len(unitPkgs) == 0 {
		fmt.Println("No unit test packages found.")
		return nil
	}
	args := append([]string{"test", "-v"}, unitPkgs...)
	return sh.RunV("go", args...)
}

// Integration builds first, then runs only integration tests.
func (Test) Integration() error {
	mg.Deps(Build)
	return sh.RunV("go", "test", "-v", "./tests/...")
}
