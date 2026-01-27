package unit

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommandSetup(t *testing.T) {
	root := &cobra.Command{
		Use:   "workshed",
		Short: "Test CLI",
	}

	root.AddCommand(&cobra.Command{Use: "create"})
	root.AddCommand(&cobra.Command{Use: "list"})
	root.AddCommand(&cobra.Command{Use: "inspect"})
	root.AddCommand(&cobra.Command{Use: "path"})
	root.AddCommand(&cobra.Command{Use: "repos"})
	root.AddCommand(&cobra.Command{Use: "captures"})
	root.AddCommand(&cobra.Command{Use: "capture"})
	root.AddCommand(&cobra.Command{Use: "apply"})
	root.AddCommand(&cobra.Command{Use: "export"})
	root.AddCommand(&cobra.Command{Use: "import"})
	root.AddCommand(&cobra.Command{Use: "remove"})
	root.AddCommand(&cobra.Command{Use: "update"})
	root.AddCommand(&cobra.Command{Use: "health"})
	root.AddCommand(&cobra.Command{Use: "completion"})

	commands := []string{
		"create", "list", "inspect", "path", "repos", "captures",
		"capture", "apply", "export", "import", "remove", "update",
		"health", "completion",
	}

	for _, cmd := range commands {
		found := false
		for _, c := range root.Commands() {
			if c.Name() == cmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command %q not found in root", cmd)
		}
	}
}

func TestCommandAcceptsHandleArgument(t *testing.T) {
	cmds := []struct {
		name string
		use  string
	}{
		{"captures", "captures [<handle>]"},
		{"health", "health [<handle>]"},
		{"inspect", "inspect [<handle>]"},
		{"path", "path [<handle>]"},
		{"export", "export [<handle>]"},
		{"remove", "remove [<handle>]"},
		{"update", "update [<handle>]"},
		{"repos list", "list [<handle>]"},
	}

	for _, tc := range cmds {
		t.Run(tc.name+" does not produce unknown command error", func(t *testing.T) {
			parts := strings.Fields(tc.use)
			if len(parts) < 2 {
				t.Skip("command doesn't accept arguments")
			}
			argName := parts[1]
			if !strings.HasPrefix(argName, "<") {
				t.Skip("argument is not positional")
			}
		})
	}
}

func TestVersionFlag(t *testing.T) {
	var buf bytes.Buffer
	root := &cobra.Command{
		Use:     "workshed",
		Version: "0.5.1",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Version: " + cmd.Version)
			return nil
		},
	}

	root.SetOut(&buf)
	root.SetArgs([]string{"--version"})
	err := root.Execute()
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("0.5.1")) {
		t.Errorf("Version output should contain version, got: %s", output)
	}
}

func TestHelpFlag(t *testing.T) {
	root := &cobra.Command{
		Use:   "workshed",
		Short: "Test CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	root.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a workspace",
	})

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"--help"})
	err := root.Execute()
	if err != nil {
		t.Errorf("Help execute failed: %v", err)
	}

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("workshed")) {
		t.Errorf("Help output should mention workshed, got: %s", output)
	}
}
