package cli

import (
	"context"
	"encoding/json"
	"text/tabwriter"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Captures(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("captures", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	reverse := fs.Bool("reverse", false, "Reverse sort order (oldest first)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed captures [<handle>]\n\n")
		logger.SafeFprintf(r.Stderr, "List captures for a workspace.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed captures\n")
		logger.SafeFprintf(r.Stderr, "  workshed captures my-workspace\n")
		logger.SafeFprintf(r.Stderr, "  workshed captures --json\n")
		logger.SafeFprintf(r.Stderr, "  workshed captures --reverse\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
		return
	}

	ctx := context.Background()

	providedHandle := ""
	if fs.NArg() >= 1 {
		providedHandle = fs.Arg(0)
	}
	handle := r.ResolveHandle(ctx, providedHandle, l)

	s := r.getStore()
	captures, err := s.ListCaptures(ctx, handle)
	if err != nil {
		l.Error("failed to list captures", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(captures, "", "  ")
		logger.SafeFprintln(r.Stdout, string(data))
		return
	}

	if len(captures) == 0 {
		l.Info("no captures found")
		return
	}

	w := tabwriter.NewWriter(r.Stdout, 0, 0, 2, ' ', 0)
	logger.SafeFprintln(w, "ID\tNAME\tKIND\tREPOS\tCREATED")

	displayCaptures := captures
	if *reverse {
		for i, j := 0, len(captures)-1; i < j; i, j = i+1, j-1 {
			displayCaptures[i], displayCaptures[j] = captures[j], captures[i]
		}
	}

	for _, cap := range displayCaptures {
		created := cap.Timestamp.Format("2006-01-02 15:04")
		logger.SafeFprintf(w, "%s\t%s\t%s\t%d\t%s\n", cap.ID, cap.Name, cap.Kind, len(cap.GitState), created)
	}

	if err := w.Flush(); err != nil {
		l.Error("failed to flush output", "error", err)
	}
}
