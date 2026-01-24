package cli

import (
	"context"
	"flag"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/logger"
	"github.com/frodi/workshed/internal/workspace"
)

const defaultCloneTimeout = 5 * time.Minute

// Create creates a new workspace with the specified purpose and optional repository.
func Create(args []string) {
	l := logger.NewLogger(logger.INFO, "create")

	fs := flag.NewFlagSet("create", flag.ExitOnError)
	purpose := fs.String("purpose", "", "Purpose of the workspace (required)")
	repo := fs.String("repo", "", "Repository URL with optional ref (format: url@ref)")

	fs.Usage = func() {
		logger.SafeFprintf(errWriter, "Usage: workshed create --purpose <purpose> [--repo url@ref]\n\n")
		logger.SafeFprintf(errWriter, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		l.Error("failed to parse flags", "error", err)
		exitFunc(1)
	}

	if *purpose == "" {
		l.Error("missing required flag", "flag", "--purpose")
		fs.Usage()
		exitFunc(1)
	}

	opts := workspace.CreateOptions{
		Purpose: *purpose,
	}

	if *repo != "" {
		url, ref := parseRepoFlag(*repo)
		opts.RepoURL = url
		opts.RepoRef = ref
	}

	store, err := workspace.NewFSStore(GetWorkshedRoot())
	if err != nil {
		l.Error("failed to create workspace store", "error", err)
		exitFunc(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultCloneTimeout)
	defer cancel()

	ws, err := store.Create(ctx, opts)
	if err != nil {
		l.Error("workspace creation failed", "purpose", opts.Purpose, "error", err)
		exitFunc(1)
	}

	l.Success("workspace created", "handle", ws.Handle, "path", ws.Path)
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
