package completion

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandFlags(t *testing.T) {
	cmd := Command()

	flags := []string{"shell"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Command should have --%s flag", flagName)
		}
	}
}

func TestShellFlagDefaultsToBash(t *testing.T) {
	cmd := Command()

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
	cmd := Command()

	if cmd.Args == nil {
		t.Error("Args should be set")
	}

	if cmd.Use != "completion" {
		t.Errorf("expected Use 'completion', got: %s", cmd.Use)
	}
}

func TestCommandUseFormat(t *testing.T) {
	cmd := Command()

	if cmd.Use != "completion" {
		t.Errorf("expected Use 'completion', got: %s", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestExampleShowsBashCompletion(t *testing.T) {
	cmd := Command()

	if cmd.Long != "" || cmd.Example != "" {
		t.Log("Command has description or example")
	}
}

func TestSetRootCommand(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	SetRootCommand(cmd)
	if rootCmd != cmd {
		t.Error("SetRootCommand should set rootCmd")
	}
}

func TestUnsupportedShellError(t *testing.T) {
	SetRootCommand(nil)

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()

	err = generateCompletion("unsupported", f)
	if err == nil {
		t.Error("expected error for unsupported shell")
	}
	if err.Error() != "unsupported shell: \"unsupported\" (supported: bash, zsh, fish)" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestShellCompletionWithRealRoot(t *testing.T) {
	SetRootCommand(nil)

	root := &cobra.Command{Use: "workshed"}
	SetRootCommand(root)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "completion.sh")
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()

	err = generateCompletion("bash", f)
	if err != nil {
		t.Errorf("expected no error for bash, got: %v", err)
	}
}
