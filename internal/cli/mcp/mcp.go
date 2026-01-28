package mcp

import (
	"context"
	"os"

	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/mcp"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run Workshed as an MCP server",
		Long: `Run Workshed as a Model Context Protocol server.

This command starts an MCP server that exposes Workshed functionality
as tools for AI assistants like Claude Desktop and Cursor.

The server communicates over stdin/stdout using JSON-RPC 2.0.

Example:
  workshed mcp

For Claude Desktop or Cursor configuration, add to your MCP config:
{
  "mcpServers": {
    "workshed": {
      "command": "workshed",
      "args": ["mcp"]
    }
  }
}`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = os.Setenv("GIT_TERMINAL_PROMPT", "0")
			r := cli.NewRunner("")
			ctx := context.Background()

			server := mcp.NewServer(r.GetStore())
			return server.Run(ctx)
		},
	}

	return cmd
}
