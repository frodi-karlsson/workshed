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
	if len(os.Args) < 2 {
		cli.RunMainDashboard()
		return
	}

	command := os.Args[1]

	switch command {
	case "create":
		cli.Create(os.Args[2:])
	case "list":
		cli.List(os.Args[2:])
	case "inspect":
		cli.Inspect(os.Args[2:])
	case "path":
		cli.Path(os.Args[2:])
	case "exec":
		cli.Exec(os.Args[2:])
	case "remove":
		cli.Remove(os.Args[2:])
	case "update":
		cli.Update(os.Args[2:])
	case "dashboard":
		cli.RunMainDashboard()
	case "version", "-v", "--version":
		cli.Version()
	case "help", "-h", "--help":
		cli.Usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		cli.Usage()
		os.Exit(1)
	}
}
