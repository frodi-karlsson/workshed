# Comprehensive CLI Test Suite Plan

## Goal

Add end-to-end CLI tests that verify commands work correctly with various argument combinations, produce correct output, and handle error cases gracefully.

## Current Coverage Assessment

### What IS Tested

| Category | Coverage |
|----------|----------|
| **Command registration** | `cli_test.go` - verifies all commands exist |
| **Flag existence** | `unit_test.go` - comprehensive flag checks for all commands |
| **Flag defaults** | `unit_test.go` - format defaults for create, list, exec |
| **Argument validators** | `unit_test.go` - commands have Args defined |
| **Handle extraction** | `handle_parsing_test.go` - 8 test cases for `ExtractHandleFromArgs` |
| **Flag parsing** | `unit_test.go` - purpose, repo, name, filter flags |
| **Update command behavior** | `update/update_test.go` - purpose flag parsing |
| **Workspace store** | `internal/workspace/integration_test.go` - create, list, get, remove |

### What IS NOT Tested

| Gap | Impact |
|-----|--------|
| **No CLI integration tests** | Commands are never invoked end-to-end with real arguments |
| **No handle + flag combinations** | e.g., `captures my-workspace --filter x` not tested |
| **No error scenarios** | Invalid handles, missing required args, invalid formats |
| **No exec command tests** | No tests for command execution with/without handle |
| **No output format tests** | JSON/raw/table outputs never validated |
| **No repos subcommand tests** | `repos add`, `repos remove`, `repos list` only checked for flags |
| **No capture/apply tests** | Capture creation, apply operations not tested at CLI level |

---

## Phase 1: Create CLI Test Environment Helper

**New file:** `internal/cli/clitest/clitest.go`

```go
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
    T       *testing.T
    Root    string
    Store   *workspace.FSStore
    OutBuf  bytes.Buffer
    ErrBuf  bytes.Buffer
    Ctx     context.Context
}

func NewCLIEnv(t *testing.T) *CLIEnv {
    root := t.TempDir()
    store, err := workspace.NewFSStore(root)
    if err != nil {
        t.Fatalf("NewFSStore failed: %v", err)
    }
    return &CLIEnv{
        T:     t,
        Root:  root,
        Store: store,
        Ctx:   context.Background(),
    }
}

func (e *CLIEnv) Cleanup() {
    // Cleanup if needed
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
        repos = []workspace.RepositoryOption{
            {URL: "https://github.com/test/repo", Ref: "main"},
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
```

---

## Phase 2: Add Tests for All Commands

**New file:** `internal/cli/clitest/commands_test.go`

### Command Test Matrix

| Command | Test Cases |
|---------|-----------|
| **captures** | no args, valid handle, filter flag, invalid handle |
| **health** | no args, valid handle, invalid handle |
| **inspect** | no args, valid handle, invalid handle |
| **path** | no args (current dir), valid handle, invalid handle, format=json/raw |
| **export** | no args, valid handle, format=json/raw, --output flag |
| **remove** | no args (confirmation), -y flag, valid handle, invalid handle, --dry-run |
| **update** | valid handle + --purpose, missing --purpose, invalid handle |
| **repos list** | no args, valid handle, invalid handle, format options |
| **repos add** | valid handle + --repo, --format options, invalid repo URL |
| **repos remove** | valid handle + --repo, --dry-run, invalid handle |
| **apply** | valid handle + capture-id, --name flag, --dry-run |
| **exec** | -- command only, handle + command, -a flag, format options |
| **create** | --purpose, --repo, multiple repos, --format options |
| **list** | --purpose filter, --page, --page-size, format options |
| **capture** | --name, --description, --tag, format options |
| **import** | --file, --preserve-handle, --force |

### Example Test: Captures Command

```go
func TestCapturesCommand(t *testing.T) {
    env := NewCLIEnv(t)
    defer env.Cleanup()
    
    ws := env.CreateWorkspace("test purpose", nil)
    
    t.Run("no args from workspace directory", func(t *testing.T) {
        err := env.Run(captures.Command(), []string{})
        if err != nil {
            t.Errorf("captures with no args should succeed, got error: %v", err)
        }
        assert.NotContains(t, env.ErrorOutput(), "unknown command")
    })
    
    t.Run("with valid handle", func(t *testing.T) {
        err := env.Run(captures.Command(), []string{ws.Handle})
        if err != nil {
            t.Errorf("captures with valid handle should succeed, got error: %v", err)
        }
        assert.NotContains(t, env.ErrorOutput(), "unknown command")
    })
    
    t.Run("with handle and filter", func(t *testing.T) {
        err := env.Run(captures.Command(), []string{ws.Handle, "--filter", "test"})
        if err != nil {
            t.Errorf("captures with filter should succeed, got error: %v", err)
        }
    })
    
    t.Run("with invalid handle", func(t *testing.T) {
        err := env.Run(captures.Command(), []string{"nonexistent-handle"})
        if err == nil {
            t.Error("captures with invalid handle should fail")
        }
        assert.Contains(t, env.ErrorOutput(), "workspace not found")
    })
}
```

