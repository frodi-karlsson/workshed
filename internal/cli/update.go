package cli

import (
	"context"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Update(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("update", flag.ExitOnError)
	purpose := fs.String("purpose", "", "New purpose for the workspace (required)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed update --purpose <purpose> [<handle>]\n\n")
		logger.SafeFprintf(r.Stderr, "Update the purpose of a workspace.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed update --purpose \"New focus area\"\n")
		logger.SafeFprintf(r.Stderr, "  workshed update --purpose \"Completed\" my-workspace\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	if *purpose == "" {
		l.Error("missing required flag", "flag", "--purpose")
		fs.Usage()
		r.ExitFunc(1)
	}

	ctx := context.Background()

	providedHandle := ""
	if fs.NArg() >= 1 {
		providedHandle = fs.Arg(0)
	}
	handle := r.ResolveHandle(ctx, providedHandle, l)

	s := r.getStore()
	if err := s.UpdatePurpose(ctx, handle, *purpose); err != nil {
		l.Error("failed to update workspace purpose", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	l.Success("workspace purpose updated", "handle", handle, "purpose", *purpose)
}
