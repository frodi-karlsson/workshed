package cli

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strconv"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Import(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("import", flag.ExitOnError)
	file := fs.String("file", "-", "Input file path (- for stdin)")
	preserveHandle := fs.Bool("preserve-handle", false, "Use original handle instead of generating new one")
	force := fs.Bool("force", false, "Overwrite workspace if handle already exists")
	format := fs.String("format", "table", "Output format (table|json)")

	fs.Usage = func() {
		logger.UncheckedFprintf(r.Stderr, "Usage: workshed import <file.json> [flags]\n\n")
		logger.UncheckedFprintf(r.Stderr, "Create a workspace from an exported JSON file.\n\n")
		logger.UncheckedFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.UncheckedFprintf(r.Stderr, "\nExamples:\n")
		logger.UncheckedFprintf(r.Stderr, "  workshed import workspace.json\n")
		logger.UncheckedFprintf(r.Stderr, "  workshed import workspace.json --preserve-handle\n")
		logger.UncheckedFprintf(r.Stderr, "  cat workspace.json | workshed import -\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
		return
	}

	var data []byte
	var err error

	if *file == "-" {
		data, err = io.ReadAll(r.Stdin)
		if err != nil {
			l.Error("reading from stdin", "error", err)
			r.ExitFunc(1)
			return
		}
	} else {
		data, err = os.ReadFile(*file)
		if err != nil {
			l.Error("reading file", "path", *file, "error", err)
			r.ExitFunc(1)
			return
		}
	}

	var wsContext workspace.WorkspaceContext
	if err := json.Unmarshal(data, &wsContext); err != nil {
		l.Error("parsing JSON", "error", err)
		r.ExitFunc(1)
		return
	}

	s := r.getStore()
	ctx := context.Background()

	ws, err := s.ImportContext(ctx, workspace.ImportOptions{
		Context:        &wsContext,
		InvocationCWD:  r.InvocationCWD,
		PreserveHandle: *preserveHandle,
		Force:          *force,
	})
	if err != nil {
		l.Error("import failed", "error", err)
		r.ExitFunc(1)
		return
	}

	output := Output{
		Columns: []ColumnConfig{
			{Type: Rigid, Name: "KEY", Min: 10, Max: 20},
			{Type: Rigid, Name: "VALUE", Min: 20, Max: 0},
		},
		Rows: [][]string{
			{"handle", ws.Handle},
			{"purpose", ws.Purpose},
			{"repos", strconv.Itoa(len(ws.Repositories))},
			{"path", ws.Path},
		},
	}

	if err := r.getOutputRenderer().Render(output, Format(*format), r.Stdout); err != nil {
		l.Error("failed to render output", "error", err)
	}
}
