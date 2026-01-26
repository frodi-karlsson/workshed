package tui

import (
	"context"
	"os"

	"github.com/frodi/workshed/internal/store"
	"github.com/frodi/workshed/internal/workspace"
)

func RunDashboard(ctx context.Context, s store.Store, invocationCtx workspace.InvocationContext) error {
	return RunStackModel(ctx, s, invocationCtx)
}

func IsHumanMode() bool {
	envFormat := os.Getenv("WORKSHED_LOG_FORMAT")
	return envFormat == "" || envFormat == "human"
}
