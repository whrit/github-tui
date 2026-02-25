# Design: Bubble Tea UI Redesign

**Date:** 2026-02-25
**Status:** Approved

## Summary

Migrate the `ght` TUI from `rivo/tview` to the charmbracelet Bubble Tea ecosystem (`bubbletea` + `lipgloss` + `bubbles`). The GitHub API layer (`github/`, `domain/`, `config/`, `utils/`) is kept intact with one refinement: domain types shed the `tcell` color dependency in favor of semantic `ColorRole` strings resolved by the new theme layer.

New features introduced alongside the migration: visible tab bar, loading spinners, and markdown rendering for issue/comment bodies.

---

## Decisions

| Question | Decision |
|---|---|
| Framework | Migrate to charmbracelet/bubbletea + lipgloss + bubbles |
| Visual style | Dark & minimal (GitHub-dark palette) |
| API layer | Keep unchanged; refine domain types to remove tcell |
| Color model | Semantic `ColorRole` in domain; theme resolves to lipgloss colors |
| New features | Visible tab bar, loading spinners, markdown rendering |
| Migration strategy | Strangler fig — build full new UI alongside old, swap in `cmd/ght/main.go`, delete old |

---

## Package Structure

```
ui/
  app.go              # bubbletea root model; page routing; global keybinds
  theme/
    theme.go          # lipgloss styles, color palette, ColorRole → Color mapping
  components/
    tabbar.go         # "● Issues  ○ Actions" persistent header
    statusbar.go      # bottom bar: contextual keyhints + context info
    table.go          # reusable list widget wrapping bubbles/table
    viewport.go       # reusable preview pane wrapping bubbles/viewport
    spinner.go        # loading state wrapper around bubbles/spinner
    search.go         # inline search input wrapping bubbles/textinput
  pages/
    issues.go         # main issues page model (filter, list, sidebar, preview, comments)
    actions.go        # actions page model (runs list, jobs drill-down, log view)
    form.go           # issue create/edit form
```

The old `ui/` package is fully replaced. No adapter shim or compatibility layer.

---

## Domain Type Changes

`domain/item.go` — replace `tcell.Color` with a semantic `ColorRole`:

```go
type ColorRole string

const (
    ColorRoleDefault ColorRole = "default"
    ColorRoleMuted   ColorRole = "muted"
    ColorRoleAccent  ColorRole = "accent"
    ColorRoleSuccess ColorRole = "success"
    ColorRoleDanger  ColorRole = "danger"
    ColorRoleWarning ColorRole = "warning"
)

type Field struct {
    Text      string
    ColorRole ColorRole
}
```

All domain files (`issue.go`, `comment.go`, `workflow_run.go`, `workflow_job.go`, etc.) updated to use `ColorRole` constants instead of `tcell.Color*` values. The `tcell` import is removed from the `domain/` package entirely.

---

## Color Palette

Based on GitHub's dark default theme:

| Role | Hex | Usage |
|---|---|---|
| Background | `#0d1117` | terminal background |
| Surface | `#161b22` | panel backgrounds |
| Border | `#30363d` | unfocused panel borders |
| BorderFocus | `#58a6ff` | active panel border |
| Text | `#e6edf3` | primary text |
| TextMuted | `#7d8590` | column headers, secondary info |
| Success | `#3fb950` | open issues, successful runs |
| Danger | `#f85149` | closed issues, failed runs |
| Warning | `#d29922` | in-progress runs, queued |
| Accent | `#58a6ff` | issue numbers, branches, selected rows |

---

## Layout

### Main / Issues Page

```
  github-tui                              owner/repo
  ────────────────────────────────────────────────────
  ● Issues    ○ Actions
  ────────────────────────────────────────────────────
  ╭─ Filters ─────────────────────────────────────────╮
  │ repo:owner/repo state:open                        │
  ╰────────────────────────────────────────────────────╯
  ╭──── Issues ──────────────────╮╭──── Preview ───────╮
  │  #    Title         St  Auth ││                   │
  │ ────  ────────────  ──  ──── ││ ## Fix login bug  │
  │▶ 42   Fix login bug ●   jon  ││                   │
  │  41   Update README     alice││ Body text here... │
  │                              ││                   │
  ╰──────────────────────────────╯╰───────────────────╯
  ╭──── Comments ───────────────╮╭──── Comment Preview╮
  │▶ jon    2h ago              ││ > quoted text...   │
  │  alice  1d ago              ││                    │
  ╰──────────────────────────────╯╰───────────────────╯
  42 open · f:fetch  n:new  e:edit  c:close  /:search  ?:help
```

