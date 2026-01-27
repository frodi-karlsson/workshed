package cli

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Capture(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	name := fs.String("name", "", "Name for this capture (required)")
	kind := fs.String("kind", workspace.CaptureKindManual, "Capture kind: manual, execution, checkpoint")
	desc := fs.String("description", "", "Description of this capture")
	tags := fs.StringSlice("tag", nil, "Tags for this capture (can be specified multiple times)")
	format := fs.String("format", "table", "Output format (table|json)")

	fs.Usage = func() {
		logger.UncheckedFprintf(r.Stderr, "Usage: workshed capture [<handle>] --name <name> [flags]\n\n")
		logger.UncheckedFprintf(r.Stderr, "Create a durable capture of git state for all repositories in a workspace.\n\n")
		logger.UncheckedFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.UncheckedFprintf(r.Stderr, "\nExamples:\n")
		logger.UncheckedFprintf(r.Stderr, "  workshed capture --name \"Before refactor\"\n")
		logger.UncheckedFprintf(r.Stderr, "  workshed capture --name \"Checkpoint 1\" --description \"API changes\"\n")
		logger.UncheckedFprintf(r.Stderr, "  workshed capture --name \"Starting point\" --tag test\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
		return
	}

	if err := ValidateFormat(Format(*format), "capture"); err != nil {
		l.Error(err.Error())
		fs.Usage()
		r.ExitFunc(1)
		return
	}

	if *name == "" {
		l.Error("missing required flag --name")
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
	capture, err := s.CaptureState(ctx, handle, workspace.CaptureOptions{
		Name:        *name,
		Kind:        *kind,
		Description: *desc,
		Tags:        *tags,
	})
	if err != nil {
		l.Error("capture failed", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	effectiveFormat := Format(*format)
	if effectiveFormat == FormatJSON {
		data, _ := json.MarshalIndent(capture, "", "  ")
		logger.UncheckedFprintln(r.Stdout, string(data))
		return
	}

	output := Output{
		Columns: []ColumnConfig{
			{Type: Rigid, Name: "KEY", Min: 10, Max: 20},
			{Type: Rigid, Name: "VALUE", Min: 20, Max: 0},
		},
		Rows: [][]string{
			{"id", capture.ID},
			{"name", capture.Name},
			{"kind", capture.Kind},
			{"repos", strconv.Itoa(len(capture.GitState))},
		},
	}

	if err := r.getOutputRenderer().Render(output, effectiveFormat, r.Stdout); err != nil {
		l.Error("failed to render output", "error", err)
	}
}
