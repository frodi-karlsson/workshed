package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/frodi/workshed/internal/workspace"
)

// Remove deletes a workspace after optional confirmation.
func Remove(args []string) {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	force := fs.Bool("force", false, "Skip confirmation prompt")

	fs.Usage = func() {
		fmt.Fprintf(errWriter, "Usage: workshed remove <handle> [--force]\n\n")
		fmt.Fprintf(errWriter, "Flags:\n")
		fs.PrintDefaults()
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

	if !*force {
		fmt.Fprintf(outWriter, "Remove workspace %q (%s)? [y/N]: ", ws.Handle, ws.Purpose)
		reader := bufio.NewReader(inReader)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(errWriter, "Error reading input: %v\n", err)
			exitFunc(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(outWriter, "Cancelled")
			return
		}
	}

	if err := store.Remove(ctx, handle); err != nil {
		fmt.Fprintf(errWriter, "Error removing workspace: %v\n", err)
		exitFunc(1)
	}

	if !*force {
		fmt.Fprintf(outWriter, "Removed workspace: %s\n", handle)
	}
}
