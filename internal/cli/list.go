package cli

import (
	"context"
	"flag"
	"fmt"
	"text/tabwriter"

	"github.com/frodi/workshed/internal/workspace"
)

// List displays all workspaces with optional filtering by purpose.
func List(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	purposeFilter := fs.String("purpose", "", "Filter by purpose (case-insensitive substring match)")

	fs.Usage = func() {
		fmt.Fprintf(errWriter, "Usage: workshed list [--purpose <filter>]\n\n")
		fmt.Fprintf(errWriter, "Flags:\n")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	store, err := workspace.NewFSStore(GetWorkshedRoot())
	if err != nil {
		fmt.Fprintf(errWriter, "Error: %v\n", err)
		exitFunc(1)
	}

	opts := workspace.ListOptions{
		PurposeFilter: *purposeFilter,
	}

	ctx := context.Background()
	workspaces, err := store.List(ctx, opts)
	if err != nil {
		fmt.Fprintf(errWriter, "Error listing workspaces: %v\n", err)
		exitFunc(1)
	}

	if len(workspaces) == 0 {
		fmt.Fprintln(outWriter, "No workspaces found")
		return
	}

	w := tabwriter.NewWriter(outWriter, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "HANDLE\tPURPOSE\tREPO\tCREATED")

	for _, ws := range workspaces {
		repo := truncate(ws.RepoURL, 40)
		created := ws.CreatedAt.Format("2006-01-02 15:04")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ws.Handle, ws.Purpose, repo, created)
	}

	w.Flush()
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
