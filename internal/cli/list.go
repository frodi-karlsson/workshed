package cli

import (
	"context"
	"flag"
	"fmt"
	"text/tabwriter"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
)

// List displays all workspaces with optional filtering by purpose.
func List(args []string) {
	l := logger.NewLogger(logger.INFO, "list")

	fs := flag.NewFlagSet("list", flag.ExitOnError)
	purposeFilter := fs.String("purpose", "", "Filter by purpose (case-insensitive substring match)")

	fs.Usage = func() {
		logger.SafeFprintf(errWriter, "Usage: workshed list [--purpose <filter>]\n\n")
		logger.SafeFprintf(errWriter, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		exitFunc(1)
	}

	store, err := workspace.NewFSStore(GetWorkshedRoot())
	if err != nil {
		l.Error("failed to create workspace store", "error", err)
		exitFunc(1)
	}

	opts := workspace.ListOptions{
		PurposeFilter: *purposeFilter,
	}

	ctx := context.Background()
	workspaces, err := store.List(ctx, opts)
	if err != nil {
		l.Error("failed to list workspaces", "error", err)
		exitFunc(1)
		return
	}

	if len(workspaces) == 0 {
		l.Info("no workspaces found")
		return
	}

	w := tabwriter.NewWriter(outWriter, 0, 0, 2, ' ', 0)
	logger.SafeFprintln(w, "HANDLE\tPURPOSE\tREPO\tCREATED")

	for _, ws := range workspaces {
		repo := truncate(ws.RepoURL, 40)
		created := ws.CreatedAt.Format("2006-01-02 15:04")
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ws.Handle, ws.Purpose, repo, created); err != nil {
			l.Error("failed to write workspace line", "error", err)
			break
		}
	}

	if err := w.Flush(); err != nil {
		l.Error("failed to flush output", "error", err)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}
