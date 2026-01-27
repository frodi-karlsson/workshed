package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Remove(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	yes := fs.BoolP("yes", "y", false, "Skip confirmation prompt")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed remove [<handle>] [--yes]\n\n")
		logger.SafeFprintf(r.Stderr, "Delete a workspace and all its repositories.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed remove\n")
		logger.SafeFprintf(r.Stderr, "  workshed remove my-workspace\n")
		logger.SafeFprintf(r.Stderr, "  workshed remove -y\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	if !*yes {
		if !term.IsTerminal(os.Stdin.Fd()) {
			l.Error("cannot prompt for confirmation in non-interactive mode", "hint", "use --yes or -y to skip confirmation")
			r.ExitFunc(1)
			return
		}
	}

	ctx := context.Background()

	providedHandle := ""
	if fs.NArg() >= 1 {
		providedHandle = fs.Arg(0)
	}
	handle := r.ResolveHandle(ctx, providedHandle, l)

	s := r.getStore()
	ws, err := s.Get(ctx, handle)
	if err != nil {
		l.Error("workspace not found", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	if !*yes {
		prompt := fmt.Sprintf("Remove workspace %q (%s)? [y/N]: ", ws.Handle, ws.Purpose)
		if _, err := fmt.Fprint(r.Stdout, prompt); err != nil {
			l.Error("failed to write prompt", "error", err)
			r.ExitFunc(1)
			return
		}

		reader := bufio.NewReader(r.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			l.Error("failed to read user input", "error", err)
			r.ExitFunc(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			l.Info("operation cancelled")
			return
		}
	}

	if err := s.Remove(ctx, handle); err != nil {
		l.Error("failed to remove workspace", "handle", handle, "error", err)
		r.ExitFunc(1)
	}

	if !*yes {
		l.Success("workspace removed", "handle", handle)
	}
}
