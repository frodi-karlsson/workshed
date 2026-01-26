package cli

import (
	"context"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Path(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("path", flag.ExitOnError)

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed path [<handle>]\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	ctx := context.Background()

	providedHandle := ""
	if fs.NArg() >= 1 {
		providedHandle = fs.Arg(0)
	}
	handle := r.ResolveHandle(ctx, providedHandle, l)

	s := r.getStore()
	path, err := s.Path(ctx, handle)
	if err != nil {
		l.Error("failed to get workspace path", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	l.Info("workspace path", "path", path)
}
