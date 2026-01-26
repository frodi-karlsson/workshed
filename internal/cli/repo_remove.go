package cli

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/frodi/workshed/internal/logger"
	flag "github.com/spf13/pflag"
)

func (r *Runner) RepoRemove(args []string) {
	l := r.getLogger()

	fs := flag.NewFlagSet("repo remove", flag.ExitOnError)
	force := fs.Bool("force", false, "Skip confirmation prompt")
	var repoName string
	fs.StringVarP(&repoName, "repo", "r", "", "Repository name to remove")

	fs.Usage = func() {
		logger.SafeFprintf(r.Stderr, "Usage: workshed repo remove <handle> --repo <name> [--force]\n\n")
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

	if repoName == "" {
		l.Error("missing --repo flag (repository name required)")
		fs.Usage()
		r.ExitFunc(1)
		return
	}

	s := r.getStore()
	ctx := context.Background()

	ws, err := s.Get(ctx, handle)
	if err != nil {
		l.Error("workspace not found", "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	repo := ws.GetRepositoryByName(repoName)
	if repo == nil {
		l.Error("repository not found", "repo", repoName, "workspace", handle)
		r.ExitFunc(1)
		return
	}

	if !*force {
		prompt := fmt.Sprintf("Remove repository %q from workspace %q? [y/N]: ", repoName, ws.Handle)
		if _, err := fmt.Fprint(r.Stdout, prompt); err != nil {
			l.Error("failed to write prompt", "error", err)
			r.ExitFunc(1)
			return
		}

		reader := bufio.NewReader(r.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			l.Error("failed to read user input", "error", err)
			r.ExitFunc(1)
			return
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			l.Info("operation cancelled")
			return
		}
	}

	if err := s.RemoveRepository(ctx, handle, repoName); err != nil {
		l.Error("failed to remove repository", "repo", repoName, "handle", handle, "error", err)
		r.ExitFunc(1)
		return
	}

	l.Success("repository removed", "handle", handle, "repo", repoName)
}
