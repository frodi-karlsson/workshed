package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/store"
	"github.com/frodi/workshed/internal/tui"
	"github.com/frodi/workshed/internal/workspace"
)

type Runner struct {
	Stderr        io.Writer
	Stdout        io.Writer
	Stdin         io.Reader
	ExitFunc      func(int)
	Store         store.Store
	Logger        *logger.Logger
	InvocationCWD string
}

func (r *Runner) GetInvocationCWD() string {
	return r.InvocationCWD
}

func NewRunner(invocationCWD string) *Runner {
	return &Runner{
		Stderr:        os.Stderr,
		Stdout:        os.Stdout,
		Stdin:         os.Stdin,
		ExitFunc:      os.Exit,
		InvocationCWD: invocationCWD,
	}
}

func (r *Runner) getWorkshedRoot() string {
	if root := os.Getenv("WORKSHED_ROOT"); root != "" {
		return root
	}
	home, err := os.UserHomeDir()
	if err != nil {
		l := r.getLogger()
		l.Error("failed to determine home directory", "error", err)
		r.ExitFunc(1)
		return ""
	}
	return filepath.Join(home, ".workshed", "workspaces")
}

func (r *Runner) getLogger() *logger.Logger {
	if r.Logger != nil {
		return r.Logger
	}
	return logger.NewLogger(logger.ERROR, "workshed")
}

func (r *Runner) getStore() store.Store {
	if r.Store != nil {
		return r.Store
	}
	l := r.getLogger()
	s, err := workspace.NewFSStore(r.getWorkshedRoot())
	if err != nil {
		l.Error("failed to create workspace store", "error", err)
		r.ExitFunc(1)
		return nil
	}
	return s
}

func (r *Runner) Usage() {
	msg := `workshed v0.2.9 - Intent-scoped local workspaces

Usage:
  workshed <command> [flags]

Commands:
  create    Create a new workspace
  list      List workspaces
  inspect   Show workspace details
  path      Show workspace path
  exec      Run a command in repositories
  repo      Manage repositories in a workspace
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
`
	logger.SafeFprintf(r.Stderr, "%s\n", msg)
}

func (r *Runner) Version() {
	logger.SafeFprintf(r.Stdout, "%s\n", version)
}

func (r *Runner) ResolveHandle(ctx context.Context, providedHandle string, l *logger.Logger) string {
	if providedHandle != "" {
		return providedHandle
	}

	s := r.getStore()
	ws, err := s.FindWorkspace(ctx, ".")
	if err != nil {
		l.Error("failed to find workspace", "error", err)
		r.ExitFunc(1)
		return ""
	}
	return ws.Handle
}

func (r *Runner) RunMainDashboard() {
	l := r.getLogger()
	s := r.getStore()
	ctx := context.Background()

	if err := tui.RunDashboard(ctx, s, r); err != nil {
		l.Error("dashboard error", "error", err)
		r.ExitFunc(1)
	}
}

func (r *Runner) Repo(args []string) {
	if len(args) < 1 {
		r.RepoUsage()
		r.ExitFunc(1)
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "add":
		r.RepoAdd(args[1:])
	case "remove":
		r.RepoRemove(args[1:])
	case "help", "-h", "--help":
		r.RepoUsage()
	default:
		logger.SafeFprintf(r.Stderr, "Unknown repo subcommand: %s\n\n", subcommand)
		r.RepoUsage()
		r.ExitFunc(1)
	}
}

func (r *Runner) RepoUsage() {
	msg := `workshed repo - Manage repositories in a workspace

Usage:
  workshed repo add <handle> --repo url[@ref]...
  workshed repo remove <handle> --repo <name>

Subcommands:
  add     Add repositories to a workspace
  remove  Remove a repository from a workspace

Examples:
  workshed repo add my-workspace --repo https://github.com/org/repo@main

  workshed repo remove my-workspace --repo my-repo
`
	logger.SafeFprintf(r.Stderr, "%s\n", msg)
}
