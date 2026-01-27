# CLI Cleanup Plan

**Date:** 2026-01-27
**Status:** POST_MIGRATION_CLEANUP

After migrating from schema-driven CLI to Cobra, the codebase has some cleanup items.

---

## Inventory Analysis

### Old CLI Code (Still in internal/cli/)

| File | Lines | Purpose | Used By |
|------|-------|---------|---------|
| apply.go | 177 | Apply command | cli_test.go |
| capture.go | 105 | Capture command | cli_test.go |
| captures.go | 162 | Captures list | cli_test.go |
| create.go | 239 | Create command | cli_test.go |
| exec.go | 191 | Exec command | cli_test.go |
| export.go | 108 | Export command | cli_test.go |
| health.go | 166 | Health command | cli_test.go |
| import.go | 131 | Import command | cli_test.go |
| inspect.go | 87 | Inspect command | cli_test.go |
| list.go | 150 | List command | cli_test.go |
| path.go | 81 | Path command | cli_test.go |
| remove.go | 107 | Remove command | cli_test.go |
| repos.go | 309 | Repos command + subcommands | cli_test.go |
| update.go | 73 | Update command | cli_test.go |
| **Subtotal** | **~2100** | | |

### New CLI Code (In cmd/workshed/)

| Directory | Files | Lines | Purpose |
|-----------|-------|-------|---------|
| create/ | create.go | ~150 | Cobra command |
| list/ | list.go | ~100 | Cobra command |
| inspect/ | inspect.go | ~60 | Cobra command |
| path/ | path.go | ~50 | Cobra command |
| exec/ | exec.go | ~120 | Cobra command |
| repos/ | repos.go, add.go, remove.go, list.go | ~200 | Cobra with subcommands |
| captures/ | captures.go | ~100 | Cobra command |
| capture/ | capture.go | ~70 | Cobra command |
| apply/ | apply.go | ~120 | Cobra command |
| export/ | export.go | ~60 | Cobra command |
| importcmd/ | import.go | ~70 | Cobra command |
| remove/ | remove.go | ~80 | Cobra command |
| update/ | update.go | ~50 | Cobra command |
| health/ | health.go | ~100 | Cobra command |
| completion/ | completion.go | ~50 | Cobra command |
| common/ | context.go, output.go | ~150 | Shared utilities |
| main.go | | ~65 | Root command |
| **Subtotal** | | **~1500** | |

### Shared Packages

| File | Lines | Purpose |
|------|-------|---------|
| internal/cli/table.go | 258 | Output rendering (table/json) |
| internal/cli/flags.go | 280 | Flag setup methods |
| internal/cli/runner.go | 212 | Runner core + exported methods |
| internal/cli/testutil.go | 199 | Test utilities |
| cmd/workshed/common/output.go | 110 | Output rendering (new) |
| cmd/workshed/common/context.go | 35 | Context wrapper |

---

## Cleanup Items

### Priority 1: Remove Old Binary

```bash
rm cmd/workshed/workshed
```

### Priority 2: Remove Duplicate Output Rendering

Two implementations of output rendering exist:

1. **internal/cli/table.go** - Used by old commands and tests
2. **cmd/workshed/common/output.go** - Used by new commands

**Option A: Keep internal/cli/table.go, remove cmd/workshed/common/output.go**
- Pros: Tests already use it, no changes to test code
- Cons: New code must import internal/cli

**Option B: Consolidate to cmd/workshed/common/output.go**
- Pros: Single source of truth
- Cons: Need to update tests to use new package or keep table.go for tests

**Recommendation**: Option A. Tests use internal/cli types directly. Keep table.go for backward compatibility, remove cmd/workshed/common/output.go and update new commands to use internal/cli.

### Priority 3: Consolidate Context Creation

Current: `cmd/workshed/common/context.go` creates a new `Runner` for each command just to access:

- `GetLogger()`  
- `GetStore()`
- `ResolveHandle()`

**Current (inefficient):**
```go
func (c *CmdContext) ResolveHandle(...) string {
    r := cli.NewRunner("")  // New runner each time
    return r.ResolveHandle(...)
}
```

**Solution**: Move these as standalone functions to `cmd/workshed/common/` or create the context once in main.go and pass it around.

### Priority 4: Remove Old Command Files After Test Migration

The old command implementations in `internal/cli/*.go` are only used by:

- `cli_test.go` - Unit tests for command logic
- `cli_integration_test.go` - Integration tests

**After migrating tests to use Cobra commands**, these files can be removed:

```
apply.go, capture.go, captures.go, create.go, exec.go, 
export.go, health.go, import.go, inspect.go, list.go, 
path.go, remove.go, repos.go, update.go
```

**However**, some utilities in these files should be preserved:

| Function | From | To Preserve? |
|----------|------|--------------|
| `validateRepoFlag()` | create.go | Yes - used by create command |
| `validateGitUrl()` | create.go | Yes - used by flags.go |
| `matchesCaptureFilter()` | captures.go | Yes - could be useful |
| `isCaptureID()` | apply.go | Yes - could be useful |
| `preflightErrorHint()` | apply.go | Yes - could be useful |

### Priority 5: Remove Dead Flag Methods

`internal/cli/flags.go` contains flag setup methods like `setupCreateFlags()`, `setupListFlags()`, etc. These are only used by the old command implementations in tests.

After removing old commands, these become dead code.

---

## Recommended Cleanup Order

### Step 1: Remove Old Binary
```bash
rm cmd/workshed/workshed
```

### Step 2: Consolidate Output Rendering
1. Delete `cmd/workshed/common/output.go`
2. Update new command files to use `internal/cli.Output`, `internal/cli.ColumnConfig`, `internal/cli.RenderListTable()`
3. May need a small wrapper for `Format` enum handling

### Step 3: Optimize Context Creation
1. Move `GetLogger()`, `GetStore()`, `ResolveHandle()` to standalone functions in `cmd/workshed/common/context.go`
2. Or: Create context once in main.go and pass to commands via command annotations

### Step 4: Delete cmd/workshed/common/
After consolidating output rendering and context, this directory may be empty or have minimal content.

### Step 5: Migrate Tests to Use New Commands (Optional)
Only if tests need to be updated to verify Cobra command behavior.

### Step 6: Delete Old Command Files (Optional)
After tests are migrated or if they're deemed unnecessary:
```bash
rm internal/cli/{apply,capture,captures,create,exec,export,health,import,inspect,list,path,remove,repos,update}.go
```

---

## Files to Remove

| File | Size | Reason |
|------|------|--------|
| cmd/workshed/workshed | 6MB | Old binary |
| cmd/workshed/common/output.go | 110 lines | Duplicate of table.go |
| cmd/workshed/common/context.go | 35 lines | Can be consolidated |
| internal/cli/flags.go | 280 lines | Dead after old commands removed |
| internal/cli/*.go (old commands) | ~2100 lines | Dead after test migration |

---

## Effort Estimate

| Task | Time |
|------|------|
| Remove old binary | 1 min |
| Consolidate output rendering | 30 min |
| Optimize context creation | 15 min |
| Delete common/ directory | 5 min |
| **Total** | **~1 hour** |

The old command files (2,100 lines) can remain if tests depend on them - they're not hurting anything and provide coverage for the Runner's command logic.