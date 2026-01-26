package cli

import (
	"context"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
	flag "github.com/spf13/pflag"
)

const defaultCloneTimeout = 5 * time.Minute

func (r *Runner) Create(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("create", flag.ExitOnError)
	purpose := fs.String("purpose", "", "Purpose of the workspace (required)")
	var repoFlags []string
	fs.StringSliceVarP(&repoFlags, "repo", "r", nil, "Repository URL with optional ref (format: url@ref). Can be specified multiple times.")
	var reposAlias []string
	fs.StringSliceVarP(&reposAlias, "repos", "", nil, "Alias for --repo (can be specified multiple times)")
	template := fs.String("template", "", "Template directory to copy into workspace")
	var templateVars []string
	fs.StringSliceVar(&templateVars, "map", nil, "Template variable substitution (format: key=value). Can be specified multiple times.")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed create --purpose <purpose> [--repo url@ref]... [--template <dir>] [--map key=value]...\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		r.ExitFunc(1)
	}

	if *purpose == "" {
		l.Error("missing required flag", "flag", "--purpose")
		fs.Usage()
		r.ExitFunc(1)
	}

	s := r.getStore()
	ctx := context.Background()

	templateVarsMap := make(map[string]string)
	for _, kv := range templateVars {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			l.Error("invalid template variable format", "expected", "key=value", "got", kv)
			r.ExitFunc(1)
			return
		}
		templateVarsMap[parts[0]] = parts[1]
	}

	repos := repoFlags
	if len(reposAlias) > 0 {
		repos = reposAlias
	}

	repoOpts := make([]workspace.RepositoryOption, 0)
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

	opts := workspace.CreateOptions{
		Purpose:      *purpose,
		Template:     *template,
		TemplateVars: templateVarsMap,
		Repositories: repoOpts,
	}

	createCtx, cancel := context.WithTimeout(ctx, defaultCloneTimeout)
	defer cancel()

	ws, err := s.Create(createCtx, opts)
	if err != nil {
		l.Error("workspace creation failed", "purpose", opts.Purpose, "error", err)
		r.ExitFunc(1)
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
	if strings.HasPrefix(repo, "git@") {
		colonIdx := strings.Index(repo, ":")
		if colonIdx != -1 {
			atIdx := strings.LastIndex(repo[colonIdx:], "@")
			if atIdx != -1 {
				actualIdx := colonIdx + atIdx
				url = repo[:actualIdx]
				ref = repo[actualIdx+1:]
				return url, ref
			}
		}
		return repo, ""
	}

	atIdx := strings.LastIndex(repo, "@")
	if atIdx != -1 {
		url = repo[:atIdx]
		ref = repo[atIdx+1:]
	} else {
		url = repo
	}

	return url, ref
}
