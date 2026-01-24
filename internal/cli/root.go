package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/frodi/workshed/internal/logger"
)

const version = "0.2.0"

var (
	exitFunc = os.Exit

	// Legacy variables for gradual migration - will be removed after all files are updated
	// These are settable for testing purposes
	outWriter io.Writer = os.Stdout
	errWriter io.Writer = os.Stderr

	// inReader is used for dependency injection in tests
	// Use os.Stdin directly in production, bufio.NewReader(inReader) for testing
	inReader = os.Stdin
)

// Usage prints the usage information to stderr.
func Usage() {
	// Use errWriter for backward compatibility with tests
	// In a production CLI, this would typically go to stderr
	msg := fmt.Sprintf(`workshed v%s - Intent-scoped local workspaces

Usage:
  workshed <command> [flags]

Commands:
  create    Create a new workspace
  list      List workspaces
  inspect   Show workspace details
  path      Show workspace path
  exec      Run a command in repositories
  remove    Remove a workspace
  update    Update workspace purpose

Flags:
  -h, --help     Show help

Environment:
  WORKSHED_ROOT  Root directory for workspaces (default: ~/.workshed/workspaces)
  WORKSHED_LOG_FORMAT  Output format (human|json|raw, default: human)

Examples:
# Create a workspace for a specific task
workshed create --purpose "Debug payment timeout" \
  --repo git@github.com:org/api@main \
  --repo git@github.com:org/worker@develop

# Create a workspace using current directory
workshed create --purpose "Local exploration"

# Commands can use current directory to find workspace
workshed exec -- make test
workshed inspect
workshed path
workshed update --purpose "New purpose"
workshed remove
workshed list
`, version)

	// These output operations should never fail in practice
	logger.SafeFprintf(errWriter, "%s\n", msg)
}

// Version prints the current version to stdout.
func Version() {
	// Use outWriter for backward compatibility with tests
	// These output operations should never fail in practice
	logger.SafeFprintf(outWriter, "%s\n", version)
}

// GetWorkshedRoot returns the root directory for workspaces, from WORKSHED_ROOT env var or default.
func GetWorkshedRoot() string {
	if root := os.Getenv("WORKSHED_ROOT"); root != "" {
		return root
	}

	home, err := os.UserHomeDir()
	if err != nil {
		l := logger.NewLogger(logger.ERROR, "workshed")
		l.Error("failed to determine home directory", "error", err)
		exitFunc(1)
	}

	return filepath.Join(home, ".workshed", "workspaces")
}
