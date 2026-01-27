package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/tui"
	"github.com/frodi/workshed/internal/workspace"
)

type Runner struct {
	Stderr         io.Writer
	Stdout         io.Writer
	Stdin          io.Reader
	ExitFunc       func(int)
	Store          workspace.Store
	Logger         *logger.Logger
	InvocationCWD  string
	TableRenderer  TableRenderer
	OutputRenderer OutputRenderer
}

func (r *Runner) GetInvocationCWD() string {
	return r.InvocationCWD
}

func NewRunner(invocationCWD string) *Runner {
	return &Runner{
		Stderr:         os.Stderr,
		Stdout:         os.Stdout,
		Stdin:          os.Stdin,
		ExitFunc:       os.Exit,
		InvocationCWD:  invocationCWD,
		TableRenderer:  &FlexTableRenderer{},
		OutputRenderer: &OutputRendererImpl{},
	}
}

type OutputRendererImpl struct{}

func (r *OutputRendererImpl) Render(output Output, format Format, out io.Writer) error {
	switch format {
	case FormatTable:
		tableRenderer := &FlexTableRenderer{}
		return tableRenderer.Render(output.Columns, output.Rows, out)
	case FormatJSON:
		jsonRenderer := &JSONRenderer{}
		return jsonRenderer.Render(output, format, out)
	case FormatStream:
		return nil
	default:
		tableRenderer := &FlexTableRenderer{}
		return tableRenderer.Render(output.Columns, output.Rows, out)
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

func (r *Runner) getStore() workspace.Store {
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

func (r *Runner) getOutputRenderer() OutputRenderer {
	if r.OutputRenderer != nil {
		return r.OutputRenderer
	}
	return &OutputRendererImpl{}
}

func (r *Runner) Usage() {
	msg := `workshed v0.3.0 - Intent-scoped local workspaces

Usage:
  workshed <command> [flags]

Commands:
  create     Create a new workspace
  list       List workspaces
  inspect    Show workspace details
  path       Show workspace path
  exec       Run a command in repositories
  repos      Manage repositories in a workspace
  captures   List captures
  capture    Create a capture
  apply      Apply a captured state
  export     Export workspace configuration
  remove     Remove a workspace
  update     Update workspace purpose
  health     Check workspace health
  completion Generate shell completion

Flags:
  -h, --help     Show help
  --format       Output format (table|json|raw) for supported commands

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
cd $(workshed path)
workshed exec -- make test
workshed inspect
workshed captures
workshed capture --name "Before changes"

# Apply a capture by name
workshed apply --name "Before changes"

# List all workspaces
workshed list --purpose "payment"

# Remove a workspace
workshed remove -y

# Generate shell completion
workshed completion --shell bash >> ~/.bash_completion
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
