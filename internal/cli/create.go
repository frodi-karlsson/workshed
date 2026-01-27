package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/git"
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
	format := fs.String("format", "table", "Output format (table|json)")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed create --purpose <purpose> [--repo url@ref]... [--template <dir>] [--map key=value]... [flags]\n\n")
		logger.SafeFprintf(r.Stderr, "Flags:\n")
		fs.PrintDefaults()
		logger.SafeFprintf(r.Stderr, "\nExamples:\n")
		logger.SafeFprintf(r.Stderr, "  workshed create --purpose \"Debug payment timeout\" --repo github.com/org/api@main\n")
		logger.SafeFprintf(r.Stderr, "  workshed create -r github.com/org/frontend@feature -r github.com/org/backend@feature\n")
		logger.SafeFprintf(r.Stderr, "  workshed create --purpose \"New feature\" --template ~/templates/react-app --map name=myapp\n")
		logger.SafeFprintf(r.Stderr, "  workshed create --purpose \"Local exploration\"\n")
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

	if len(repos) == 0 {
		currentURL, err := git.RealGit{}.GetRemoteURL(ctx, ".")
		if err != nil {
			l.Error("no repository specified and not in a git repository with origin", "hint", "specify a repository with --repo url@ref")
			r.ExitFunc(1)
			return
		}
		repoOpts = append(repoOpts, workspace.RepositoryOption{URL: currentURL})
	} else {
		for _, repo := range repos {
			repo = strings.TrimSpace(repo)
			if repo == "" {
				continue
			}

			if err := validateRepoFlag(repo); err != nil {
				l.Error("invalid repository specification", "repo", repo, "error", err)
				r.ExitFunc(1)
				return
			}

			url, ref := workspace.ParseRepoFlag(repo)
			repoOpts = append(repoOpts, workspace.RepositoryOption{
				URL: url,
				Ref: ref,
			})
		}
	}

	opts := workspace.CreateOptions{
		Purpose:       *purpose,
		Template:      *template,
		TemplateVars:  templateVarsMap,
		Repositories:  repoOpts,
		InvocationCWD: r.InvocationCWD,
	}

	createCtx, cancel := context.WithTimeout(ctx, defaultCloneTimeout)
	defer cancel()

	ws, err := s.Create(createCtx, opts)
	if err != nil {
		l.Error("workspace creation failed", "purpose", opts.Purpose, "error", err)
		r.ExitFunc(1)
		return
	}

	var rows [][]string
	rows = append(rows, []string{"handle", ws.Handle})
	rows = append(rows, []string{"path", ws.Path})
	rows = append(rows, []string{"purpose", ws.Purpose})
	for _, repo := range ws.Repositories {
		var repoInfo string
		if repo.Ref != "" {
			repoInfo = repo.Name + " @ " + repo.Ref
		} else {
			repoInfo = repo.Name
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

func validateRepoFlag(repo string) error {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return nil
	}

	atIdx := strings.LastIndex(repo, "@")
	url := repo
	if atIdx != -1 {
		url = repo[:atIdx]
	}

	if strings.HasPrefix(url, "git@") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid git SSH URL format (expected: git@host:path)")
		}
		if parts[1] == "" {
			return fmt.Errorf("missing path in git SSH URL")
		}
		return nil
	}

	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		if url == "https://" || url == "http://" {
			return fmt.Errorf("incomplete URL: missing host")
		}
		return nil
	}

	if strings.HasPrefix(url, "file://") {
		path := strings.TrimPrefix(url, "file://")
		if path == "" {
			return fmt.Errorf("missing path in file:// URL")
		}
		if !filepath.IsAbs(path) {
			return fmt.Errorf("file:// URL must use absolute path")
		}
		return nil
	}

	if strings.Contains(url, "://") {
		return fmt.Errorf("unsupported URL scheme: use https://, git@, or local path")
	}

	if !strings.Contains(url, "/") && !strings.Contains(url, "\\") {
		if _, err := os.Stat(url); err == nil {
			return nil
		}
		if strings.HasSuffix(url, ".git") {
			return fmt.Errorf("repository not found: %s", url)
		}
	}

	if _, err := os.Stat(url); err == nil {
		return nil
	}

	return nil
}
