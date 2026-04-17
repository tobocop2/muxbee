# beetrix -- Development Guide

## Project
Single-binary self-hosted unified messenger. Go 1.23+, Cobra CLI, Bubble Tea TUI, embedded Dendrite homeserver, mautrix bridgev2 bridges. Task tracking with `beads` (`bd`). Learned behaviors with `floop`.

**Framing:** Lead with "self-hosted unified messenger" -- not "Matrix wrapper" or "bridge manager." beetrix is a single binary that gives you all your chats in one place.

## Task Tracking (beads)
```bash
bd ready                      # See what's ready to work on
bd update <id> -s in_progress # Claim a task
bd close <id>                 # Mark done
bd list                       # All issues
```
Every code change MUST be tracked as a beads task. Create tasks before starting work, close them when done.

## Commands
```bash
make build                    # Build binary (go build with version ldflags)
make test                     # Tests with race detector + coverage
make lint                     # go vet
make check                    # Run all gates: lint + test (same as CI)
make clean                    # Remove binary
make install                  # go install with version ldflags
make qa                       # Automated QA via tmux send-keys
go test ./...                 # Quick test run
go test -race ./internal/config/... # Test specific package with race detector
./beetrix --version           # Verify build
```

## Architecture
- `cmd/` -- CLI commands (Cobra). Each file is one command.
- `internal/bridges/` -- Bridge registry (embedded YAML, go:embed)
- `internal/config/` -- Settings struct, XDG paths, load/save, token generation
- `internal/homeserver/` -- Dendrite embedding (in-process monolith with SQLite)
- `internal/megabridge/` -- mautrix-go bridgev2 bridge lifecycle manager
- `internal/megabridge/connectors/` -- Per-network connector wrappers
- `internal/server/` -- Orchestrator composing homeserver + megabridge
- `internal/matrix/` -- Matrix client for admin operations and bot setup
- `internal/tui/` -- Terminal UI (Bubble Tea, Elm architecture)
- `main.go` -- Entry point, calls `cmd.Execute()`

Everything under `internal/` is private to the module. Only `cmd/` and `main.go` import these packages.

Leaf packages (no internal dependencies): `config`, `bridges`.

## Configuration
XDG directory layout:
- `~/.config/beetrix/` -- Configuration (settings.yaml)
- `~/.local/share/beetrix/` -- Persistent data (Dendrite DB, bridge databases, signing keys)

Settings override via environment:
- `XDG_CONFIG_HOME` -- Override config base directory
- `XDG_DATA_HOME` -- Override data base directory

All secrets (passwords, tokens) are auto-generated and persisted in `settings.yaml`.

## Code Quality Rules

### Test-Driven Development
- Write tests BEFORE or alongside implementation, not after
- Every exported function MUST have at least one test
- Every test must assert something meaningful -- no zero-assertion coverage padding
- Use `testify/assert` and `testify/require` (project convention)
- Use `t.TempDir()` for filesystem tests -- never write to real paths
- Use `t.Setenv()` for environment variable tests (auto-restored)
- **Table-driven tests** for functions with multiple input/output combinations
- Test edge cases and error paths, not just the happy path
- Tests are documentation -- name them descriptively (`TestConfig_MissingFile_ReturnsDefault`, not `TestConfig3`)
- Never run tests in background -- user must see output
- Kill tests that hang past 2-3 minutes -- investigate the hang, don't wait
- Race detector (`-race`) on every test run

### DRY & Modularity
- Don't Repeat Yourself -- extract shared logic into helpers when duplicated
- Single Responsibility -- each function does one thing well
- Small functions -- max ~20 lines, max 2 levels of nesting
- Low cyclomatic complexity -- extract helpers when branches exceed 3
- Use maps for classification/dispatch instead of long switch chains
- Compose small functions rather than writing monolithic ones
- If you need to copy-paste code, refactor into a shared function instead

### Go Idioms
- **Accept interfaces, return structs** -- function parameters should be interfaces when possible, return types should be concrete
- **Error wrapping** -- use `fmt.Errorf("context: %w", err)` to wrap errors with context. Never swallow errors silently.
- **Error handling** -- check every error. No `_` for error returns unless explicitly justified with a comment.
- **Context propagation** -- pass `context.Context` as the first parameter to functions that do I/O or may be cancelled
- **No init() functions** except for the bridge registry (documented exception). All other initialization must be explicit via constructors.
- **Factory constructors** -- use `New*()` functions, not direct struct initialization. Struct fields are private; interaction is through methods.
- **Structured logging** -- use `log/slog` for new code. No `fmt.Println` for operational output.
- **Godoc comments** on every exported type, function, and constant. Start with the name: `// Config represents the beetrix settings`.
- **go:embed** for static assets (YAML, templates). No runtime file reads for bundled resources.
- **Sentinel errors** -- define package-level `var ErrSomething = errors.New("...")` for errors that callers need to check with `errors.Is`.

