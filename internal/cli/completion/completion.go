package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewCommand(root *cobra.Command) *cobra.Command {
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
			return generateCompletion(root, shell, os.Stdout)
		},
	}

	cmd.Flags().String("shell", "bash", "Shell type (bash, zsh, fish)")

	return cmd
}

func generateCompletion(root *cobra.Command, shell string, out *os.File) error {
	switch shell {
	case "bash":
		return root.GenBashCompletion(out)
	case "zsh":
		return root.GenZshCompletion(out)
	case "fish":
		return root.GenFishCompletion(out, true)
	default:
		return fmt.Errorf("unsupported shell: %q (supported: bash, zsh, fish)", shell)
	}
}
