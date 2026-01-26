package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/frodi/workshed/internal/workspace"
)

func (r *Runner) Health(args []string) {
	l := r.getLogger()
	ctx := context.Background()

	handle := r.ResolveHandle(ctx, "", l)
	if handle == "" {
		return
	}

	s := r.getStore()

	execs, err := s.ListExecutions(ctx, handle, workspace.ListExecutionsOptions{Limit: 100})
	if err != nil {
		l.Error("failed to list executions", "error", err)
		r.ExitFunc(1)
		return
	}

	hasIssues := false
	staleCount := 0

	for _, e := range execs {
		if time.Since(e.Timestamp) > 30*24*time.Hour {
			staleCount++
		}
	}

	if staleCount > 0 {
		hasIssues = true
		fmt.Printf("Issues found:\n\n")
		fmt.Printf("Stale Executions:\n")
		fmt.Printf("  • %d executions older than 30 days\n", staleCount)
	}

	if !hasIssues {
		l.Success("workspace is healthy", "handle", handle)
	} else {
		l.Info("health check completed", "handle", handle, "issues", "found")
	}
}
