// workshed is a CLI tool for managing intent-scoped local development workspaces.
//
// Usage:
//
//	workshed create --purpose "Debug payment timeout"
//	workshed list
//	workshed inspect aquatic-fish-motion
//	workshed path aquatic-fish-motion
//	workshed remove aquatic-fish-motion
package main

import (
	"fmt"
	"os"

	"github.com/frodi/workshed/internal/cli"
)

func main() {
	r := cli.NewRunner()

	if len(os.Args) < 2 {
		r.RunMainDashboard()
		return
	}

	command := os.Args[1]

	switch command {
	case "create":
		r.Create(os.Args[2:])
	case "list":
		r.List(os.Args[2:])
	case "inspect":
		r.Inspect(os.Args[2:])
	case "path":
		r.Path(os.Args[2:])
	case "exec":
		r.Exec(os.Args[2:])
	case "remove":
		r.Remove(os.Args[2:])
	case "update":
		r.Update(os.Args[2:])
	case "repo":
		r.Repo(os.Args[2:])
	case "dashboard":
		r.RunMainDashboard()
	case "version", "-v", "--version":
		r.Version()
	case "help", "-h", "--help":
		r.Usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		r.Usage()
		os.Exit(1)
	}
}
