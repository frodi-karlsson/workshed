package cli

import (
	"context"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/tui"
	flag "github.com/spf13/pflag"
)

func Inspect(args []string) {
	l := logger.NewLogger(logger.INFO, "inspect")

	fs := flag.NewFlagSet("inspect", flag.ExitOnError)

	fs.Usage = func() {
		logger.SafeFprintf(errWriter, "Usage: workshed inspect [<handle>]\n")
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
		l.Error("failed to get workspace", "handle", handle, "error", err)
		exitFunc(1)
		return
	}

	if tui.IsHumanMode() {
		if err := tui.ShowInspectModal(ctx, store, handle); err != nil {
			l.Error("failed to show inspect modal", "error", err)
			exitFunc(1)
		}
		return
	}

	l.Info("workspace details", "handle", ws.Handle, "purpose", ws.Purpose, "created", ws.CreatedAt.Format("2006-01-02 15:04:05"), "path", ws.Path)

	for _, repo := range ws.Repositories {
		if repo.Ref != "" {
			l.Info("repository", "name", repo.Name, "url", repo.URL, "ref", repo.Ref)
		} else {
			l.Info("repository", "name", repo.Name, "url", repo.URL)
		}
	}
}
