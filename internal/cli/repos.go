package cli

import (
	"context"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

func (r *Runner) Repos(args []string) {
	if len(args) < 1 {
		r.ReposUsage()
		r.ExitFunc(1)
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "add":
		r.ReposAdd(args[1:])
	case "remove":
		r.ReposRemove(args[1:])
	case "help", "-h", "--help":
		r.ReposUsage()
	default:
		logger.SafeFprintf(r.Stderr, "Unknown repos subcommand: %s\n\n", subcommand)
		logger.SafeFprintf(r.Stderr, "Use a workspace handle, or run from within a workspace directory:\n")
		logger.SafeFprintf(r.Stderr, "  workshed repos add [<handle>] --repo <url>\n")
		logger.SafeFprintf(r.Stderr, "  workshed repos remove [<handle>] --repo <name>\n\n")
		logger.SafeFprintf(r.Stderr, "Or cd into a workspace and omit the handle:\n")
		logger.SafeFprintf(r.Stderr, "  cd $(workshed path)\n")
		logger.SafeFprintf(r.Stderr, "  workshed repos add --repo <url>\n\n")
		r.ExitFunc(1)
	}
}

func (r *Runner) ReposUsage() {
	msg := `workshed repos - Manage repositories in a workspace

Usage:
  workshed repos add [<handle>] --repo url[@ref]...
  workshed repos remove [<handle>] --repo <name>

Subcommands:
  add     Add repositories to a workspace
  remove  Remove a repository from a workspace

Examples:
  workshed repos add --repo https://github.com/org/repo@main

  workshed repos remove --repo my-repo

  # From within a workspace:
  cd $(workshed path)
  workshed repos add --repo https://github.com/org/repo@main
`
	logger.SafeFprintf(r.Stderr, "%s\n", msg)
}

func (r *Runner) ReposAdd(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("repos add", flag.ExitOnError)
	var repoFlags []string
	fs.StringSliceVarP(&repoFlags, "repo", "r", nil, "Repository URL with optional ref (format: url@ref). Can be specified multiple times.")
	var reposAlias []string
	fs.StringSliceVarP(&reposAlias, "repos", "", nil, "Alias for --repo (can be specified multiple times)")
	format := fs.String("format", "table", "Output format (table|json)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed repos add [<handle>] --repo url[@ref]... [flags]\n\n")
		logger.SafeFprintf(r.Stderr, "Add repositories to a workspace.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed repos add --repo github.com/org/new-repo@main\n")
		logger.SafeFprintf(r.Stderr, "  workshed repos add -r github.com/org/repo1 -r github.com/org/repo2\n")
		logger.SafeFprintf(r.Stderr, "  workshed repos add my-workspace --repo ./local-lib\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
		return
	}

	if err := ValidateFormat(Format(*format), "repos"); err != nil {
		l.Error(err.Error())
		fs.Usage()
		r.ExitFunc(1)
		return
	}

	handle := r.ResolveHandle(context.Background(), fs.Arg(0), true, l)
	if handle == "" {
		return
	}

	repos := repoFlags
	if len(reposAlias) > 0 {
		repos = append(repos, reposAlias...)
	}

	if len(repos) == 0 {
		l.Error("at least one --repo flag is required")
		fs.Usage()
		r.ExitFunc(1)
		return
	}

	var repoOpts []workspace.RepositoryOption
	for _, repo := range repos {
		repo = strings.TrimSpace(repo)
		if repo == "" {
			continue
		}
		url, ref := workspace.ParseRepoFlag(repo)
		repoOpts = append(repoOpts, workspace.RepositoryOption{
			URL: url,
			Ref: ref,
		})
	}

	s := r.getStore()
	ctx := context.Background()

	addCtx, cancel := context.WithTimeout(ctx, defaultCloneTimeout*time.Duration(len(repoOpts)+1))
	defer cancel()

	if err := s.AddRepositories(addCtx, handle, repoOpts, r.InvocationCWD); err != nil {
		l.Error("failed to add repository", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	var rows [][]string
	rows = append(rows, []string{"handle", handle})
	for _, opt := range repoOpts {
		var repoInfo string
		if opt.Ref != "" {
			repoInfo = opt.URL + " @ " + opt.Ref
		} else {
			repoInfo = opt.URL
		}
		rows = append(rows, []string{"repo", repoInfo})
	}

	output := Output{
		Columns: []ColumnConfig{
			{Type: Rigid, Name: "KEY", Min: 10, Max: 20},
			{Type: Rigid, Name: "VALUE", Min: 20, Max: 0},
		},
		Rows: rows,
	}

	if err := r.getOutputRenderer().Render(output, Format(*format), r.Stdout); err != nil {
		l.Error("failed to render output", "error", err)
	}
}

func (r *Runner) ReposRemove(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("repos remove", flag.ExitOnError)
	repoName := fs.String("repo", "", "Repository name to remove")
	format := fs.String("format", "table", "Output format (table|json)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed repos remove [<handle>] --repo <name> [flags]\n\n")
		logger.SafeFprintf(r.Stderr, "Remove a repository from a workspace.\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed repos remove --repo my-repo\n")
		logger.SafeFprintf(r.Stderr, "  workshed repos remove my-workspace --repo my-repo\n")
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
		return
	}

	if err := ValidateFormat(Format(*format), "repos"); err != nil {
		l.Error(err.Error())
		fs.Usage()
		r.ExitFunc(1)
		return
	}

	handle := r.ResolveHandle(context.Background(), fs.Arg(0), true, l)
	if handle == "" {
		return
	}

	if *repoName == "" {
		l.Error("--repo flag is required")
		fs.Usage()
		r.ExitFunc(1)
		return
	}

	s := r.getStore()
	ctx := context.Background()

	if err := s.RemoveRepository(ctx, handle, *repoName); err != nil {
		l.Error("failed to remove repository", "handle", handle, "repo", *repoName, "error", err)
		r.ExitFunc(1)
		return
	}

	output := Output{
		Columns: []ColumnConfig{
			{Type: Rigid, Name: "KEY", Min: 10, Max: 20},
			{Type: Rigid, Name: "VALUE", Min: 20, Max: 0},
		},
		Rows: [][]string{
			{"handle", handle},
			{"repo", *repoName},
		},
	}

	if err := r.getOutputRenderer().Render(output, Format(*format), r.Stdout); err != nil {
		l.Error("failed to render output", "error", err)
	}
}
