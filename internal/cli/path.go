package cli

import (
	"context"
	"flag"
	"fmt"

	"github.com/frodi/workshed/internal/workspace"
)

// Path prints the filesystem path for a workspace.
func Path(args []string) {
	fs := flag.NewFlagSet("path", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Fprintf(errWriter, "Usage: workshed path <handle>\n")
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
	path, err := store.Path(ctx, handle)
	if err != nil {
		fmt.Fprintf(errWriter, "Error: %v\n", err)
		exitFunc(1)
	}

	fmt.Fprintln(outWriter, path)
}
