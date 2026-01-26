package cli

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func isCaptureID(s string) bool {
	return len(s) >= 26 && !strings.Contains(s, " ")
}

func (r *Runner) Apply(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	nameFlag := fs.String("name", "", "Capture name to apply (alternative to providing capture ID)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed apply [<handle>] <capture-id>\n")
		logger.SafeFprintf(r.Stderr, "       workshed apply [<handle>] --name <capture-name>\n\n")
		logger.SafeFprintf(r.Stderr, "Apply a captured git state to all repositories in a workspace.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed apply 01HVABCDEFG\n")
		logger.SafeFprintf(r.Stderr, "  workshed apply --name \"Before refactor\"\n")
		logger.SafeFprintf(r.Stderr, "  workshed apply my-workspace --name \"Starting point\"\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
		return
	}

	ctx := context.Background()

	providedHandle := ""
	argIdx := 0
	if fs.NArg() >= 1 && !isCaptureID(fs.Arg(0)) && fs.Arg(0) != "--name" {
		providedHandle = fs.Arg(0)
		argIdx = 1
	}
	handle := r.ResolveHandle(ctx, providedHandle, l)

	s := r.getStore()

	captureID := ""
	if *nameFlag != "" {
		captures, err := s.ListCaptures(ctx, handle)
		if err != nil {
			l.Error("failed to list captures", "error", err)
			r.ExitFunc(1)
			return
		}
		found := false
		for _, c := range captures {
			if c.Name == *nameFlag {
				captureID = c.ID
				found = true
				break
			}
		}
		if !found {
			l.Error("capture not found", "name", *nameFlag)
			r.ExitFunc(1)
			return
		}
	} else if fs.NArg() > argIdx {
		captureID = fs.Arg(argIdx)
	} else {
		l.Error("missing required argument: <capture-id>")
		fs.Usage()
		r.ExitFunc(1)
		return
	}

	capture, err := s.GetCapture(ctx, handle, captureID)
	if err != nil {
		l.Error("failed to get capture", "handle", handle, "capture", captureID, "error", err)
		r.ExitFunc(1)
		return
	}

	if err := s.ApplyCapture(ctx, handle, captureID); err != nil {
		l.Error("apply failed", "handle", handle, "capture", captureID, "error", err)
		r.ExitFunc(1)
		return
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(capture, "", "  ")
		logger.SafeFprintln(r.Stdout, string(data))
	} else {
		l.Success("applied capture", "id", captureID, "name", capture.Name, "repos", len(capture.GitState))
	}
}
