package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const version = "0.1.0"

var (
	exitFunc  = os.Exit
	inReader  io.Reader = os.Stdin
	outWriter io.Writer = os.Stdout
	errWriter io.Writer = os.Stderr
)

// Usage prints the usage information to stderr.
func Usage() {
	fmt.Fprintf(errWriter, `workshed v%s - Intent-scoped local workspaces

Usage:
  workshed <command> [flags]

Commands:
  create    Create a new workspace
  list      List workspaces
  inspect   Show workspace details
  path      Show workspace path
  remove    Remove a workspace

Flags:
  -h, --help     Show help

Environment:
  WORKSHED_ROOT  Root directory for workspaces (default: ~/.workshed/workspaces)

Examples:
  workshed create --purpose "Debug payment timeout"
  workshed create --purpose "Add login" --repo git@github.com:org/api@main
  workshed list
  workshed list --purpose debug
  workshed inspect aquatic-fish-motion
  cd $(workshed path aquatic-fish-motion)
  workshed remove aquatic-fish-motion
`, version)
}

// Version prints the current version to stdout.
func Version() {
	fmt.Fprintln(outWriter, version)
}

// GetWorkshedRoot returns the root directory for workspaces, from WORKSHED_ROOT env var or default.
func GetWorkshedRoot() string {
	if root := os.Getenv("WORKSHED_ROOT"); root != "" {
		return root
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(errWriter, "Error: could not determine home directory: %v\n", err)
		exitFunc(1)
	}

	return filepath.Join(home, ".workshed", "workspaces")
}
