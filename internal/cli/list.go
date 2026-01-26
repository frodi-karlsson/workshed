package cli

import (
	"context"
	"fmt"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

func (r *Runner) List(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("list", flag.ExitOnError)
	purposeFilter := fs.String("purpose", "", "Filter by purpose (case-insensitive substring match)")
	format := fs.String("format", "table", "Output format (table|json)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed list [--purpose <filter>] [--format <table|json>]\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed list\n")
		logger.SafeFprintf(r.Stderr, "  workshed list --purpose payment\n")
		logger.SafeFprintf(r.Stderr, "  workshed list --purpose \"API\" --format json\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	s := r.getStore()

	opts := workspace.ListOptions{
		PurposeFilter: *purposeFilter,
	}

	ctx := context.Background()
	workspaces, err := s.List(ctx, opts)
	if err != nil {
		l.Error("failed to list workspaces", "error", err)
		r.ExitFunc(1)
		return
	}

	if len(workspaces) == 0 {
		if *format == "json" {
			logger.SafeFprintln(r.Stdout, "[]")
		} else {
			l.Info("no workspaces found")
		}
		return
	}

	var rows [][]string
	for _, ws := range workspaces {
		repoCount := len(ws.Repositories)
		var repoInfo string
		if repoCount == 1 {
			repoInfo = ws.Repositories[0].Name
		} else if repoCount > 1 {
			repoInfo = fmt.Sprintf("%d repos", repoCount)
		} else {
			repoInfo = "(empty)"
		}
		created := ws.CreatedAt.Format("2006-01-02 15:04")
		rows = append(rows, []string{ws.Handle, ws.Purpose, repoInfo, created})
	}

	output := Output{
		Columns: []ColumnConfig{
			{Type: Rigid, Name: "HANDLE", Min: 15, Max: 20},
			{Type: Shrinkable, Name: "PURPOSE", Min: 15, Max: 0},
			{Type: Rigid, Name: "REPO", Min: 8, Max: 15},
			{Type: Rigid, Name: "CREATED", Min: 16, Max: 16},
		},
		Rows: rows,
	}

	if err := r.getOutputRenderer().Render(output, Format(*format), r.Stdout); err != nil {
		l.Error("failed to render output", "error", err)
	}
}
