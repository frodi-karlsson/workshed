package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion",
		Long: `Generate shell completion scripts.

Examples:
  workshed completion --shell bash >> ~/.bash_completion
  workshed completion --shell zsh > _workshed`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			shell, _ := cmd.Flags().GetString("shell")
			return generateCompletion(shell, os.Stdout)
		},
	}

	cmd.Flags().String("shell", "bash", "Shell type (bash, zsh, fish)")

	return cmd
}

func generateCompletion(shell string, out *os.File) error {
	switch shell {
	case "bash":
		return rootCmd.GenBashCompletion(out)
	case "zsh":
		return rootCmd.GenZshCompletion(out)
	case "fish":
		return rootCmd.GenFishCompletion(out, true)
	default:
		return fmt.Errorf("unsupported shell: %q (supported: bash, zsh, fish)", shell)
	}
}

var rootCmd *cobra.Command

func SetRootCommand(cmd *cobra.Command) {
	rootCmd = cmd
}
