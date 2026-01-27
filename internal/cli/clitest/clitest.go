package clitest

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/frodi/workshed/internal/workspace"
	"github.com/spf13/cobra"
)

type CLIEnv struct {
	T        *testing.T
	Root     string
	Store    *workspace.FSStore
	OutBuf   bytes.Buffer
	ErrBuf   bytes.Buffer
	Ctx      context.Context
	origRoot string
}

func NewCLIEnv(t *testing.T) *CLIEnv {
	root := t.TempDir()
	store, err := workspace.NewFSStore(root)
	if err != nil {
		t.Fatalf("NewFSStore failed: %v", err)
	}
	origRoot := os.Getenv("WORKSHED_ROOT")
	if err := os.Setenv("WORKSHED_ROOT", root); err != nil {
		t.Fatalf("Setenv WORKSHED_ROOT failed: %v", err)
	}
	return &CLIEnv{
		T:        t,
		Root:     root,
		Store:    store,
		Ctx:      context.Background(),
		origRoot: origRoot,
	}
}

func (e *CLIEnv) Cleanup() {
	if e.origRoot != "" {
		if err := os.Setenv("WORKSHED_ROOT", e.origRoot); err != nil {
			e.T.Errorf("Setenv WORKSHED_ROOT failed: %v", err)
		}
	} else {
		if err := os.Unsetenv("WORKSHED_ROOT"); err != nil {
			e.T.Errorf("Unsetenv WORKSHED_ROOT failed: %v", err)
		}
	}
}

func (e *CLIEnv) Run(cmd *cobra.Command, args []string) error {
	e.OutBuf.Reset()
	e.ErrBuf.Reset()
	cmd.SetArgs(args)
	cmd.SetOut(&e.OutBuf)
	cmd.SetErr(&e.ErrBuf)
	return cmd.Execute()
}

func (e *CLIEnv) Output() string {
	return e.OutBuf.String()
}

func (e *CLIEnv) ErrorOutput() string {
	return e.ErrBuf.String()
}

func (e *CLIEnv) CreateWorkspace(purpose string, repos []workspace.RepositoryOption) *workspace.Workspace {
	if repos == nil {
		localRepo := workspace.CreateLocalGitRepo(e.T, "testrepo", map[string]string{"README.md": "# Test"})
		repos = []workspace.RepositoryOption{
			{URL: localRepo, Ref: "main"},
		}
	}
	opts := workspace.CreateOptions{
		Purpose:      purpose,
		Repositories: repos,
	}
	ws, err := e.Store.Create(e.Ctx, opts)
	if err != nil {
		e.T.Fatalf("Create failed: %v", err)
	}
	return ws
}
