package main

import (
	"fmt"
	"os"

	"github.com/frodi/workshed/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		cli.Usage()
		os.Exit(1)
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
	case "remove":
		cli.Remove(os.Args[2:])
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
