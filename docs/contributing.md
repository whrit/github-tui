# Contributing and Release Guide

This document covers everything needed to develop, test, and release `ght` — a keyboard-driven terminal UI client for GitHub built in Go with the Bubble Tea TUI framework.

---

## Table of Contents

1. [Development Setup](#1-development-setup)
2. [Running Tests](#2-running-tests)
3. [Project Structure](#3-project-structure)
4. [Architecture Overview](#4-architecture-overview)
5. [Adding Features](#5-adding-features)
6. [Code Conventions](#6-code-conventions)
7. [Release Process](#7-release-process)
8. [GitHub Actions Workflows](#8-github-actions-workflows)
9. [Debugging](#9-debugging)

---

## 1. Development Setup

### Requirements

| Tool | Minimum Version | Notes |
|------|-----------------|-------|
| Go | 1.24 | See `go.mod` — `go 1.24.2` |
| git | Any recent | Required for remote auto-detection |
| GitHub PAT | — | Classic or fine-grained (see below) |

### Token Scopes

Create a classic Personal Access Token at `https://github.com/settings/tokens` with these scopes:

| Scope | Purpose |
|-------|---------|
| `repo` | Read/write issues, comments; also grants Actions access |
| `project` or `read:org` | GitHub Projects support |

Fine-grained PATs are also accepted. Scope validation is skipped for fine-grained tokens, so ensure the token has repository read/write, Actions read, and Projects read permissions manually.

### Clone and Build

```bash
git clone https://github.com/whrit/github-tui
cd github-tui
go build ./cmd/ght
```

### Run From Source

```bash
# Auto-detects owner/repo from the git remote of the current directory:
./ght

# Specify a repository explicitly:
./ght owner/repo
```

The binary parses `git remote get-url origin` and supports SSH, git protocol, and HTTPS remote URL formats.

### First-Run Configuration Wizard

On the first launch, if no config file exists, `ght` prompts interactively:

```
Welcome to ght! No config file found.
You need a GitHub Personal Access Token (classic) with repo, workflow, and read:org scopes.
Create one at: https://github.com/settings/tokens

Enter your GitHub token: <your token here>
```

The token is written to `config.yaml` in the OS config directory with permissions `0600`.

### Manual Configuration

If you prefer to write the config file before launching, create it at the path returned by `os.UserConfigDir()` + `/ght/config.yaml`:

| OS | Path |
|----|------|
| macOS | `~/Library/Application Support/ght/config.yaml` |
| Linux (XDG) | `$XDG_CONFIG_HOME/ght/config.yaml` or `~/.config/ght/config.yaml` |
| Windows | `%AppData%\ght\config.yaml` |

File contents:

```yaml
github:
  token: ghp_your_token_here
```

---

## 2. Running Tests

### Run All Tests

```bash
go test ./...
```

### Run Tests for a Specific Package

```bash
go test ./ui/pages/ -v
go test ./ui/components/ -v
go test ./ui/theme/ -v
go test ./github/ -v
go test ./domain/ -v
```

### Run a Single Named Test

```bash
go test ./ui/pages/ -run TestIssuesModel_LoadedMsg_SetsItems -v
go test ./github/ -run TestRateLimiter -v
go test ./github/ -run TestCleanLog -v
go test ./domain/ -run TestWorkflowRunStatusDisplay -v
```

### Test Coverage

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Philosophy

All page and component tests follow the Bubble Tea pattern: `Update()` is a pure value-receiver function that takes a `tea.Msg` and returns a new model and an optional `tea.Cmd`. There is no I/O, no goroutines, and no external dependencies inside `Update()`. This means tests can be written without mocks or test doubles — they simply send message values and assert on the returned model state.

Example from `/Users/beckett/Projects/github_clones/github-tui/ui/pages/issues_test.go`:

```go
func TestIssuesModel_LoadedMsg_SetsItems(t *testing.T) {
    th := theme.Default()
    m := pages.NewIssuesModel(th)

    msg := pages.IssuesLoadedMsg{
        Items: []domain.Item{
            &domain.Issue{ID: "1", Number: "1", Title: "Fix bug", State: "OPEN"},
        },
        PageInfo: &github.PageInfo{HasNextPage: githubv4.Boolean(false)},
    }

    updated, _ := m.Update(msg)
    m2 := updated.(pages.IssuesModel)

    if m2.Loading() {
        t.Error("loading should be false after IssuesLoadedMsg")
    }
    if len(m2.Issues()) != 1 {
        t.Errorf("expected 1 issue, got %d", len(m2.Issues()))
    }
}
```

Key patterns:
- Construct the model with `theme.Default()` — the theme is passed in, not a global.
- Send `tea.Msg` values directly to `m.Update(msg)`.
- Type-assert the returned `tea.Model` back to the concrete page type (`updated.(pages.IssuesModel)`).
- To test a command, call `cmd()` and assert on the returned message type.
- For `github/` package tests, use `package github` (internal test package) to access unexported helpers; for UI tests, use `package pages_test` or `package components_test` (external test package).

The rate limiter tests in `/Users/beckett/Projects/github_clones/github-tui/github/rate_limiter_test.go` use a `mockTransport` struct that implements `http.RoundTripper` to intercept HTTP calls without touching the network.

---

## 3. Project Structure

```
github-tui/
├── cmd/
│   └── ght/
│       └── main.go          # Entry point: args, config.Init(), token validation, ui.Start()
├── config/
│   └── config.go            # Reads config.yaml; first-run token wizard; sets config.GitHub.*
├── domain/
│   ├── item.go              # Item interface, Field struct, ColorRole constants
│   ├── issue.go             # Issue struct (implements Item)
│   ├── comment.go           # Comment struct
│   ├── label.go             # Label struct
│   ├── assignees.go         # Assignee struct
│   ├── milestone.go         # Milestone struct
│   ├── project.go           # Project struct
│   ├── workflow_run.go      # WorkflowRun struct (implements Item, Fields() returns color-coded status)
│   ├── workflow_job.go      # WorkflowJob struct (implements Item)
│   └── item_test.go         # Tests for ColorRole constants and status display logic
├── github/
│   ├── client.go            # NewClient(): initializes GraphQL + REST clients sharing a RateLimiter
│   ├── rate_limiter.go      # RateLimiter: token bucket + semaphore, wraps http.Transport
│   ├── token_validator.go   # ValidateTokenScopes(): checks PAT scopes via X-OAuth-Scopes header
│   ├── query.go             # GraphQL query structs (Issues, Comments, Labels, etc.)
│   ├── query_issue.go       # GetIssues() function
│   ├── query_comment.go     # GetComments() function
│   ├── query_label.go       # GetLabels() function
│   ├── query_assignees.go   # GetAssignees() function
│   ├── query_milestone.go   # GetMilestones() function
│   ├── query_project.go     # GetProjects() function
│   ├── query_repository.go  # GetRepo() function
│   ├── mutation_issue.go    # CreateIssue() mutation
│   ├── mutation_comment.go  # CreateComment() mutation
│   ├── actions.go           # ListWorkflowRuns(), ListWorkflowJobs(), GetWorkflowJobLog(), ConvertWorkflowRun(), ConvertWorkflowJob()
│   ├── actions_test.go      # Tests for log cleaning, duration formatting, run/job conversion
│   ├── rate_limiter_test.go # Tests for rate limiting, header parsing, concurrency semaphore
│   └── token_validator_test.go # Tests for PAT scope validation (uses httptest.Server)
├── ui/
│   ├── app.go               # AppModel: root tea.Model, routes msgs to IssuesModel/ActionsModel
│   ├── app_test.go          # Tests for page switching
│   ├── theme/
│   │   ├── theme.go         # Palette constants (GitHub dark), Theme struct, Default(), Resolve(), StyleField()
│   │   └── theme_test.go    # Tests for Resolve() and panel border style
│   ├── components/
│   │   ├── tabbar.go        # TabBar: renders "● Issues  ○ Actions" header
│   │   ├── tabbar_test.go   # Tests for tab rendering and active indicator placement
│   │   ├── statusbar.go     # StatusBar: renders key hints (left) + context (right)
│   │   └── statusbar_test.go
│   └── pages/
│       ├── issues.go        # IssuesModel: issues table, details, comments, preview panes
│       ├── issues_test.go   # Tests for loading, pagination, key events, View()
│       ├── actions.go       # ActionsModel: runs table, jobs drill-down, log viewport
│       ├── actions_test.go  # Tests for page switching and RunsLoadedMsg
│       ├── form.go          # CreateIssueFormModel: multi-field issue creation form
│       └── form_test.go     # Tests for Esc dismiss and Tab focus cycling
├── utils/                   # URL opening helpers, string utilities
├── docs/
│   ├── setup.md             # Initial setup documentation
│   ├── contributing.md      # This file
│   └── plans/               # Feature planning documents
├── .goreleaser.yaml         # Release pipeline: builds, archives, checksums, Homebrew formula
├── go.mod                   # Module: github.com/skanehira/ght, Go 1.24.2
└── CLAUDE.md                # AI coding assistant guidance
```

---

## 4. Architecture Overview

### The Bubble Tea Model

`ght` uses the Elm Architecture via the `charmbracelet/bubbletea` framework. The application is a hierarchy of `tea.Model` values:

```
AppModel  (ui/app.go)
├── IssuesModel  (ui/pages/issues.go)
│   ├── textinput.Model    (filter bar)
│   ├── table.Model        (issues table)
│   ├── table.Model        (comments table)
│   ├── viewport.Model     (issue body preview)
│   ├── viewport.Model     (comment preview)
│   ├── viewport.Model     (details: labels, assignees, milestone, projects)
│   └── spinner.Model
└── ActionsModel  (ui/pages/actions.go)
    ├── table.Model        (workflow runs)
    ├── table.Model        (workflow jobs)
    ├── viewport.Model     (job log)
    └── spinner.Model
```

The root `AppModel.Update()` in `/Users/beckett/Projects/github_clones/github-tui/ui/app.go` handles window resize events (forwarded to both pages simultaneously) and `SwitchPageMsg` (produced when `Ctrl+A` or `Ctrl+I` is pressed). All other messages are delegated to the currently active page model.

### Page Navigation

| Key | Action |
|-----|--------|
| `Ctrl+A` | Switch to Actions page |
| `Ctrl+I` | Switch back to Issues page |
| `Ctrl+C` | Quit |

Navigation between pages works via a `SwitchPageMsg` struct. Pressing `Ctrl+A` in `IssuesModel.Update()` returns a `tea.Cmd` that, when executed by the runtime, produces `SwitchPageMsg{Page: "actions"}`. The root `AppModel` catches that message and sets `m.currentPage`.

### Async Commands

Fetches from the GitHub API happen in `tea.Cmd` functions — closures that run in a goroutine managed by the Bubble Tea runtime. When they complete, they return a message value that the runtime delivers to `Update()` on the main goroutine. No explicit goroutine management or channel wiring is required in page code.

Pattern used throughout the codebase:

```go
// 1. Define the async command (in pages/issues.go):
func fetchIssues(query string, cursor *string) tea.Cmd {
    return func() tea.Msg {
        // ... call github.GetIssues(variables) ...
        return IssuesLoadedMsg{Items: items, PageInfo: pi}
    }
}

// 2. Return it from Update():
case "ctrl+r":
    return m, fetchIssues(m.query, nil)

// 3. Handle the result message in Update():
case IssuesLoadedMsg:
    m.loading = false
    m.issues = msg.Items
```

### Theme System

The theme is defined in `/Users/beckett/Projects/github_clones/github-tui/ui/theme/theme.go`. It uses the GitHub dark color palette:

```go
var Palette = struct {
    Background  string  // "#0d1117"
    Surface     string  // "#161b22"
    Border      string  // "#30363d"
    BorderFocus string  // "#58a6ff"
    Text        string  // "#e6edf3"
    TextMuted   string  // "#7d8590"
    Success     string  // "#3fb950"
    Danger      string  // "#f85149"
    Warning     string  // "#d29922"
    Accent      string  // "#58a6ff"
}{ ... }
```

The `Theme` struct holds pre-built `lipgloss.Style` values. `theme.Default()` constructs the canonical instance. Every page model and component receives a `Theme` at construction time — the theme is not a global.

`domain.ColorRole` is a string type used in `domain.Field` to express semantic color intent without importing lipgloss from the domain package. The theme's `Resolve(role ColorRole)` method maps a role to a `lipgloss.Color`.

### GitHub API Layer

Two clients are initialized in `github/client.go`:

- **GraphQL** via `github.com/shurcooL/githubv4` — used for issues, comments, labels, milestones, projects, and mutations.
- **REST** via `github.com/google/go-github/v68` — used for Actions (workflow runs, jobs, logs).

Both clients share a single `RateLimiter` instance. The rate limiter wraps each client's HTTP transport via `WrapTransport()`. It enforces:

1. A token bucket rate limiter (`golang.org/x/time/rate`) for both REST and GraphQL, defaulting to approximately 1.39 requests per second (5000/hour).
2. A concurrency semaphore capped at 90 concurrent requests (below GitHub's 100-request limit).
3. Parsing of `X-RateLimit-Remaining`, `X-RateLimit-Limit`, and `X-RateLimit-Reset` response headers from REST calls to track actual quota usage.

GraphQL requests (identified as `POST /graphql`) do not update REST rate limit stats and use a separate `graphQLLimiter`.

---

## 5. Adding Features

### Adding a New Keybinding to the Issues Page

1. Open `/Users/beckett/Projects/github_clones/github-tui/ui/pages/issues.go`.
2. Find the `tea.KeyMsg` case in `IssuesModel.Update()` — look for the `switch msg.String()` block.
3. Add your key:

```go
case "ctrl+r":
    m.loading = true
    return m, fetchIssues(m.query, nil)
```

4. Add a `KeyHint` entry in the `hints` slice inside `IssuesModel.View()`:

```go
hints := []components.KeyHint{
    {Key: "Ctrl+N", Desc: "next"},
    {Key: "Ctrl+P", Desc: "prev"},
    {Key: "n",      Desc: "new"},
    {Key: "f",      Desc: "fetch"},
    {Key: "/",      Desc: "search"},
    {Key: "Ctrl+R", Desc: "refresh"},  // your new hint
}
```

5. Add a test in `/Users/beckett/Projects/github_clones/github-tui/ui/pages/issues_test.go` following the existing pattern:

```go
func TestIssuesModel_CtrlR_RefreshesIssues(t *testing.T) {
    th := theme.Default()
    m := pages.NewIssuesModel(th)
    _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
    if cmd == nil {
        t.Error("Ctrl+R should return a fetch command")
    }
}
```

### Adding a New Domain Type

Follow these three steps:

**Step 1 — Define the domain struct** in `domain/`:

```go
// domain/release.go
package domain

type Release struct {
    ID      string
    TagName string
    Name    string
    Body    string
}

func (r *Release) Key() string { return r.ID }

func (r *Release) Fields() []Field {
    return []Field{
        {Text: r.TagName, ColorRole: ColorRoleAccent},
        {Text: r.Name,    ColorRole: ColorRoleDefault},
    }
}
```

Use `ColorRole` constants — never import lipgloss from the domain package.

**Step 2 — Add a GitHub API call** in `github/`:

For GraphQL, define a query struct with `graphql:"..."` struct tags (see `github/query_issue.go` as a reference). For REST, call the appropriate `go-github` client method (see `github/actions.go` for the REST pattern).

**Step 3 — Wire it into a page model** in `ui/pages/`:

Create a fetch command function returning a message struct, handle the message in `Update()`, and render the data in `View()`. Model state is stored as fields on the page struct — no globals.

### Adding a New Component

Components live in `ui/components/`. A component is a struct that receives a `theme.Theme` at construction and has a `View()` method returning a string. Components do not implement `tea.Model` — they are pure rendering helpers called from page `View()` methods.

```go
// ui/components/mycomponent.go
package components

import "github.com/skanehira/ght/ui/theme"

type MyComponent struct {
    th theme.Theme
}

func NewMyComponent(th theme.Theme) MyComponent {
    return MyComponent{th: th}
}

func (c MyComponent) View(label string) string {
    return c.th.Accent.Render(label)
}
```

Write tests in `ui/components/mycomponent_test.go` using `package components_test`.

---

## 6. Code Conventions

### Bubble Tea: Update Must Be Pure

`Update()` is a value receiver method. It must contain no I/O, no side effects, and no goroutine spawning. Side effects happen exclusively in `tea.Cmd` closures returned from `Update()`.

```go
// Correct: pure value receiver, I/O in a returned Cmd.
func (m IssuesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case tea.KeyMsg:
        return m, func() tea.Msg {
            result, err := github.GetIssues(...)
            return IssuesLoadedMsg{Items: result, Err: err}
        }
}

// Wrong: I/O directly in Update().
func (m IssuesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case tea.KeyMsg:
        result, _ := github.GetIssues(...) // Do not do this.
        m.issues = result
        return m, nil
}
```

### Styling: Use the Theme, Not Inline Styles

Always use pre-built styles from the injected `Theme` value. Do not call `lipgloss.NewStyle()` inside `View()` except where a panel style must be parameterized by width (e.g., `panelStyle()` in `pages/issues.go`).

```go
// Correct:
label := m.th.Accent.Render("Issues")

// Incorrect:
label := lipgloss.NewStyle().Foreground(lipgloss.Color("#58a6ff")).Render("Issues")
```

### Spinner Initialization

Initialize spinners with `th.Accent`, not a new lipgloss style:

```go
sp := spinner.New()
sp.Spinner = spinner.Dot
sp.Style = th.Accent   // Correct.
```

### Viewport Resize

When the terminal is resized, update viewport dimensions by assigning fields directly. Do not call `viewport.New(w, h)` — that creates a new viewport and loses the current scroll position.

```go
// Correct: preserves scroll position.
m.issuePreview.Width = newWidth
m.issuePreview.Height = newHeight

// Incorrect: discards scroll position.
m.issuePreview = viewport.New(newWidth, newHeight)
```

### Error Handling

Errors from async commands are returned in the message struct, not panicked or logged to stderr. Surface them in the status bar or in an `err` field on the model that `View()` renders:

```go
type IssuesLoadedMsg struct {
    Items    []domain.Item
    PageInfo *github.PageInfo
    Err      error   // Set when the fetch failed.
}

// In Update():
case IssuesLoadedMsg:
    m.loading = false
    if msg.Err != nil {
        m.err = msg.Err  // Render this in View().
        return m, nil
    }
    m.issues = msg.Items
```

### Dependency Pinning

`lipgloss` is pinned to `v1.1.0` in `go.mod`. Do not run `go get github.com/charmbracelet/lipgloss@latest` — it resolves to an unstable pre-release pseudo-version. If a lipgloss update is needed, pin to a specific stable release tag:

```bash
go get github.com/charmbracelet/lipgloss@v1.1.0
```

Similarly, keep the other Charm library versions (`bubbles`, `bubbletea`, `glamour`) aligned — the Charm ecosystem releases are tightly coupled.

---

## 7. Release Process

Releases are fully automated via GoReleaser. No manual binary uploads are needed.

### Creating a Release

```bash
git tag v1.2.3
git push origin v1.2.3
```

Pushing a tag matching `v*` triggers the release GitHub Actions workflow, which:

1. Runs `go mod tidy` (configured as a GoReleaser `before.hooks` step).
2. Compiles four binaries: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64.
3. Packages each into a `.tar.gz` archive containing the binary and `README.md`.
4. Generates a `checksums.txt` (SHA-256).
5. Creates a GitHub Release with all archives and checksums attached.
6. Pushes an updated Homebrew formula to `whrit/homebrew-tap`.

Build flags used: `CGO_ENABLED=0`, `-s -w` (strips debug info and DWARF for smaller binaries).

Archive naming convention: `ght_<os>_<arch>.tar.gz` (e.g., `ght_darwin_arm64.tar.gz`).

### Changelog Filtering

GoReleaser auto-generates the release changelog from commit messages between the previous and new tag. The following commit prefixes are excluded from the changelog:

| Prefix | Excluded |
|--------|----------|
| `docs:` | Yes |
| `test:` | Yes |
| `chore:` | Yes |
| `Merge pull request` | Yes |
| `Merge branch` | Yes |

To ensure a commit appears in the changelog, use `feat:`, `fix:`, `refactor:`, or similar conventional commit prefixes.

### Required GitHub Actions Secrets

| Secret | Purpose |
|--------|---------|
| `GITHUB_TOKEN` | Auto-provisioned by GitHub Actions; used to create the GitHub Release and upload assets |
| `HOMEBREW_TAP_GITHUB_TOKEN` | A PAT with `repo` write access to `whrit/homebrew-tap`; used by GoReleaser to push the formula |

### Setting Up the Homebrew Tap Repository

1. Create a public repository at `github.com/whrit/homebrew-tap` (can be empty initially).
2. Create a GitHub classic PAT with `repo` scope for your account (or a bot account).
3. Add it as the `HOMEBREW_TAP_GITHUB_TOKEN` secret in the `github-tui` repository at Settings > Secrets and variables > Actions.

After the first successful release, the tap repo will contain `Formula/ght.rb`. Users install with:

```bash
brew tap whrit/tap
brew install ght
```

### Local Dry Run (No Publishing)

Test the release pipeline locally without publishing to GitHub or Homebrew:

```bash
# Install goreleaser (macOS):
brew install goreleaser

# Build snapshot archives (no git tag required, no upload):
goreleaser release --snapshot --clean
```

Artifacts appear in `dist/`. Inspect them to confirm binary names, archive contents, and checksums before tagging a real release.

---

## 8. GitHub Actions Workflows

Create these two workflow files in `.github/workflows/`.

### Release Workflow

`.github/workflows/release.yml` — triggered when a `v*` tag is pushed:

```yaml
name: Release
on:
  push:
    tags:
      - 'v*'
jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

The `fetch-depth: 0` checkout is required — GoReleaser needs full git history to compute the changelog between tags.

`permissions: contents: write` is required for the default `GITHUB_TOKEN` to create a release.

### CI Workflow

`.github/workflows/ci.yml` — runs on every push to `main` and on pull requests:

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go test ./...
      - run: go build ./cmd/ght
```

The `go-version-file: go.mod` directive reads the Go version directly from `go.mod`, so the workflow stays in sync automatically when `go.mod` is updated.

---

## 9. Debugging

### Log File Location

The application writes a debug log on every run. Location by OS:

| OS | Path |
|----|------|
| macOS | `~/Library/Application Support/ght/debug.log` |
| Linux (XDG) | `$XDG_CONFIG_HOME/ght/debug.log` or `~/.config/ght/debug.log` |
| Windows | `%AppData%\ght\debug.log` |

The log file is created fresh on each launch (`os.Create`, not `os.OpenFile` with append). Previous runs' logs are overwritten.

Standard library `log` output goes to both the log file and stderr. The Bubble Tea alternate screen hides stderr from the terminal while the TUI is running, so the log file is the primary way to inspect output.

### Watching the Log in Real Time

```bash
# macOS:
tail -f ~/Library/Application\ Support/ght/debug.log

# Linux:
tail -f ~/.config/ght/debug.log
```

Run `ght` in one terminal and `tail -f` in another.

### Adding Temporary Log Statements

```go
import "log"

// Inside any function — output goes to the debug log.
log.Printf("fetchIssues: query=%q cursor=%v", query, cursor)
```

Remove log statements before opening a pull request unless they are intentional user-facing messages.

### Token Validation Errors

If `ght` exits immediately with a token error, the cause is printed to stderr before the TUI starts. Run in a terminal without redirecting stderr:

```bash
./ght 2>&1 | head -20
```

Common causes:

| Error | Fix |
|-------|-----|
| `Token validation failed: HTTP 401` | Token is invalid or expired |
| `Token scope check failed: missing scopes: [repo]` | Re-generate the token with the required scopes |
| `cannot deserialize config file` | The YAML in `config.yaml` is malformed |
| `invalid repo` | No git remote found and no `owner/repo` argument provided |

### Running Tests in Verbose Mode

```bash
go test ./ui/pages/ -v -run TestIssuesModel
```

The `-v` flag prints each test name and PASS/FAIL inline, which is helpful when a test panics inside `Update()` or `View()`.
