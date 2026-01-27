package create

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/git"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const defaultCloneTimeout = 5 * time.Minute

func Command() *cobra.Command {
	var purpose string
	var repos []string
	var reposAlias []string
	var localMap []string
	var template string
	var templateVars []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workspace for a specific task",
		Long: `Create a new workspace for a specific task.

Examples:
  workshed create --purpose "Debug payment timeout" --repo github.com/org/api@main
  workshed create -r github.com/org/frontend@feature -r github.com/org/backend@feature
  workshed create --purpose "New feature" --template ~/templates/react-app --map name=myapp
  workshed create --purpose "Local exploration"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")
			ctx := context.Background()

			isInteractive := term.IsTerminal(int(os.Stdin.Fd()))

			if purpose == "" {
				if isInteractive {
					return fmt.Errorf("missing required flag: --purpose")
				}
				return fmt.Errorf("missing required flag: --purpose")
			}

			repos = append(repos, reposAlias...)

			if len(repos) == 0 && isInteractive {
				fmt.Print("Repository URL (optional, press Enter to use current directory's git remote): ")
				repoInput, err := cli.ReadLine(r.Stdin)
				if err != nil {
					return fmt.Errorf("reading repository: %w", err)
				}
				repoInput = strings.TrimSpace(repoInput)
				if repoInput != "" {
					repos = append(repos, repoInput)
				}
			}

			for _, repo := range repos {
				if err := validateRepoFlag(repo); err != nil {
					return fmt.Errorf("invalid repository %q: %w", repo, err)
				}
			}

			repoOpts := make([]workspace.RepositoryOption, 0)

			if len(repos) == 0 {
				currentURL, err := git.RealGit{}.GetRemoteURL(context.Background(), ".")
				if err != nil {
					return fmt.Errorf("no repository specified and not in a git repository with origin: %w", err)
				}
				repoOpts = append(repoOpts, workspace.RepositoryOption{URL: currentURL})
			} else {
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
			}

			for _, local := range localMap {
				if err := validateLocalMapFlag(local); err != nil {
					return fmt.Errorf("invalid local-map %q: %w", local, err)
				}
				parts := strings.SplitN(local, ":", 2)
				if len(parts) == 2 {
					repoOpts = append(repoOpts, workspace.RepositoryOption{
						URL: "file://" + parts[1],
						Ref: parts[0],
					})
				}
			}

			templateVarsMap := make(map[string]string)
			for _, kv := range templateVars {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid template variable %q (expected key=value)", kv)
				}
				templateVarsMap[parts[0]] = parts[1]
			}

			if template != "" {
				if _, err := os.Stat(template); err != nil {
					return fmt.Errorf("template not found: %s", template)
				}
			}

			opts := workspace.CreateOptions{
				Purpose:       purpose,
				Template:      template,
				TemplateVars:  templateVarsMap,
				Repositories:  repoOpts,
				InvocationCWD: r.GetInvocationCWD(),
			}

			createCtx, cancel := context.WithTimeout(ctx, defaultCloneTimeout)
			defer cancel()

			ws, err := r.GetStore().Create(createCtx, opts)
			if err != nil {
				return fmt.Errorf("workspace creation failed: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()
			if format == "raw" {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), ws.Handle)
				return nil
			}

			data := map[string]string{
				"handle":  ws.Handle,
				"path":    ws.Path,
				"purpose": ws.Purpose,
			}
			for _, repo := range ws.Repositories {
				var repoInfo string
				if repo.Ref != "" {
					repoInfo = repo.Name + " @ " + repo.Ref
				} else {
					repoInfo = repo.Name
				}
				data["repo"] = repoInfo
			}

			return cli.RenderKeyValue(data, format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&purpose, "purpose", "", "Workspace purpose")
	cmd.Flags().StringSliceVarP(&repos, "repo", "r", nil, "Repository URL with optional ref")
	cmd.Flags().StringSliceVar(&reposAlias, "repos", nil, "Alias for --repo (can be specified multiple times)")
	cmd.Flags().StringSliceVar(&localMap, "local-map", nil, "Map a local directory as a repository")
	cmd.Flags().StringVar(&template, "template", "", "Template name or path")
	cmd.Flags().StringSliceVar(&templateVars, "map", nil, "Template variable (key=value)")
	cmd.Flags().String("format", "table", "Output format (table|json)")
	_ = cmd.MarkFlagRequired("purpose")

	return cmd
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

	if strings.HasPrefix(url, "git://") {
		if url == "git://" {
			return fmt.Errorf("incomplete URL: missing host")
		}
		return nil
	}

	if strings.HasPrefix(url, "ssh://") {
		if url == "ssh://" || url == "ssh:///" {
			return fmt.Errorf("incomplete SSH URL: missing host")
		}
		return nil
	}

	if strings.Contains(url, "://") {
		return fmt.Errorf("unsupported URL scheme: use https://, git@, git://, ssh://, or local path")
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

func validateLocalMapFlag(local string) error {
	if !strings.Contains(local, ":") {
		return fmt.Errorf("invalid local-map format (expected name:/path/to/dir)")
	}
	return nil
}
