package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
)

type Runner struct {
	Stderr        io.Writer
	Stdout        io.Writer
	Stdin         io.Reader
	ExitFunc      func(int)
	Store         workspace.Store
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

func (r *Runner) GetLogger() *logger.Logger {
	if r.Logger != nil {
		return r.Logger
	}
	return logger.NewLogger(logger.ERROR, "workshed")
}

func (r *Runner) getLogger() *logger.Logger {
	return r.GetLogger()
}

func (r *Runner) GetStore() workspace.Store {
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

func (r *Runner) getStore() workspace.Store {
	return r.GetStore()
}

func (r *Runner) Usage() {
	msg := `workshed - Intent-scoped local development workspaces

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
	logger.UncheckedFprintf(r.Stderr, "%s\n", msg)
}

type WorkspaceNotFoundError struct {
	Handle  string
	Context string
}

func (e *WorkspaceNotFoundError) Error() string {
	if e.Handle != "" {
		return fmt.Sprintf("workspace %q not found", e.Handle)
	}
	return "not in a workspace directory"
}

func (r *Runner) ResolveHandle(ctx context.Context, providedHandle string, validate bool, l *logger.Logger) (string, error) {
	if providedHandle != "" {
		if validate {
			s := r.getStore()
			_, err := s.Get(ctx, providedHandle)
			if err != nil {
				return "", &WorkspaceNotFoundError{Handle: providedHandle}
			}
		}
		return providedHandle, nil
	}

	s := r.getStore()
	ws, err := s.FindWorkspace(ctx, ".")
	if err != nil {
		return "", &WorkspaceNotFoundError{Context: "run from workspace directory or use -- <handle>"}
	}
	return ws.Handle, nil
}

func IsCaptureID(s string) bool {
	return len(s) >= 26 && !strings.Contains(s, " ")
}

func PreflightErrorHint(reason string) string {
	switch reason {
	case "dirty_working_tree":
		return "Commit, stash, or discard your changes before applying"
	case "missing_repository":
		return "The capture references repos not in your workspace. To apply, either add these repos with 'repos add', or choose a capture that matches your workspace"
	case "not_a_git_repository":
		return "This directory is not a git repository"
	case "checkout_failed":
		return "Check that the ref exists in the repository"
	case "head_mismatch":
		return "The branch has diverged; reset or merge first"
	default:
		return ""
	}
}

func MatchesCaptureFilter(cap workspace.Capture, filter string) bool {
	filterLower := strings.ToLower(filter)

	tagFilter := ""
	if strings.HasPrefix(filterLower, "tag:") {
		tagFilter = strings.TrimPrefix(filterLower, "tag:")
	}

	if tagFilter != "" {
		for _, tag := range cap.Metadata.Tags {
			if strings.Contains(strings.ToLower(tag), tagFilter) {
				return true
			}
		}
		return false
	}

	if strings.Contains(strings.ToLower(cap.Name), filterLower) {
		return true
	}

	for _, gitRef := range cap.GitState {
		if strings.Contains(strings.ToLower(gitRef.Repository), filterLower) {
			return true
		}
		if strings.Contains(strings.ToLower(gitRef.Branch), filterLower) {
			return true
		}
	}

	for _, tag := range cap.Metadata.Tags {
		if strings.Contains(strings.ToLower(tag), filterLower) {
			return true
		}
	}

	return false
}
