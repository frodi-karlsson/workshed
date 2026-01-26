package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Derive(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("derive", flag.ExitOnError)
	output := fs.String("output", "", "Output file path (default: <workspace>/.workshed/context.json)")
	format := fs.String("format", "", "Output format (table|json), defaults based on --output extension")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed derive [<handle>] [flags]\n\n")
		logger.SafeFprintf(r.Stderr, "Derive workspace context including repository paths, git state, and metadata.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed derive\n")
		logger.SafeFprintf(r.Stderr, "  workshed derive --format json | jq '.captures'\n")
		logger.SafeFprintf(r.Stderr, "  workshed derive --output /tmp/context.json\n")
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
	wsPath, err := s.Path(ctx, handle)
	if err != nil {
		l.Error("failed to get workspace path", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	contextData, err := s.DeriveContext(ctx, handle)
	if err != nil {
		l.Error("derive failed", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	outputPath := *output
	if outputPath == "" {
		outputPath = filepath.Join(wsPath, ".workshed", "context.json")
	}

	data, err := json.MarshalIndent(contextData, "", "  ")
	if err != nil {
		l.Error("marshaling context", "error", err)
		r.ExitFunc(1)
		return
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		l.Error("writing context", "path", outputPath, "error", err)
		r.ExitFunc(1)
		return
	}

	effectiveFormat := Format(*format)
	if effectiveFormat == "" {
		effectiveFormat = DetectFormatFromFilePath(outputPath)
	}

	if effectiveFormat == FormatJSON {
		logger.SafeFprintln(r.Stdout, string(data))
	} else {
		output := Output{
			Columns: []ColumnConfig{
				{Type: Rigid, Name: "KEY", Min: 10, Max: 20},
				{Type: Rigid, Name: "VALUE", Min: 20, Max: 0},
			},
			Rows: [][]string{
				{"path", outputPath},
				{"repos", strconv.Itoa(len(contextData.Repositories))},
			},
		}
		if err := r.getOutputRenderer().Render(output, FormatTable, r.Stdout); err != nil {
			l.Error("failed to render output", "error", err)
		}
	}
}
