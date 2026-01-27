package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Captures(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("captures", flag.ExitOnError)
	format := fs.String("format", "table", "Output format (table|json)")
	reverse := fs.Bool("reverse", false, "Reverse sort order (oldest first)")
	filter := fs.String("filter", "", "Filter captures by repository name or branch (case-insensitive substring match)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed captures [<handle>] [--filter <repo|branch>] [flags]\n\n")
		logger.SafeFprintf(r.Stderr, "List captures for a workspace.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed captures\n")
		logger.SafeFprintf(r.Stderr, "  workshed captures my-workspace\n")
		logger.SafeFprintf(r.Stderr, "  workshed captures --filter api\n")
		logger.SafeFprintf(r.Stderr, "  workshed captures --format json\n")
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
	handle := r.ResolveHandle(ctx, providedHandle, true, l)

	s := r.getStore()
	captures, err := s.ListCaptures(ctx, handle)
	if err != nil {
		l.Error("failed to list captures", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	if len(captures) == 0 {
		if *format == "json" {
			logger.SafeFprintln(r.Stdout, "[]")
		} else {
			l.Info("no captures found")
		}
		return
	}

	var filteredCaptures []workspace.Capture
	if *filter != "" {
		filterLower := strings.ToLower(*filter)
		for _, cap := range captures {
			match := false
			if strings.Contains(strings.ToLower(cap.Name), filterLower) {
				match = true
			}
			for _, gitRef := range cap.GitState {
				if strings.Contains(strings.ToLower(gitRef.Repository), filterLower) {
					match = true
				}
				if strings.Contains(strings.ToLower(gitRef.Branch), filterLower) {
					match = true
				}
			}
			if match {
				filteredCaptures = append(filteredCaptures, cap)
			}
		}
	} else {
		filteredCaptures = captures
	}

	if len(filteredCaptures) == 0 {
		if *format == "json" {
			logger.SafeFprintln(r.Stdout, "[]")
		} else {
			l.Info("no captures match filter: " + *filter)
		}
		return
	}

	displayCaptures := filteredCaptures
	if *reverse {
		for i, j := 0, len(displayCaptures)-1; i < j; i, j = i+1, j-1 {
			displayCaptures[i], displayCaptures[j] = displayCaptures[j], displayCaptures[i]
		}
	}

	if *format == "raw" {
		for _, cap := range displayCaptures {
			logger.SafeFprintln(r.Stdout, cap.ID)
		}
		return
	}

	var rows [][]string
	for _, cap := range displayCaptures {
		created := cap.Timestamp.Format("2006-01-02 15:04")
		rows = append(rows, []string{cap.ID, cap.Name, cap.Kind, fmt.Sprintf("%d", len(cap.GitState)), created})
	}

	output := Output{
		Columns: []ColumnConfig{
			{Type: Rigid, Name: "ID", Min: 26, Max: 26},
			{Type: Shrinkable, Name: "NAME", Min: 15, Max: 0},
			{Type: Rigid, Name: "KIND", Min: 8, Max: 15},
			{Type: Rigid, Name: "REPOS", Min: 6, Max: 8},
			{Type: Rigid, Name: "CREATED", Min: 16, Max: 16},
		},
		Rows: rows,
	}

	if err := r.getOutputRenderer().Render(output, Format(*format), r.Stdout); err != nil {
		l.Error("failed to render output", "error", err)
	}
}