### Example Test: Path Command

```go
func TestPathCommand(t *testing.T) {
    env := NewCLIEnv(t)
    defer env.Cleanup()
    
    ws := env.CreateWorkspace("test purpose", nil)
    
    t.Run("with valid handle returns path", func(t *testing.T) {
        err := env.Run(path.Command(), []string{ws.Handle})
        if err != nil {
            t.Errorf("path with valid handle should succeed: %v", err)
        }
        assert.Contains(t, env.Output(), ws.Handle)
    })
    
    t.Run("with invalid handle", func(t *testing.T) {
        err := env.Run(path.Command(), []string{"nonexistent"})
        if err == nil {
            t.Error("path with invalid handle should fail")
        }
        assert.Contains(t, env.ErrorOutput(), "workspace not found")
    })
    
    t.Run("format json", func(t *testing.T) {
        err := env.Run(path.Command(), []string{ws.Handle, "--format", "json"})
        if err != nil {
            t.Errorf("path --format json should work: %v", err)
        }
    })
    
    t.Run("format raw", func(t *testing.T) {
        err := env.Run(path.Command(), []string{ws.Handle, "--format", "raw"})
        if err != nil {
            t.Errorf("path --format raw should work: %v", err)
        }
        // Raw format should be just the path, no column headers
        assert.NotContains(t, env.Output(), "path")
    })
}
```

---

## Phase 3: Test Output Formats

**New file:** `internal/cli/clitest/output_test.go`

```go
func TestCapturesOutputFormats(t *testing.T) {
    env := NewCLIEnv(t)
    defer env.Cleanup()
    
    ws := env.CreateWorkspace("test", nil)
    
    t.Run("table format has columns", func(t *testing.T) {
        env.Run(captures.Command(), []string{ws.Handle})
        assert.Contains(t, env.Output(), "ID")   // table header
        assert.Contains(t, env.Output(), "NAME")
    })
    
    t.Run("json format is valid", func(t *testing.T) {
        env.Run(captures.Command(), []string{ws.Handle, "--format", "json"})
        var data interface{}
        err := json.Unmarshal([]byte(env.Output()), &data)
        if err != nil {
            t.Errorf("Expected valid JSON output, got: %s", env.Output())
        }
    })
    
    t.Run("raw format is just IDs", func(t *testing.T) {
        env.Run(captures.Command(), []string{ws.Handle, "--format", "raw"})
        assert.NotContains(t, env.Output(), "ID")
        assert.NotContains(t, env.Output(), "NAME")
        assert.Contains(t, env.Output(), "ws_")  // capture IDs start with ws_
    })
}

func TestExportOutputFormats(t *testing.T) {
    env := NewCLIEnv(t)
    defer env.Cleanup()
    
    ws := env.CreateWorkspace("test", nil)
    
    t.Run("export json", func(t *testing.T) {
        env.Run(export.Command(), []string{ws.Handle, "--format", "json"})
        var data map[string]interface{}
        err := json.Unmarshal([]byte(env.Output()), &data)
        if err != nil {
            t.Errorf("Expected valid JSON: %v", err)
        }
        assert.Contains(t, data, "handle")
        assert.Contains(t, data, "purpose")
        assert.Contains(t, data, "repositories")
    })
    
    t.Run("export compact without captures", func(t *testing.T) {
        env.Run(export.Command(), []string{ws.Handle, "--compact"})
        var data map[string]interface{}
        json.Unmarshal([]byte(env.Output()), &data)
        // Compact should exclude captures
    })
}
```

---

## Phase 4: Test Error Handling

**New file:** `internal/cli/clitest/errors_test.go`

```go
func TestInvalidHandleErrors(t *testing.T) {
    tests := []struct {
        name    string
        cmd     *cobra.Command
        args    []string
        wantErr string
    }{
        {"captures invalid", captures.Command(), []string{"nonexistent"}, "workspace not found"},
        {"path invalid", path.Command(), []string{"nonexistent"}, "workspace not found"},
        {"health invalid", health.Command(), []string{"nonexistent"}, "workspace not found"},
        {"inspect invalid", inspect.Command(), []string{"nonexistent"}, "workspace not found"},
        {"export invalid", export.Command(), []string{"nonexistent"}, "workspace not found"},
        {"remove invalid", remove.Command(), []string{"-y", "nonexistent"}, "workspace not found"},
        {"update invalid", update.Command(), []string{"--purpose", "x", "nonexistent"}, "workspace not found"},
        {"repos list invalid", repos.ListCommand(), []string{"nonexistent"}, "workspace not found"},
        {"repos add invalid", repos.AddCommand(), []string{"nonexistent", "--repo", "url"}, "workspace not found"},
        {"repos remove invalid", repos.RemoveCommand(), []string{"nonexistent", "--repo", "url"}, "workspace not found"},
        {"apply invalid", apply.Command(), []string{"nonexistent", "cap-id"}, "workspace not found"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            env := NewCLIEnv(t)
            err := env.Run(tt.cmd, tt.args)
            if err == nil {
                t.Error("Expected error for invalid handle")
            }
            assert.Contains(t, env.ErrorOutput(), tt.wantErr)
        })
    }
}

func TestMissingRequiredFlags(t *testing.T) {
    tests := []struct {
        name string
        cmd  *cobra.Command
        args []string
    }{
        {"capture missing --name", capture.Command(), []string{}},
        {"update missing --purpose", update.Command(), []string{}},
        {"repos add missing --repo", repos.AddCommand(), []string{}},
        {"repos remove missing --repo", repos.RemoveCommand(), []string{}},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            env := NewCLIEnv(t)
            err := env.Run(tt.cmd, tt.args)
            if err == nil {
                t.Error("Expected error for missing required flag")
            }
            assert.Contains(t, env.ErrorOutput(), "required")
        })
    }
}
```

