package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

func (r *Runner) List(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("list", flag.ExitOnError)
	purposeFilter := fs.String("purpose", "", "Filter by purpose (case-insensitive substring match)")
	jsonOutput := fs.Bool("json", false, "Output as JSON")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed list [--purpose <filter>] [--json]\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed list\n")
		logger.SafeFprintf(r.Stderr, "  workshed list --purpose payment\n")
		logger.SafeFprintf(r.Stderr, "  workshed list --purpose \"API\" --json\n")
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
		if *jsonOutput {
			logger.SafeFprintln(r.Stdout, "[]")
		} else {
			l.Info("no workspaces found")
		}
		return
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(workspaces, "", "  ")
		logger.SafeFprintln(r.Stdout, string(data))
		return
	}

	w := tabwriter.NewWriter(r.Stdout, 0, 0, 2, ' ', 0)
	logger.SafeFprintln(w, "HANDLE\tPURPOSE\tREPO\tCREATED")

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
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ws.Handle, ws.Purpose, repoInfo, created); err != nil {
			l.Error("failed to write workspace line", "error", err)
			break
		}
	}

	if err := w.Flush(); err != nil {
		l.Error("failed to flush output", "error", err)
	}
}
