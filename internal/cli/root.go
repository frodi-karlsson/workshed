package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/tui"
	"github.com/frodi/workshed/internal/workspace"
)

const version = "0.2.0"

var (
	exitFunc = os.Exit

	// Settable for testing purposes
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

// GetOrCreateStore creates a workspace store or exits on error.
func GetOrCreateStore(l *logger.Logger) *workspace.FSStore {
	store, err := workspace.NewFSStore(GetWorkshedRoot())
	if err != nil {
		l.Error("failed to create workspace store", "error", err)
		exitFunc(1)
		return nil
	}
	return store
}

// ResolveHandle resolves a workspace handle from args or current directory.
// If no handle is provided and auto-discovery fails, it attempts TUI selection.
// Returns the handle or exits on error.
func ResolveHandle(ctx context.Context, store *workspace.FSStore, providedHandle string, l *logger.Logger) string {
	if providedHandle != "" {
		return providedHandle
	}

	ws, err := store.FindWorkspace(ctx, ".")
	if err != nil {
		l.Error("failed to find workspace", "error", err)
		if h, ok := tui.TrySelectWorkspace(ctx, store, err, l); ok {
			return h
		}
		exitFunc(1)
		return ""
	}
	return ws.Handle
}

func RunMainDashboard() {
	l := logger.NewLogger(logger.ERROR, "workshed")
	store := GetOrCreateStore(l)
	ctx := context.Background()

	if err := tui.RunDashboard(ctx, store); err != nil {
		l.Error("dashboard error", "error", err)
		exitFunc(1)
	}
}
