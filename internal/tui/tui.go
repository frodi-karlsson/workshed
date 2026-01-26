package tui

import (
	"context"

	"github.com/frodi/workshed/internal/store"
	"github.com/frodi/workshed/internal/workspace"
)

func RunDashboard(ctx context.Context, s store.Store, invocationCtx workspace.InvocationContext) error {
	return RunStackModel(ctx, s, invocationCtx)
}
