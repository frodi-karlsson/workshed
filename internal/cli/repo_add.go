package cli

import (
	"context"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

func (r *Runner) RepoAdd(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("repo add", flag.ExitOnError)
	var repoFlags []string
	fs.StringSliceVarP(&repoFlags, "repo", "r", nil, "Repository URL with optional ref (format: url@ref). Can be specified multiple times.")
	var reposAlias []string
	fs.StringSliceVarP(&reposAlias, "repos", "", nil, "Alias for --repo (can be specified multiple times)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed repo add <handle> --repo url[@ref]...\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
		return
	}

	if fs.NArg() < 1 {
		l.Error("missing workspace handle")
		fs.Usage()
		r.ExitFunc(1)
		return
	}

	handle := fs.Arg(0)

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
		url, ref := parseRepoFlag(repo)
		repoOpts = append(repoOpts, workspace.RepositoryOption{
			URL: url,
			Ref: ref,
		})
	}

	s := r.getStore()
	ctx := context.Background()

	addCtx, cancel := context.WithTimeout(ctx, defaultCloneTimeout*time.Duration(len(repoOpts)+1))
	defer cancel()

	if err := s.AddRepositories(addCtx, handle, repoOpts); err != nil {
		l.Error("failed to add repository", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	if len(repoOpts) == 1 {
		l.Success("repository added", "handle", handle, "repo", repoOpts[0].URL)
	} else {
		urls := make([]string, len(repoOpts))
		for i, opt := range repoOpts {
			urls[i] = opt.URL
		}
		l.Success("repositories added", "handle", handle, "repos", strings.Join(urls, ", "))
	}
}
