# CLI Migration Fixes Plan

Date: 2026-01-27

## Overview

Post-migration fixes required before merging the Cobra-based CLI refactor.

## Critical Issues

### 1. Missing --repos Alias Flag

**Location**: `cmd/workshed/create/create.go`

**Problem**: Old CLI supported `--repos` as an alias for `--repo`. This alias is missing in the new implementation.

**Impact**: User scripts using `--repos` will break.

**Fix**: Add `--repos` flag and merge with `repos` before processing.

### 2. Completion Shell Default Missing

**Location**: `cmd/workshed/completion/completion.go`

**Problem**: Old code had `shell: "bash"` as default. New code has no default, causing silent failure.

**Fix**: Set default to "bash".

### 3. Duplicated Functions

**Problem**: Helper functions are defined in multiple places.

| Function | Locations |
|----------|-----------|
| `isCaptureID()` | `apply.go:17`, `runner.go:211` |
| `preflightErrorHint()` | `apply.go:131`, `runner.go:215` |
| `matchesCaptureFilter()` | `captures.go:17`, `runner.go:232` |

**Fix**: Keep only in `runner.go`, remove from command files.

### 4. Unused Code in runner.go

**Problem**: Dead code retained from old implementation.

- `RunMainDashboard()` - dashboard feature removed
- `Usage()` - Cobra handles help automatically
- `getOutputRenderer()` and `OutputRenderer` interface - not used

**Fix**: Remove unused code.

### 5. No Test Coverage

**Problem**: All test files deleted with no replacement.

**Fix**: Create new test files:
- `cmd/workshed/cli_test.go` - Command setup tests
- `cmd/workshed/create/create_test.go` - Flag parsing tests
- `cmd/workshed/exec/exec_test.go` - Separator handling tests

## Implementation Tasks

- [x] Fix create.go --repos alias
- [x] Fix completion default shell
- [x] Deduplicate isCaptureID() - remove from apply.go
- [x] Deduplicate preflightErrorHint() - remove from apply.go
- [x] Deduplicate matchesCaptureFilter() - remove from captures.go
- [x] Remove unused RunMainDashboard() from runner.go
- [x] Remove unused Usage() from runner.go
- [x] Remove unused getOutputRenderer() and OutputRenderer interface from runner.go
- [x] Remove OutputRendererImpl from runner.go
- [x] Create cmd/workshed/cli_test.go
- [x] Create cmd/workshed/create/create_test.go
- [x] Create cmd/workshed/exec/exec_test.go

## Verification

After implementation:
```bash
go build ./cmd/workshed/...
go test ./cmd/workshed/...
workshed create --repos github.com/org/repo  # should work
workshed completion --shell bash | head -5   # should generate bash script
```