package cli

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Validate(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	path := fs.String("path", "", "Path to AGENTS.md file (default: AGENTS.md in workspace directory)")
	jsonOutput := fs.Bool("json", false, "Output as JSON")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed validate [--path <file>] [<handle>]\n\n")
		logger.SafeFprintf(r.Stderr, "Validate AGENTS.md file structure and required sections.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed validate\n")
		logger.SafeFprintf(r.Stderr, "  workshed validate --path ./AGENTS.md\n")
		logger.SafeFprintf(r.Stderr, "  workshed validate --json\n")
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

	effectivePath := *path
	if effectivePath == "" {
		effectivePath = filepath.Join(wsPath, "AGENTS.md")
	} else if !filepath.IsAbs(effectivePath) {
		effectivePath = filepath.Join(wsPath, effectivePath)
	}

	result, err := s.ValidateAgents(ctx, handle, effectivePath)
	if err != nil {
		l.Error("validation error", "path", *path, "error", err)
		r.ExitFunc(1)
		return
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(result, "", "  ")
		logger.SafeFprintln(r.Stdout, string(data))
	} else {
		if result.Valid {
			l.Success("validation passed", "explanation", result.Explanation)
		} else {
			l.Error("validation failed", "explanation", result.Explanation)
			for _, e := range result.Errors {
				logger.SafeFprintf(r.Stderr, "  - %s\n", e.Message)
			}
		}
	}

	if !result.Valid {
		r.ExitFunc(1)
	}
}
