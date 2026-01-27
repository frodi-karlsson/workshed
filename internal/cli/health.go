package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
		logger.SafeFprintf(r.Stderr, "Usage: workshed health [<handle>] [flags]\n\n")
		logger.SafeFprintf(r.Stderr, "Check workspace health and report issues.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	handle := r.ResolveHandle(ctx, "", true, l)
	if handle == "" {
		return
	}

	s := r.getStore()

	execs, err := s.ListExecutions(ctx, handle, workspace.ListExecutionsOptions{Limit: 100})
	if err != nil {
		l.Error("failed to list executions", "error", err)
		r.ExitFunc(1)
		return
	}

	staleCount := 0
	for _, e := range execs {
		if time.Since(e.Timestamp) > 30*24*time.Hour {
			staleCount++
		}
	}

	var rows [][]string
	rows = append(rows, []string{"handle", handle})
	if staleCount > 0 {
		rows = append(rows, []string{"status", "issues found"})
		rows = append(rows, []string{"stale_executions", strconv.Itoa(staleCount)})
	} else {
		rows = append(rows, []string{"status", "healthy"})
	}

	if *format == "table" && staleCount > 0 {
		fmt.Printf("Issues found:\n\n")
		fmt.Printf("Stale Executions:\n")
		fmt.Printf("  • %d executions older than 30 days\n", staleCount)
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
