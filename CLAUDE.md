# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`ght` is a terminal UI client for GitHub built in Go using the `tview` TUI framework. It provides issue management, comments, and GitHub Actions workflow viewing through a keyboard-driven interface. The module path is `github.com/skanehira/ght`.

## Build and Run Commands

```bash
# Build the binary
go build ./cmd/ght

# Install globally
go install ./cmd/ght

# Run (auto-detects owner/repo from git remote)
./ght
# Or specify a repo
./ght owner/repo

# Run tests
go test ./...

# Run tests for a specific package
go test ./github/

# Run a single test
go test ./github/ -run TestRateLimiter
```

## Architecture

### Package Structure

- **`cmd/ght/`** — Entry point. Parses args, inits config, validates token, starts UI.
- **`config/`** — Reads `config.yaml` (GitHub PAT) from OS-specific config dir (`$XDG_CONFIG_HOME/ght/` or equivalent).
- **`github/`** — All GitHub API interaction. Uses two clients:
  - **GraphQL** (`shurcooL/githubv4`) — Issues, comments, labels, milestones, projects, mutations.
  - **REST** (`google/go-github/v68`) — GitHub Actions (workflow runs, jobs, logs).
  - Both share a single `RateLimiter` that wraps the HTTP transport with token bucket rate limiting and a concurrency semaphore.
- **`domain/`** — Data types shared between `github/` and `ui/`. All displayable types implement the `Item` interface (`Key() string`, `Fields() []Field`).
- **`ui/`** — TUI built on `rivo/tview`. All UI state is global (package-level vars).
- **`utils/`** — Small helpers for opening URLs and string manipulation.

### UI Architecture

The app uses `tview.Pages` for navigation between two main pages:
- **"main"** — Issue-centric grid: filter bar, issues table, sidebar lists (assignees/labels/milestones/projects), comments, and preview panes.
- **"actions"** — GitHub Actions page with workflow runs list, jobs drill-down, and log viewing.

Key UI abstractions:
- **`SelectUI`** (`ui/select.go`) — Reusable table-based list widget with selection, pagination (`GetList`/`FetchList`), search filtering, and keybinding injection via `CaptureFunc`. Used for issues, comments, assignees, labels, milestones, projects, workflow runs, and workflow jobs.
- **`ViewUI`** (`ui/view.go`) — Read-only text view with search highlighting and full-screen toggle.
- **`FilterUI`** (`ui/filter.go`) — Input field for issue search queries.

UI updates from goroutines must go through `UI.updater` channel → `app.QueueUpdateDraw()` to be thread-safe with tview.

### Adding a New Domain Type

1. Create struct in `domain/` implementing `Item` interface (Key + Fields).
2. Add GitHub API call in `github/` (GraphQL query struct or REST client call).
3. Create UI in `ui/` using `NewSelectListUI()` with appropriate `getList` and `capture` functions.

### Page Navigation

- `Ctrl+A` switches to Actions page, `Ctrl+I` switches back to Issues (main).
- `Ctrl+N`/`Ctrl+P` cycles focus between primitives on the main page only.

### GitHub API Patterns

- GraphQL queries use struct-tag-based query building (`graphql:"..."` tags on struct fields).
- REST pagination uses page numbers; GraphQL uses cursor-based pagination. Both are adapted to `github.PageInfo`.
- All API clients share rate limiting via `RateLimiter.WrapTransport()`.

---

<!-- quikgraph-injected -->
# Quikgraph Integration

Quikgraph provides semantic code search and analysis. **Prefer CLI commands over MCP** — they're faster and provide richer output.

## Quick Reference

Use `qg` (short alias) instead of `quikgraph` for all commands.

### Smart Query (Auto Intent Detection)

| Need | Command |
|------|---------|
| Let Quikgraph decide query type | `qg query "your question" --agent` |

### Semantic Search (Use Instead of Grep/Glob)

| Need | Command |
|------|---------|
| Search code by meaning | `qg search "query" --code --agent` |
| Search docs/markdown | `qg search "query" --docs --agent` |
| Search by language | `qg search "query" --lang rust --agent` |

### Graph Analysis (Code Structure)

| Need | Command |
|------|---------|
| Who calls this function | `qg graph callers <symbol>` |
| What does this call | `qg graph callees <symbol>` |
| All references to symbol | `qg graph refs <symbol>` |
| Dependencies of symbol | `qg graph deps <symbol>` |
| What depends on this | `qg graph dependents <symbol>` |
| Path between symbols | `qg graph path <from> <to>` |

### Impact & Risk Analysis

| Need | Command |
|------|---------|
| Impact of uncommitted changes | `qg impact` |
| Full PR impact analysis | `qg impact-analyze --base main` |
| Which tests to run | `qg suggest-tests --base main` |

### Index Management

| Need | Command |
|------|---------|
| Check index status | `qg status` |
| Start/attach daemon | `qg watch .` |
| Health check | `qg doctor` |

## When to Use Quikgraph vs Built-in Tools

**Use `qg search --agent`** for: natural language queries, concept searches, finding code by behavior
**Use Grep** for: literal strings, exact regex patterns, known symbol names
**Use `qg graph`** for: callers, callees, dependencies, paths between symbols

## Important Notes

- **Always use `--agent` flag** for search/query — returns JSON optimized for LLMs
- **If Quikgraph not indexed**: Fall back to Grep/Glob
