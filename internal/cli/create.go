package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/frodi/workshed/internal/workspace"
)

// Create creates a new workspace with the specified purpose and optional repository.
func Create(args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	purpose := fs.String("purpose", "", "Purpose of the workspace (required)")
	repo := fs.String("repo", "", "Repository URL with optional ref (format: url@ref)")

	fs.Usage = func() {
		fmt.Fprintf(errWriter, "Usage: workshed create --purpose <purpose> [--repo url@ref]\n\n")
		fmt.Fprintf(errWriter, "Flags:\n")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	if *purpose == "" {
		fmt.Fprintf(errWriter, "Error: --purpose is required\n\n")
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
		fmt.Fprintf(errWriter, "Error: %v\n", err)
		exitFunc(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ws, err := store.Create(ctx, opts)
	if err != nil {
		fmt.Fprintf(errWriter, "Error creating workspace: %v\n", err)
		exitFunc(1)
		return
	}

	fmt.Fprintf(outWriter, "Created workspace: %s\n", ws.Handle)
	fmt.Fprintf(outWriter, "Path: %s\n", ws.Path)
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
