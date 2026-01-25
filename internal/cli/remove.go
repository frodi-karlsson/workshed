package cli

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

// Remove deletes a workspace after optional confirmation.
func Remove(args []string) {
	l := logger.NewLogger(logger.INFO, "remove")

	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	force := fs.Bool("force", false, "Skip confirmation prompt")

	fs.Usage = func() {
		logger.SafeFprintf(errWriter, "Usage: workshed remove [<handle>] [--force]\n\n")
		logger.SafeFprintf(errWriter, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		exitFunc(1)
	}

	store := GetOrCreateStore(l)
	ctx := context.Background()

	providedHandle := ""
	if fs.NArg() >= 1 {
		providedHandle = fs.Arg(0)
	}
	handle := ResolveHandle(ctx, store, providedHandle, l)

	ws, err := store.Get(ctx, handle)
	if err != nil {
		l.Error("workspace not found", "handle", handle, "error", err)
		exitFunc(1)
		return
	}

	if !*force {
		prompt := fmt.Sprintf("Remove workspace %q (%s)? [y/N]: ", ws.Handle, ws.Purpose)
		logger.SafeFprintf(outWriter, "%s", prompt)

		reader := bufio.NewReader(inReader)
		response, err := reader.ReadString('\n')
		if err != nil {
			l.Error("failed to read user input", "error", err)
			exitFunc(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			l.Info("operation cancelled")
			return
		}
	}

	if err := store.Remove(ctx, handle); err != nil {
		l.Error("failed to remove workspace", "handle", handle, "error", err)
		exitFunc(1)
	}

	if !*force {
		l.Success("workspace removed", "handle", handle)
	}
}
