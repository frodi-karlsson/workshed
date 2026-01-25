package cli

import (
	"context"

	"github.com/frodi/workshed/internal/logger"
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
	}

	store := GetOrCreateStore(l)
	ctx := context.Background()

	providedHandle := ""
	if fs.NArg() >= 1 {
		providedHandle = fs.Arg(0)
	}
	handle := ResolveHandle(ctx, store, providedHandle, l)

	if err := store.UpdatePurpose(ctx, handle, *purpose); err != nil {
		l.Error("failed to update workspace purpose", "handle", handle, "error", err)
		exitFunc(1)
		return
	}

	l.Success("workspace purpose updated", "handle", handle, "purpose", *purpose)
}
