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
	page := fs.Int("page", 1, "Page number (1-indexed)")
	pageSize := fs.Int("page-size", 20, "Number of items per page")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed list [--purpose <filter>] [--page <n>] [--page-size <n>] [--format <table|json>]\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed list\n")
		logger.SafeFprintf(r.Stderr, "  workshed list --purpose payment\n")
		logger.SafeFprintf(r.Stderr, "  workshed list --purpose \"API\" --format json\n")
		logger.SafeFprintf(r.Stderr, "  workshed list --page 2 --page-size 10\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	if err := ValidateFormat(Format(*format), "list"); err != nil {
		l.Error(err.Error())
		fs.Usage()
		r.ExitFunc(1)
		return
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

	total := len(workspaces)
	if *page < 1 {
		*page = 1
	}
	if *pageSize < 1 {
		*pageSize = 20
	}

	startIdx := (*page - 1) * *pageSize
	endIdx := startIdx + *pageSize

	if startIdx >= total {
		if *format == "json" {
			logger.SafeFprintln(r.Stdout, "[]")
		} else {
			l.Info(fmt.Sprintf("page %d is empty (total: %d items)", *page, total))
		}
		return
	}

	if endIdx > total {
		endIdx = total
	}

	pagedWorkspaces := workspaces[startIdx:endIdx]

	if *format == "raw" {
		for _, ws := range pagedWorkspaces {
			logger.SafeFprintln(r.Stdout, ws.Handle)
		}
		return
	}

	var rows [][]string
	for _, ws := range pagedWorkspaces {
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

	if *format != "json" && total > *pageSize {
		l.Info(fmt.Sprintf("showing %d-%d of %d workspaces (page %d of %d)", startIdx+1, endIdx, total, *page, (total+*pageSize-1)/(*pageSize)))
	}
}