### Code Style
- **Named constants for magic numbers AND magic strings** -- never use raw string literals when a constant exists. Grep globally for the raw value of any new constant to ensure no duplicates remain.
- Descriptive variable names -- `bridgeTokens` not `bt`, `configPath` not `p`
- Short variable names acceptable in small scopes (loop vars, receivers)
- Receiver names: one or two letters, consistent across methods (`c` for Config, `m` for Model)
- No hardcoded values -- all configurable through config structs with sensible defaults
- No bare `panic()` in library code -- return errors
- Line length: keep reasonable (~100 chars), break at logical points
- No em dashes in comments or strings -- use periods or commas instead
- No divider comments (`// ------`) -- use separate files or types instead

### Import Discipline
- **Standard library first**, then third-party, then local -- separated by blank lines:
  ```go
  import (
      "fmt"
      "os"

      "github.com/spf13/cobra"

      "github.com/tobocop2/beetrix/internal/config"
  )
  ```
- No dot imports (`import . "pkg"`)
- No blank imports (`import _ "pkg"`) unless for side effects (database drivers), with a comment explaining why
- **Meaningful package names** -- no `util`, `common`, `misc`, `helpers`. If you can't name it, the abstraction is wrong.
- **Internal packages** for encapsulation -- everything that shouldn't be part of the public API goes in `internal/`

### Configuration & State
- All config lives in the `Config` struct (`internal/config`)
- No mutable package-level globals (the bridge registry is an exception, loaded once at init)
- Prefer dependency injection (pass values as parameters) over reading globals inside functions
- Config flows down from `cmd/` through constructors

### TUI Rules (Bubble Tea)
- Elm architecture: Model, Init, Update, View
- Each screen is its own sub-model composed into the main Model
- Async operations use typed messages (`tea.Msg`), never goroutines that write to shared state
- Styles live in `styles.go`, not inline

### Tests Before Deletion
- When removing code, first make existing tests pass with the new implementation, then delete redundant tests
- Never delete tests and implementation in the same step

### YAGNI & Simplicity
- Don't add features, abstractions, or config that isn't needed yet
- Three similar lines are better than a premature abstraction
- Only validate at system boundaries (user input, external APIs) -- trust internal code
- No backwards-compatibility shims -- if something is unused, delete it

## Git & Workflow
- Every change tracked as a beads task (`bd create` -> `bd close`)
- Run `make check` before closing any task -- it mirrors CI exactly
- Tests and lint must pass before closing a task
- **Never git push without explicit user approval** -- ask before pushing
- No Co-Authored-By lines in commits
- No "Phase N" numbering in commit messages
- **Pre-commit checklist**: `make lint && make test` before every commit
- **Pre-push checklist**: run full review workflow, verify binary builds, run affected tests. Never push optimistically.
- PR descriptions: short, human-readable, no implementation details or internal names
- Don't rename branches with open PRs; use `gh pr edit` instead
- Use worktrees for parallel work, separate branches for distinct features
- Never stop to ask "should I keep going?" during multi-phase work. Just finish.
- Fix root causes individually, not downstream bandaids

## Code Review Standards

### Quality Gates
- **Low complexity** -- max ~3 branches per function, extract helpers when exceeded
- **DRY** -- reusable shared logic, no copy-paste
- **No private API leaks** -- unexported functions/fields stay internal to their package
- **Go idioms** -- error wrapping, table-driven tests, interface-based design, factory constructors
- **Named types over anonymous structs** -- any repeated struct shape should be a named type
- **Minimal changes** -- make smallest possible edit, don't rewrite large blocks for small fixes
- **Exhaustive review** -- multiple review passes until no new findings emerge
- **Compile before test** -- `go build ./...` before `go test ./...`

### Issue Categories
- **Critical** (must fix before merge): compilation errors, data loss risk, security holes, broken public API, missing error handling on I/O
- **Important** (should fix): missing tests for new paths, poor error messages, naming that misleads, unnecessary coupling
- **Minor** (nice to have): style nits, comment improvements, slight refactors

Every issue must include `file:line` reference. Clear verdict: Ready to merge? Yes / No / With fixes.

### Plan Alignment
When implementing from a plan:
1. Compare implementation against plan -- all items addressed?
2. No scope creep -- nothing added that wasn't in the plan
3. Architecture matches plan's design, not a "close enough" approximation

## Self-Review Checklist (before every push)

Run this before claiming work is done. Combines blast-radius analysis with Go-specific checks.

### 1. Map Change Surface
- [ ] List every file changed. Categorize: **core** (config, types), **feature** (single package), **leaf** (tests, docs).
- [ ] Core changes get full graph tracing. Leaf changes get spot checks.

