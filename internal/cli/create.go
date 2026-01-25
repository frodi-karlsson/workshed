package cli

import (
	"context"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/tui"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

const defaultCloneTimeout = 5 * time.Minute

func Create(args []string) {
	l := logger.NewLogger(logger.INFO, "create")

	fs := flag.NewFlagSet("create", flag.ExitOnError)
	purpose := fs.String("purpose", "", "Purpose of the workspace (required)")
	var repoFlags []string
	fs.StringSliceVarP(&repoFlags, "repo", "r", nil, "Repository URL with optional ref (format: url@ref). Can be specified multiple times.")
	var reposAlias []string
	fs.StringSliceVarP(&reposAlias, "repos", "", nil, "Alias for --repo (can be specified multiple times)")

	fs.Usage = func() {
		logger.SafeFprintf(errWriter, "Usage: workshed create --purpose <purpose> [--repo url@ref]...\n\n")
		logger.SafeFprintf(errWriter, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		exitFunc(1)
	}

	store := GetOrCreateStore(l)
	ctx := context.Background()

	opts := workspace.CreateOptions{
		Repositories: []workspace.RepositoryOption{},
	}

	if *purpose == "" {
		if tui.IsHumanMode() {
			result, err := tui.RunCreateWizard(ctx, store)
			if err != nil {
				l.Help("wizard cancelled")
				exitFunc(0)
				return
			}
			opts.Purpose = result.Purpose
			opts.Repositories = result.Repositories
		} else {
			l.Error("missing required flag", "flag", "--purpose")
			fs.Usage()
			exitFunc(1)
		}
	} else {
		opts.Purpose = *purpose

		repos := repoFlags
		if len(reposAlias) > 0 {
			repos = reposAlias
		}

		for _, repo := range repos {
			repo = strings.TrimSpace(repo)
			if repo == "" {
				continue
			}
			url, ref := parseRepoFlag(repo)
			opts.Repositories = append(opts.Repositories, workspace.RepositoryOption{
				URL: url,
				Ref: ref,
			})
		}
	}

	createCtx, cancel := context.WithTimeout(ctx, defaultCloneTimeout)
	defer cancel()

	ws, err := store.Create(createCtx, opts)
	if err != nil {
		l.Error("workspace creation failed", "purpose", opts.Purpose, "error", err)
		exitFunc(1)
		return
	}

	if len(ws.Repositories) > 0 {
		repoNames := make([]string, len(ws.Repositories))
		for i, repo := range ws.Repositories {
			repoNames[i] = repo.Name
		}
		l.Success("workspace created", "handle", ws.Handle, "path", ws.Path, "repos", strings.Join(repoNames, ", "))
	} else {
		l.Success("workspace created", "handle", ws.Handle, "path", ws.Path)
	}
}

func parseRepoFlag(repo string) (url, ref string) {
	// For SSH URLs (git@host:path@ref), we need to find the last @ after the first :
	// For other URLs, we split on the last @

	if strings.HasPrefix(repo, "git@") {
		// Find the first : which separates host from path
		colonIdx := strings.Index(repo, ":")
		if colonIdx != -1 {
			// Look for @ after the colon
			atIdx := strings.LastIndex(repo[colonIdx:], "@")
			if atIdx != -1 {
				// Found a ref separator
				actualIdx := colonIdx + atIdx
				url = repo[:actualIdx]
				ref = repo[actualIdx+1:]
				return url, ref
			}
		}
		// No ref found, return whole string as URL
		return repo, ""
	}

	// For non-SSH URLs, split on the last @
	atIdx := strings.LastIndex(repo, "@")
	if atIdx != -1 {
		url = repo[:atIdx]
		ref = repo[atIdx+1:]
	} else {
		url = repo
	}

	return url, ref
}
