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
	invocationCWD, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	r := cli.NewRunner(invocationCWD)

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
	case "repos":
		r.Repos(os.Args[2:])
	case "capture":
		r.Capture(os.Args[2:])
	case "apply":
		r.Apply(os.Args[2:])
	case "derive":
		r.Derive(os.Args[2:])
	case "captures":
		r.Captures(os.Args[2:])
	case "health":
		r.Health(os.Args[2:])
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
