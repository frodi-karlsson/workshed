package cli

import (
	"context"
	"flag"
	"fmt"

	"github.com/frodi/workshed/internal/workspace"
)

// Inspect displays detailed information about a workspace.
func Inspect(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Fprintf(errWriter, "Usage: workshed inspect <handle>\n")
	}

	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(errWriter, "Error: handle is required\n\n")
		fs.Usage()
		exitFunc(1)
	}

	handle := fs.Arg(0)

	store, err := workspace.NewFSStore(GetWorkshedRoot())
	if err != nil {
		fmt.Fprintf(errWriter, "Error: %v\n", err)
		exitFunc(1)
	}

	ctx := context.Background()
	ws, err := store.Get(ctx, handle)
	if err != nil {
		fmt.Fprintf(errWriter, "Error: %v\n", err)
		exitFunc(1)
	}

	fmt.Fprintf(outWriter, "Handle:   %s\n", ws.Handle)
	fmt.Fprintf(outWriter, "Purpose:  %s\n", ws.Purpose)
	fmt.Fprintf(outWriter, "Created:  %s\n", ws.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(outWriter, "Path:     %s\n", ws.Path)

	if ws.RepoURL != "" {
		fmt.Fprintf(outWriter, "Repo:     %s\n", ws.RepoURL)
		if ws.RepoRef != "" {
			fmt.Fprintf(outWriter, "Ref:      %s\n", ws.RepoRef)
		}
	}
}
