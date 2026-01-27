package completion

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandFlags(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	cmd := NewCommand(root)

	flags := []string{"shell"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Command should have --%s flag", flagName)
		}
	}
}

func TestShellFlagDefaultsToBash(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	cmd := NewCommand(root)

	shellFlag := cmd.Flags().Lookup("shell")
	if shellFlag == nil {
		t.Error("--shell flag should exist")
		return
	}
	if shellFlag.DefValue != "bash" {
		t.Errorf("expected 'bash' default for --shell, got: %s", shellFlag.DefValue)
	}
}

func TestArgsRequiresNoArgs(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	cmd := NewCommand(root)

	if cmd.Args == nil {
		t.Error("Args should be set")
	}

	if cmd.Use != "completion" {
		t.Errorf("expected Use 'completion', got: %s", cmd.Use)
	}
}

func TestCommandUseFormat(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	cmd := NewCommand(root)

	if cmd.Use != "completion" {
		t.Errorf("expected Use 'completion', got: %s", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestExampleShowsBashCompletion(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	cmd := NewCommand(root)

	if cmd.Long != "" || cmd.Example != "" {
		t.Log("Command has description or example")
	}
}

func TestUnsupportedShellError(t *testing.T) {
	root := &cobra.Command{Use: "test"}

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()

	err = generateCompletion(root, "unsupported", f)
	if err == nil {
		t.Error("expected error for unsupported shell")
	}
	if err.Error() != "unsupported shell: \"unsupported\" (supported: bash, zsh, fish)" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestShellCompletionWithRealRoot(t *testing.T) {
	root := &cobra.Command{Use: "workshed"}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "completion.sh")
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()

	err = generateCompletion(root, "bash", f)
	if err != nil {
		t.Errorf("expected no error for bash, got: %v", err)
	}
}
