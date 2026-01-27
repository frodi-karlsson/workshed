package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/git"
	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Health(args []string) {
	l := r.getLogger()
	ctx := context.Background()

	fs := flag.NewFlagSet("health", flag.ExitOnError)
	format := fs.String("format", "table", "Output format (table|json)")

	fs.Usage = func() {
		logger.UncheckedFprintf(r.Stderr, "Usage: workshed health [<handle>] [flags]\n\n")
		logger.UncheckedFprintf(r.Stderr, "Check workspace health and report issues.\n\n")
		logger.UncheckedFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	providedHandle := ""
	if fs.NArg() >= 1 {
		providedHandle = fs.Arg(0)
	}

	s := r.getStore()

	var handle string
	var ws *workspace.Workspace
	var err error

	if providedHandle != "" {
		ws, err = s.Get(ctx, providedHandle)
		if err != nil {
			l.Error("workspace not found", "handle", providedHandle, "error", err)
			logger.UncheckedFprintf(r.Stderr, "\nHint: Workshed uses 'workshed <command> [<handle>]' syntax.\n")
			logger.UncheckedFprintf(r.Stderr, "  Example: workshed health my-workspace\n")
			logger.UncheckedFprintf(r.Stderr, "  Run 'workshed list' to see available workspaces.\n")
			r.ExitFunc(1)
			return
		}
		handle = ws.Handle
	} else {
		ws, err = s.FindWorkspace(ctx, ".")
		if err != nil {
			l.Error("not in a workspace directory", "error", err)
			logger.UncheckedFprintf(r.Stderr, "\nHint: Run from within a workspace directory or specify a handle:\n")
			logger.UncheckedFprintf(r.Stderr, "  workshed health my-workspace\n")
			logger.UncheckedFprintf(r.Stderr, "  workshed list  # to see available workspaces\n")
			r.ExitFunc(1)
			return
		}
		handle = ws.Handle
	}

	execs, err := s.ListExecutions(ctx, handle, workspace.ListExecutionsOptions{Limit: 100})
	if err != nil {
		l.Error("failed to list executions", "error", err)
		r.ExitFunc(1)
		return
	}

	captures, _ := s.ListCaptures(ctx, handle)

	healthIssues := r.runHealthChecks(ctx, l, ws, execs, captures)

	var rows [][]string
	rows = append(rows, []string{"handle", handle})

	if len(healthIssues) > 0 {
		rows = append(rows, []string{"status", "issues found"})
	} else {
		rows = append(rows, []string{"status", "healthy"})
	}

	if *format == "table" && len(healthIssues) > 0 {
		fmt.Printf("Issues found:\n\n")
		for _, issue := range healthIssues {
			fmt.Printf("  %s\n", issue)
		}
		fmt.Println()
	}

	output := Output{
		Columns: []ColumnConfig{
			{Type: Rigid, Name: "KEY", Min: 20, Max: 20},
			{Type: Rigid, Name: "VALUE", Min: 20, Max: 0},
		},
		Rows: rows,
	}

	if err := r.getOutputRenderer().Render(output, Format(*format), r.Stdout); err != nil {
		l.Error("failed to render output", "error", err)
	}
}

func (r *Runner) runHealthChecks(ctx context.Context, l *logger.Logger, ws *workspace.Workspace, execs []workspace.ExecutionRecord, captures []workspace.Capture) []string {
	var issues []string

	staleThreshold := 30 * 24 * time.Hour
	staleCount := 0
	for _, e := range execs {
		if time.Since(e.Timestamp) > staleThreshold {
			staleCount++
		}
	}
	if staleCount > 0 {
		issues = append(issues, fmt.Sprintf("%d stale executions older than 30 days", staleCount))
	}

	gitClient := git.RealGit{}

	for _, repo := range ws.Repositories {
		repoDir := filepath.Join(ws.Path, repo.Name)
		_, err := os.Stat(repoDir)
		if err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("missing repository directory: %s", repo.Name))
			}
			continue
		}

		gitDir := filepath.Join(repoDir, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("%s is not a git repository", repo.Name))
			}
		} else {
			status, _ := gitClient.StatusPorcelain(ctx, repoDir)
			if strings.TrimSpace(status) != "" {
				issues = append(issues, fmt.Sprintf("%s has uncommitted changes", repo.Name))
			}
		}
	}

	for _, cap := range captures {
		for _, ref := range cap.GitState {
			repoDir := filepath.Join(ws.Path, ref.Repository)
			if _, err := os.Stat(repoDir); err != nil {
				if os.IsNotExist(err) {
					issues = append(issues, fmt.Sprintf("capture '%s' references missing repository: %s", cap.Name, ref.Repository))
				}
			}
		}
	}

	return issues
}