Sidebar panels (Assignees, Labels, Milestones, Projects) appear inline within the issues panel or as a collapsible sidebar — TBD during implementation based on available space.

### Actions Page

```
  github-tui                              owner/repo
  ────────────────────────────────────────────────────
  ○ Issues    ● Actions
  ────────────────────────────────────────────────────
  ╭──── Workflow Runs ─────────────────────────────────╮
  │  Status    Workflow      Branch   Event   Duration │
  │ ─────────  ────────────  ───────  ──────  ──────── │
  │▶ ✓ success  CI            main     push    1m23s   │
  │  ✕ failure  Deploy        feat/x   push    0m45s   │
  │  ⠋ running  Tests         pr/42    pr      ---     │
  ╰────────────────────────────────────────────────────╯
  Status: all · Workflow: all · s:status  w:workflow  r:refresh
```

Jobs drill-down replaces the runs list in-place (same panel) with an Escape to go back, consistent with current behavior.

### Loading State

```
  ╭──── Issues ──────────────────╮
  │                              │
  │    ⠸ Fetching issues...      │
  │                              │
  ╰──────────────────────────────╯
```

---

## Key Technical Decisions

### 1. Bubble Tea Root Model

`app.go` holds a `Model` with:
- `currentPage` enum (issues, actions)
- sub-models for each page
- theme reference

Global key handlers (tab switching, quit) are processed in the root `Update`; all other events are delegated to the active page model.

### 2. Goroutine → Message Pattern

API calls run in goroutines via `tea.Cmd`. The current `UI.updater` channel pattern is replaced with the standard Bubble Tea pattern:

```go
func fetchIssues(query string) tea.Cmd {
    return func() tea.Msg {
        items, pageInfo := github.GetIssues(...)
        return IssuesLoadedMsg{Items: items, PageInfo: pageInfo}
    }
}
```

### 3. Bubbles Component Usage

| Component | From bubbles | Usage |
|---|---|---|
| `table.Model` | `bubbles/table` | Issues list, workflow runs, jobs, comments |
| `viewport.Model` | `bubbles/viewport` | Issue preview, comment preview, log view |
| `textinput.Model` | `bubbles/textinput` | Filter bar, search bar |
| `spinner.Model` | `bubbles/spinner` | Loading state per-panel |

### 4. Markdown Rendering

`charmbracelet/glamour` renders issue/comment bodies. The `glamour.Render(text, "dark")` call (already present but commented out in the old `view.go`) is wired into the viewport update path. Output is passed directly as the viewport content.

### 5. Theme Singleton

`theme.Default()` returns a `Theme` struct initialized once at startup. Components reference it for consistent styling. No global mutable state — the theme is immutable after initialization.

---

## Dependencies to Add

```
charmbracelet/bubbletea     v1.x   # app loop
charmbracelet/lipgloss      v1.x   # styling
charmbracelet/bubbles        v1.x  # table, viewport, textinput, spinner
charmbracelet/glamour        v1.x  # markdown rendering
```

`rivo/tview` and `gdamore/tcell` are removed from `go.mod` once the old `ui/` package is deleted.

---

## Out of Scope

- Help overlay (`?` key) — deprioritized per user input; can be added as a follow-on
- Mouse support
- PRs (already listed in README as "Still Under Development")
- Config for custom keybindings

---

## Migration Steps (high-level)

1. Update `domain/` — replace `tcell.Color` with `ColorRole`, remove tcell import
2. Add new dependencies to `go.mod`
3. Build `ui/theme/` — color palette + style definitions
4. Build `ui/components/` — tabbar, statusbar, table, viewport, spinner, search
5. Build `ui/pages/issues.go` — full issues page model
6. Build `ui/pages/actions.go` — full actions page model
7. Build `ui/pages/form.go` — issue create/edit
8. Build `ui/app.go` — root model wiring all pages
9. Update `cmd/ght/main.go` — swap `ui.New().Start()` for new entry point
10. Delete old tview-based code; remove tview/tcell from `go.mod`
