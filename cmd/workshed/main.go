package main

import (
	"context"
	"os"

	"github.com/frodi/workshed/cmd/workshed/apply"
	"github.com/frodi/workshed/cmd/workshed/capture"
	"github.com/frodi/workshed/cmd/workshed/captures"
	"github.com/frodi/workshed/cmd/workshed/completion"
	"github.com/frodi/workshed/cmd/workshed/create"
	"github.com/frodi/workshed/cmd/workshed/exec"
	"github.com/frodi/workshed/cmd/workshed/export"
	"github.com/frodi/workshed/cmd/workshed/health"
	"github.com/frodi/workshed/cmd/workshed/importcmd"
	"github.com/frodi/workshed/cmd/workshed/inspect"
	"github.com/frodi/workshed/cmd/workshed/list"
	"github.com/frodi/workshed/cmd/workshed/path"
	"github.com/frodi/workshed/cmd/workshed/remove"
	"github.com/frodi/workshed/cmd/workshed/repos"
	"github.com/frodi/workshed/cmd/workshed/update"
	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/tui"
	"github.com/spf13/cobra"
)

var version = "0.4.0"

func main() {
	if len(os.Args) < 2 {
		runDashboard()
		return
	}

	root := &cobra.Command{
		Use:     "workshed",
		Version: version,
		Short:   "Intent-scoped local development workspaces",
		Long: `workshed - Intent-scoped local development workspaces

Create isolated workspaces for specific tasks with their own purpose,
repositories, and state captures.

Examples:
  workshed create --purpose "Debug payment timeout" --repo github.com/org/api@main
  workshed list
  workshed exec -- make test
  workshed capture --name "Before changes"
  workshed apply --name "Before changes"`,
	}

	root.AddCommand(create.Command())
	root.AddCommand(list.Command())
	root.AddCommand(inspect.Command())
	root.AddCommand(path.Command())
	root.AddCommand(repos.Command())
	root.AddCommand(captures.Command())
	root.AddCommand(capture.Command())
	root.AddCommand(apply.Command())
	root.AddCommand(exec.Command())
	root.AddCommand(export.Command())
	root.AddCommand(importcmd.Command())
	root.AddCommand(remove.Command())
	root.AddCommand(update.Command())
	root.AddCommand(health.Command())

	completion.SetRootCommand(root)
	root.AddCommand(completion.Command())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runDashboard() {
	if !tui.IsHumanMode() {
		root := &cobra.Command{Use: "workshed"}
		root.SetArgs([]string{"--help"})
		_ = root.Execute()
		return
	}

	r := cli.NewRunner("")
	ctx := context.Background()
	err := tui.RunDashboard(ctx, r.GetStore(), r)
	if err != nil {
		os.Exit(1)
	}
}
