package cli

import (
	"context"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Inspect(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	format := fs.String("format", "table", "Output format (table|json)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed inspect [<handle>] [flags]\n\n")
		logger.SafeFprintf(r.Stderr, "Show workspace details including repositories and creation time.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed inspect\n")
		logger.SafeFprintf(r.Stderr, "  workshed inspect aquatic-fish-motion\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	if err := ValidateFormat(Format(*format), "inspect"); err != nil {
		l.Error(err.Error())
		fs.Usage()
		r.ExitFunc(1)
		return
	}

	ctx := context.Background()

	providedHandle := ""
	if fs.NArg() >= 1 {
		providedHandle = fs.Arg(0)
	}
	handle := r.ResolveHandle(ctx, providedHandle, true, l)

	s := r.getStore()
	ws, err := s.Get(ctx, handle)
	if err != nil {
		l.Error("failed to get workspace", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	var rows [][]string
	rows = append(rows, []string{"handle", ws.Handle})
	rows = append(rows, []string{"purpose", ws.Purpose})
	rows = append(rows, []string{"path", ws.Path})
	rows = append(rows, []string{"created", ws.CreatedAt.Format("2006-01-02 15:04:05")})
	for _, repo := range ws.Repositories {
		var repoRow string
		if repo.Ref != "" {
			repoRow = repo.Name + " @ " + repo.Ref
		} else {
			repoRow = repo.Name
		}
		rows = append(rows, []string{"repo", repoRow})
	}

	output := Output{
		Columns: []ColumnConfig{
			{Type: Rigid, Name: "KEY", Min: 10, Max: 20},
			{Type: Rigid, Name: "VALUE", Min: 20, Max: 0},
		},
		Rows: rows,
	}

	if err := r.getOutputRenderer().Render(output, Format(*format), r.Stdout); err != nil {
		l.Error("failed to render output", "error", err)
	}
}
