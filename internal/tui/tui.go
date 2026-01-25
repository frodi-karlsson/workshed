package tui

import (
	"context"

	"github.com/frodi/workshed/internal/store"
)

func RunDashboard(ctx context.Context, s store.Store) error {
	return RunStackModel(ctx, s)
}
