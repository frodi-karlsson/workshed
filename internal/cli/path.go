package cli

import (
	"context"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Path(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("path", flag.ExitOnError)
	format := fs.String("format", "raw", "Output format (raw|table|json)")

	fs.Usage = func() {
		logger.UncheckedFprintf(r.Stderr, "Usage: workshed path [<handle>] [flags]\n\n")
		logger.UncheckedFprintf(r.Stderr, "Print the workspace directory path.\n\n")
		logger.UncheckedFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.UncheckedFprintf(r.Stderr, "\nExamples:\n")
		logger.UncheckedFprintf(r.Stderr, "  workshed path\n")
		logger.UncheckedFprintf(r.Stderr, "  workshed path my-workspace\n")
		logger.UncheckedFprintf(r.Stderr, "  cd $(workshed path)\n")
		logger.UncheckedFprintf(r.Stderr, "  ls $(workshed path)\n")
		logger.UncheckedFprintf(r.Stderr, "  workshed path --format table\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	if err := ValidateFormat(Format(*format), "path"); err != nil {
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
	path, err := s.Path(ctx, handle)
	if err != nil {
		l.Error("failed to get workspace path", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	if *format == "raw" {
		logger.UncheckedFprintln(r.Stdout, path)
		return
	}

	output := Output{
		Columns: []ColumnConfig{
			{Type: Rigid, Name: "KEY", Min: 10, Max: 20},
			{Type: Rigid, Name: "VALUE", Min: 20, Max: 0},
		},
		Rows: [][]string{
			{"path", path},
		},
	}

	if err := r.getOutputRenderer().Render(output, Format(*format), r.Stdout); err != nil {
		l.Error("failed to render output", "error", err)
	}
}