---

## Phase 5: Test Exec Command Specifically

**New file:** `internal/cli/clitest/exec_test.go`

```go
func TestExecCommand(t *testing.T) {
    env := NewCLIEnv(t)
    defer env.Cleanup()
    
    ws := env.CreateWorkspace("test", nil)
    
    t.Run("command without handle from workspace directory", func(t *testing.T) {
        // This should work - uses current workspace directory
        err := env.Run(exec.Command(), []string{"pwd"})
        // If not in a workspace, should get clear error
        if err != nil {
            assert.Contains(t, env.ErrorOutput(), "not in a workspace directory")
        }
    })
    
    t.Run("command with handle", func(t *testing.T) {
        err := env.Run(exec.Command(), []string{ws.Handle, "pwd"})
        if err != nil {
            t.Errorf("exec with valid handle should succeed, got: %v", err)
        }
        assert.NotContains(t, env.ErrorOutput(), "unknown command")
    })
    
    t.Run("command with -a flag", func(t *testing.T) {
        err := env.Run(exec.Command(), []string{"-a", "pwd"})
        if err != nil {
            t.Errorf("exec with -a should work: %v", err)
        }
    })
    
    t.Run("missing command error", func(t *testing.T) {
        err := env.Run(exec.Command(), []string{})
        if err == nil {
            t.Error("exec with no command should fail")
        }
        assert.Contains(t, env.ErrorOutput(), "missing command")
    })
    
    t.Run("format json", func(t *testing.T) {
        env.Run(exec.Command(), []string{ws.Handle, "pwd", "--format", "json"})
        var results []ExecResultOutput
        json.Unmarshal([]byte(env.Output()), &results)
        assert.Len(t, results, 1)
    })
}
```

---

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `internal/cli/clitest/clitest.go` | Create | Test environment helper |
| `internal/cli/clitest/commands_test.go` | Create | Command-specific tests |
| `internal/cli/clitest/output_test.go` | Create | Format validation tests |
| `internal/cli/clitest/errors_test.go` | Create | Error handling tests |
| `internal/cli/clitest/exec_test.go` | Create | Exec command tests |
| `cmd/workshed/unit_test.go` | Modify | Remove redundant tests |

---

## Running the Tests

```bash
# Run all CLI tests
go test ./internal/cli/clitest/... -v

# Run specific test
go test ./internal/cli/clitest/... -run TestCapturesCommand -v

# Run with coverage
go test ./internal/cli/clitest/... -coverprofile coverage.out
```

---

## Estimated Effort

| Phase | Effort |
|-------|--------|
| Phase 1: CLI Test Environment Helper | 1-2 hours |
| Phase 2: Command Tests (14 commands) | 4-6 hours |
| Phase 3: Output Format Tests | 2 hours |
| Phase 4: Error Handling Tests | 2 hours |
| Phase 5: Exec Command Tests | 1-2 hours |

**Total: ~10-15 hours** of development time

---

## Expected Output

A new `internal/cli/clitest` package providing:

- Reusable `CLIEnv` test environment for all CLI tests
- 50+ test cases covering all command argument shapes
- Comprehensive error case coverage with clear error messages
- Output format validation (table, json, raw)
- Pattern for future CLI test additions

---

## Test Quality Guidelines

From AGENTS.md:

> Tests for confidence, not coverage. Easy to read, easy to delete.

### Principles Applied

1. **Clear test names** - `TestCapturesCommand_withValidHandle`
2. **Table-driven where appropriate** - reduces boilerplate
3. **Single assertion per subtest** - easier debugging
4. **No complex setups** - use `NewCLIEnv()` helper
5. **Delete-able tests** - if a test is flaky or unclear, remove it
6. **Focus on user-facing behavior** - what the CLI actually does

---

## See Also

- `/cmd/workshed/unit_test.go` - existing command validation tests
- `/cmd/workshed/handle_parsing_test.go` - handle extraction tests
- `/internal/workspace/integration_test.go` - store integration tests
- `/internal/workspace/testutil.go` - test utilities for workspace creation