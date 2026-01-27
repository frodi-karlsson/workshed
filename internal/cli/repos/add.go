package repos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/workspace"
	"github.com/spf13/cobra"
)

func AddCommand() *cobra.Command {
	var repos []string
	var reposAlias []string

	cmd := &cobra.Command{
		Use:   "add [<handle>] --repo url[@ref]...",
		Short: "Add repositories to a workspace",
		Long: `Add repositories to a workspace.

Examples:
  workshed repos add --repo github.com/org/repo@main
  workshed repos add -r github.com/org/repo1 -r github.com/org/repo2
  workshed repos add my-workspace --repo ./local-lib`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := cli.NewRunner("")

			repos = append(repos, reposAlias...)

			if len(repos) == 0 {
				return fmt.Errorf("missing required flag: --repo")
			}

			ctx := context.Background()
			providedHandle, _ := cli.ExtractHandleFromArgs(args)
			handle, err := r.ResolveHandle(ctx, providedHandle, true, r.GetLogger())
			if err != nil {
				return fmt.Errorf("failed to resolve workspace: %w", err)
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

			addCtx, cancel := context.WithTimeout(ctx, defaultCloneTimeout*time.Duration(len(repoOpts)+1))
			defer cancel()

			if err := r.GetStore().AddRepositories(addCtx, handle, repoOpts, r.GetInvocationCWD()); err != nil {
				return fmt.Errorf("failed to add repository: %w", err)
			}

			format := cmd.Flags().Lookup("format").Value.String()
			if format == "raw" {
				for _, opt := range repoOpts {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), opt.URL)
				}
				return nil
			}

			data := map[string]string{"handle": handle}
			for _, opt := range repoOpts {
				if opt.Ref != "" {
					data["repo"] = opt.URL + " @ " + opt.Ref
				} else {
					data["repo"] = opt.URL
				}
			}

			return cli.RenderKeyValue(data, format, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringSliceVarP(&repos, "repo", "r", nil, "Repository URL with optional ref")
	cmd.Flags().StringSliceVar(&reposAlias, "repos", nil, "Alias for --repo (can be specified multiple times)")
	cmd.Flags().String("format", "table", "Output format (table|json|raw)")
	_ = cmd.MarkFlagRequired("repo")

	return cmd
}

const defaultCloneTimeout = 5 * time.Minute