### 2. Build Consumer Graph
- [ ] For every changed exported symbol, grep for who imports that package. Go 2-3 levels deep for shared packages (`internal/config`, `internal/bridges`).
- [ ] For changed interfaces, find all implementations and all callers.

### 3. Trace Changes Through Graph
- [ ] **Types/structs** -- field added/removed? All constructors, YAML tags, serializers updated?
- [ ] **Constants/enums** -- new constant used everywhere the raw value was? Grep globally for the raw value.
- [ ] **Function signatures** -- changed params or returns? Every call site updated?
- [ ] **Exports** -- new export actually needed? Removed export has no dangling references?

### 4. Enforce Project Rules
- [ ] **Constants over literals** -- no magic strings or numbers. Grep for raw values.
- [ ] **Error handling** -- every `error` return checked. Wrapped with `%w` and context.
- [ ] **No orphaned code** -- removed/renamed functions have no remaining references. No dead imports.
- [ ] **Import hygiene** -- stdlib / third-party / local grouping. No unused imports.
- [ ] **Godoc** -- every new exported symbol has a doc comment starting with its name.
- [ ] **No init()** -- unless extending the bridge registry (documented exception).

### 5. Assess Test Impact
- [ ] Every new code path has a test with meaningful assertions.
- [ ] Changed function signatures -- all test call sites updated?
- [ ] Table-driven tests for functions with multiple cases.
- [ ] Race detector passes (`go test -race ./...`).
- [ ] Tests use `t.TempDir()` and `t.Setenv()`, not real paths or global state.

### 6. Risk Score (mental model)
Before pushing, rate on each axis (low/medium/high):
- **Blast radius** -- how many packages import the changed code?
- **API breakage** -- any changed exported types, functions, or interfaces?
- **Test coverage** -- are all new paths covered?
- **Data risk** -- could this corrupt settings.yaml or user data?
- **Reversibility** -- can this be reverted cleanly?

If any axis is "high", get a second review before merging.

## Receiving Code Review

When receiving feedback on a PR:
- **Verify before implementing** -- reproduce the concern. Don't blindly apply suggestions.
- **Push back with technical reasoning** when feedback is wrong. No performative agreement.
- **Clarify unclear items** before implementing anything.
- **YAGNI check** -- if a reviewer suggests adding features that aren't needed yet, push back.
- **No drive-by refactors** -- review feedback should address the PR's scope, not unrelated code.

## QA Protocol
- QA must run in a dedicated worktree -- branch drift corrupts runs
- QA sessions: file bugs as beads issues, don't fix inline
- Use tmux send-keys for automated CLI and TUI testing
- Always ask before creating issues, PRs, or anything externally visible

## Behavior Learning (floop)
- `floop` captures corrections and learned behaviors across sessions
- Hooks run automatically via `~/.claude/settings.json`
- `floop active` -- show behaviors active in current context
- `floop learn` -- manually capture a correction/behavior
- `floop list` -- list all learned behaviors

## Key Files
- `cmd/root.go` -- CLI entry point, version
- `internal/config/config.go` -- Settings struct, defaults, token generation
- `internal/config/paths.go` -- XDG directory resolution
- `internal/bridges/bridges.yaml` -- Bridge registry (embedded)
- `internal/bridges/registry.go` -- Registry loader and lookup functions
- `internal/homeserver/homeserver.go` -- Dendrite embedding
- `internal/megabridge/manager.go` -- Bridge lifecycle manager
- `internal/megabridge/connectors/` -- Per-network connector wrappers
- `internal/server/server.go` -- Orchestrator (homeserver + megabridge)
- `internal/matrix/client.go` -- Matrix client for admin/bot operations
- `internal/tui/app.go` -- Bubble Tea main model
- `Makefile` -- Build, test, lint, check, qa targets

<!-- BEGIN BEADS INTEGRATION -->
## Issue Tracking with bd (beads)

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Auto-syncs to JSONL for version control
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

### Quick Start

**Check for ready work:**

```bash
bd ready --json
```

**Create new issues:**

```bash
bd create "Issue title" --description="Detailed context" -t bug|feature|task -p 0-4 --json
bd create "Issue title" --description="What this issue is about" -p 1 --deps discovered-from:bd-123 --json
```

**Claim and update:**

```bash
bd update bd-42 --status in_progress --json
bd update bd-42 --priority 1 --json
```

**Complete work:**

```bash
bd close bd-42 --reason "Completed" --json
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" --description="Details about what was found" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`

### Auto-Sync

bd automatically syncs with git:

- Exports to `.beads/issues.jsonl` after changes (5s debounce)
- Imports from JSONL when newer (e.g., after `git pull`)
- No manual export/import needed!

### Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic use
- ✅ Link discovered work with `discovered-from` dependencies
- ✅ Check `bd ready` before asking "what should I work on?"
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems

For more details, see README.md and docs/QUICKSTART.md.

<!-- END BEADS INTEGRATION -->

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
