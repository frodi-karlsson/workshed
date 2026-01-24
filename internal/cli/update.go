package cli

import (
	"context"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

func Update(args []string) {
	l := logger.NewLogger(logger.INFO, "update")

	fs := flag.NewFlagSet("update", flag.ExitOnError)
	purpose := fs.String("purpose", "", "New purpose for the workspace (required)")

	fs.Usage = func() {
		logger.SafeFprintf(errWriter, "Usage: workshed update --purpose <purpose> [<handle>]\n\n")
		logger.SafeFprintf(errWriter, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		exitFunc(1)
	}

	if *purpose == "" {
		l.Error("missing required flag", "flag", "--purpose")
		fs.Usage()
		exitFunc(1)
		return
	}

	store, err := workspace.NewFSStore(GetWorkshedRoot())
	if err != nil {
		l.Error("failed to create workspace store", "error", err)
		exitFunc(1)
	}

	ctx := context.Background()

	var handle string
	if fs.NArg() >= 1 {
		handle = fs.Arg(0)
	} else {
		ws, err := store.FindWorkspace(ctx, ".")
		if err != nil {
			l.Error("failed to find workspace", "error", err)
			exitFunc(1)
			return
		}
		handle = ws.Handle
	}

	if err := store.UpdatePurpose(ctx, handle, *purpose); err != nil {
		l.Error("failed to update workspace purpose", "handle", handle, "error", err)
		exitFunc(1)
		return
	}

	l.Success("workspace purpose updated", "handle", handle, "purpose", *purpose)
}
