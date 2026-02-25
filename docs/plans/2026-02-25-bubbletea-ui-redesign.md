# Bubble Tea UI Redesign Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the rivo/tview UI with a charmbracelet Bubble Tea UI using a dark GitHub-palette design, adding a tab bar, loading spinners, and markdown rendering.

**Architecture:** Strangler-fig replacement — all files in `ui/` are replaced; `github/`, `domain/` (minus tcell), `config/`, and `utils/` are kept. Domain types drop the `tcell.Color` field for a semantic `ColorRole` string; the new theme layer maps roles to lipgloss colors. The Bubble Tea root model in `ui/app.go` owns page routing and delegates to `ui/pages/issues.go` and `ui/pages/actions.go`.

**Tech Stack:** `charmbracelet/bubbletea` (app loop), `charmbracelet/lipgloss` (styling), `charmbracelet/bubbles` (table, viewport, textinput, spinner), `charmbracelet/glamour` (markdown rendering).

**Design doc:** `docs/plans/2026-02-25-bubbletea-ui-redesign-design.md`

---

## Context: Layout and Color Palette

The issues page splits the terminal vertically into five horizontal bands:

```
┌ Header: "github-tui  owner/repo" ──────────────────────────────────┐
│ Tab bar: ● Issues   ○ Actions                                       │
├─────────────────────────────────────────────────────────────────────┤
│ Filter input (full width, 3 rows)                                   │
├─────────────┬────────────────────────┬──────────────────────────────┤
│ Details(20%)│ Issues table (45%)     │ Issue Preview (35%)          │
│ Labels      │  ▶ #42 Fix login ● jon │ ## Fix login bug             │
│ Assignees   │    #41 Update    alice  │ Body text...                 │
│ Milestone   │                        │                              │
├─────────────┴──────────────────┬─────┴──────────────────────────────┤
│ Comments table (55%)           │ Comment Preview (45%)              │
│  ▶ jon  2h   └ Fix login       │ > quoted text...                   │
│    alice  1d                   │                                    │
├────────────────────────────────┴────────────────────────────────────┤
│ Status bar (1 row): keyhints + context                              │
└─────────────────────────────────────────────────────────────────────┘
```

The actions page replaces everything between header and status bar with a single wide workflow-runs table (or jobs table when drilling in).

**Color constants (hex)** — define once in `ui/theme/theme.go`:

```
Background   #0d1117     Surface      #161b22
Border       #30363d     BorderFocus  #58a6ff
Text         #e6edf3     TextMuted    #7d8590
Success      #3fb950     Danger       #f85149
Warning      #d29922     Accent       #58a6ff
```

---

## Context: Bubble Tea Key Concepts

Bubbletea models satisfy:
```go
type Model interface {
    Init() tea.Cmd
    Update(tea.Msg) (tea.Model, tea.Cmd)
    View() string
}
```

- `Init()` returns a `tea.Cmd` to run on startup (e.g. `fetchIssues(...)`).
- `Update()` receives a `tea.Msg`, returns an updated model + optional command. **It must be a pure function** — all side effects happen in `tea.Cmd` goroutines.
- `View()` renders the current state to a string. Lipgloss styles the strings.

API calls return `tea.Cmd` functions:
```go
func fetchIssues(query string, cursor *string) tea.Cmd {
    return func() tea.Msg {
        items, pageInfo := github.GetIssues(...)
        return IssuesLoadedMsg{Items: items, PageInfo: pageInfo}
    }
}
```

Terminal size is available from `tea.WindowSizeMsg{Width, Height}` — store it in the root model and pass dimensions down to child models for layout.

---

## Task 1: Add Bubble Tea dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add dependencies**
```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/glamour@latest
```

**Step 2: Verify go.mod contains all four**
```bash
grep charmbracelet go.mod
```
Expected: four lines containing bubbletea, lipgloss, bubbles, glamour.

**Step 3: Verify it compiles (existing code unchanged)**
```bash
go build ./...
```
Expected: no errors (old ui/ still uses tview; we haven't touched it yet).

**Step 4: Commit**
```bash
git add go.mod go.sum
git commit -m "chore: add charmbracelet bubbletea/lipgloss/bubbles/glamour deps"
```

---

## Task 2: Update domain/ — replace tcell.Color with ColorRole

This removes the `github.com/gdamore/tcell/v2` import from `domain/`. After this task `go build ./...` will fail because the old `ui/` still references `f.Color` as `tcell.Color` — that is expected and resolved in Task 13.

**Files:**
- Modify: `domain/item.go`
- Modify: `domain/issue.go`
- Modify: `domain/comment.go`
- Modify: `domain/label.go`
- Modify: `domain/assignees.go`
- Modify: `domain/milestone.go`
- Modify: `domain/project.go`
- Modify: `domain/workflow_run.go`
- Modify: `domain/workflow_job.go`
- Create: `domain/item_test.go`

**Step 1: Write failing tests for domain/item.go**

```go
// domain/item_test.go
package domain_test

import (
    "testing"
    "github.com/skanehira/ght/domain"
)

func TestColorRoleConstants(t *testing.T) {
    roles := []domain.ColorRole{
        domain.ColorRoleDefault,
        domain.ColorRoleMuted,
        domain.ColorRoleAccent,
        domain.ColorRoleSuccess,
        domain.ColorRoleDanger,
        domain.ColorRoleWarning,
    }
    for _, r := range roles {
        if string(r) == "" {
            t.Errorf("ColorRole constant must not be empty string")
        }
    }
}

func TestFieldHasColorRole(t *testing.T) {
    f := domain.Field{Text: "hello", ColorRole: domain.ColorRoleSuccess}
    if f.ColorRole != domain.ColorRoleSuccess {
        t.Errorf("expected ColorRoleSuccess, got %s", f.ColorRole)
    }
}
```

**Step 2: Run to verify it fails**
```bash
go test ./domain/ -run TestColorRole -v
go test ./domain/ -run TestFieldHasColorRole -v
```
Expected: compile error — `domain.ColorRole` undefined.

**Step 3: Update domain/item.go**
```go
package domain

// ColorRole is a semantic color identifier resolved by the UI theme layer.
type ColorRole string

const (
    ColorRoleDefault ColorRole = "default"
    ColorRoleMuted   ColorRole = "muted"
    ColorRoleAccent  ColorRole = "accent"
    ColorRoleSuccess ColorRole = "success"
    ColorRoleDanger  ColorRole = "danger"
    ColorRoleWarning ColorRole = "warning"
)

type Item interface {
    Key() string
    Fields() []Field
}

type Field struct {
    Text      string
    ColorRole ColorRole
}
```

**Step 4: Run tests to verify they pass**
```bash
go test ./domain/ -run TestColorRole -v
go test ./domain/ -run TestFieldHasColorRole -v
```
Expected: PASS.

**Step 5: Update domain/issue.go**

Replace the `Fields()` method — map old tcell colors to ColorRole semantics:
- `tcell.ColorLightSalmon` (repo) → `ColorRoleMuted`
- `tcell.ColorBlue` (number) → `ColorRoleAccent`
- `tcell.ColorGreen`/`tcell.ColorRed` (state) → `ColorRoleSuccess`/`ColorRoleDanger`
- `tcell.ColorYellow` (author) → `ColorRoleWarning`
- `tcell.ColorWhite` (title) → `ColorRoleDefault`

```go
package domain

type Issue struct {
    ID        string
    Repo      string
    RepoOwner string
    Number    string
    State     string
    Title     string
    Body      string
    Author    string
    URL       string
    Labels    []Item
    Assignees []Item
    Comments  []Item
    MileStone []Item
    Projects  []Item
}

func (i *Issue) Key() string { return i.ID }

func (i *Issue) Fields() []Field {
    stateRole := ColorRoleSuccess
    if i.State == "CLOSED" {
        stateRole = ColorRoleDanger
    }
    return []Field{
        {Text: i.Repo + "/" + i.RepoOwner, ColorRole: ColorRoleMuted},
        {Text: i.Number, ColorRole: ColorRoleAccent},
        {Text: i.State, ColorRole: stateRole},
        {Text: i.Author, ColorRole: ColorRoleWarning},
        {Text: i.Title, ColorRole: ColorRoleDefault},
    }
}
```

**Step 6: Update remaining domain files**

Apply the same pattern — remove `tcell` import, replace `tcell.Color*` with `ColorRole` constants:

`domain/comment.go`:
- Author → `ColorRoleWarning`
- UpdatedAt → `ColorRoleMuted`

`domain/label.go`:
- Name → `ColorRoleWarning`

`domain/assignees.go`:
- Login → `ColorRoleAccent`

`domain/milestone.go`:
- Title → `ColorRoleSuccess`

`domain/project.go`:
- Name → `ColorRoleMuted`

`domain/workflow_run.go` — update `statusDisplay()` to return `ColorRole` instead of `tcell.Color`:
```go
func statusDisplay(status, conclusion string) (string, ColorRole) {
    switch status {
    case "completed":
        switch conclusion {
        case "success":
            return conclusion, ColorRoleSuccess
        case "failure":
            return conclusion, ColorRoleDanger
        default:
            return conclusion, ColorRoleMuted
        }
    case "in_progress":
        return status, ColorRoleWarning
    default:
        return status, ColorRoleMuted
    }
}
```

`domain/workflow_job.go` — same `statusDisplay` call, now returns `ColorRole`.

**Step 7: Verify domain package compiles**
```bash
go build ./domain/
```
Expected: success (no tcell references remain).

**Step 8: Run all existing tests**
```bash
go test ./github/ -v
```
Expected: all pass (github/ doesn't import domain color fields).

**Step 9: Verify build fails only for old ui/**
```bash
go build ./... 2>&1 | grep -v "^ui/"
```
Expected: errors only from `ui/` package (which references the old `f.Color tcell.Color` field). That's correct — the old ui/ is being replaced.

**Step 10: Commit**
```bash
git add domain/
git commit -m "refactor(domain): replace tcell.Color with semantic ColorRole"
```

---

## Task 3: Build ui/theme/theme.go

**Files:**
- Create: `ui/theme/theme.go`
- Create: `ui/theme/theme_test.go`

**Step 1: Write failing test**
```go
// ui/theme/theme_test.go
package theme_test

import (
    "testing"
    "github.com/charmbracelet/lipgloss"
    "github.com/skanehira/ght/domain"
    "github.com/skanehira/ght/ui/theme"
)

func TestThemeResolveColorRole(t *testing.T) {
    th := theme.Default()
    cases := []struct {
        role domain.ColorRole
    }{
        {domain.ColorRoleDefault},
        {domain.ColorRoleMuted},
        {domain.ColorRoleAccent},
        {domain.ColorRoleSuccess},
        {domain.ColorRoleDanger},
        {domain.ColorRoleWarning},
    }
    for _, c := range cases {
        color := th.Resolve(c.role)
        if color == (lipgloss.Color("")) {
            t.Errorf("Resolve(%s) returned empty color", c.role)
        }
    }
}

func TestThemeStylesNonZero(t *testing.T) {
    th := theme.Default()
    if th.Panel.GetBorderStyle() == (lipgloss.Border{}) {
        t.Error("Panel style must have a border")
    }
}
```

**Step 2: Run to verify failure**
```bash
go test ./ui/theme/ -v
```
Expected: compile error — package not found.

**Step 3: Implement ui/theme/theme.go**
```go
package theme

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/skanehira/ght/domain"
)

// Palette holds the raw hex color values.
var Palette = struct {
    Background  string
    Surface     string
    Border      string
    BorderFocus string
    Text        string
    TextMuted   string
    Success     string
    Danger      string
    Warning     string
    Accent      string
}{
    Background:  "#0d1117",
    Surface:     "#161b22",
    Border:      "#30363d",
    BorderFocus: "#58a6ff",
    Text:        "#e6edf3",
    TextMuted:   "#7d8590",
    Success:     "#3fb950",
    Danger:      "#f85149",
    Warning:     "#d29922",
    Accent:      "#58a6ff",
}

// Theme holds pre-built lipgloss styles.
type Theme struct {
    // Text styles
    Text    lipgloss.Style
    Muted   lipgloss.Style
    Accent  lipgloss.Style
    Success lipgloss.Style
    Danger  lipgloss.Style
    Warning lipgloss.Style

    // Panel styles
    Panel        lipgloss.Style
    PanelFocused lipgloss.Style

    // Table header row style
    TableHeader lipgloss.Style

    // Table selected row style
    TableSelected lipgloss.Style

    // Tab styles
    TabActive   lipgloss.Style
    TabInactive lipgloss.Style

    // Status bar
    StatusBar lipgloss.Style

    // Title bar
    TitleBar lipgloss.Style
}

// Default returns the canonical dark GitHub-palette theme.
func Default() Theme {
    p := Palette
    return Theme{
        Text:    lipgloss.NewStyle().Foreground(lipgloss.Color(p.Text)),
        Muted:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.TextMuted)),
        Accent:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.Accent)),
        Success: lipgloss.NewStyle().Foreground(lipgloss.Color(p.Success)),
        Danger:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.Danger)),
        Warning: lipgloss.NewStyle().Foreground(lipgloss.Color(p.Warning)),

        Panel: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color(p.Border)),

        PanelFocused: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color(p.BorderFocus)),

        TableHeader: lipgloss.NewStyle().
            Foreground(lipgloss.Color(p.TextMuted)).
            Bold(true),

        TableSelected: lipgloss.NewStyle().
            Foreground(lipgloss.Color(p.Text)).
            Background(lipgloss.Color("#1f2937")).
            Bold(false),

        TabActive: lipgloss.NewStyle().
            Foreground(lipgloss.Color(p.Accent)).
            Bold(true),

        TabInactive: lipgloss.NewStyle().
            Foreground(lipgloss.Color(p.TextMuted)),

        StatusBar: lipgloss.NewStyle().
            Foreground(lipgloss.Color(p.TextMuted)).
            Background(lipgloss.Color(p.Surface)).
            PaddingLeft(1).PaddingRight(1),

        TitleBar: lipgloss.NewStyle().
            Foreground(lipgloss.Color(p.Text)).
            Background(lipgloss.Color(p.Surface)).
            Bold(true).
            PaddingLeft(1).PaddingRight(1),
    }
}

// Resolve maps a domain ColorRole to a lipgloss.Color.
func (th Theme) Resolve(role domain.ColorRole) lipgloss.Color {
    switch role {
    case domain.ColorRoleSuccess:
        return lipgloss.Color(Palette.Success)
    case domain.ColorRoleDanger:
        return lipgloss.Color(Palette.Danger)
    case domain.ColorRoleWarning:
        return lipgloss.Color(Palette.Warning)
    case domain.ColorRoleAccent:
        return lipgloss.Color(Palette.Accent)
    case domain.ColorRoleMuted:
        return lipgloss.Color(Palette.TextMuted)
    default:
        return lipgloss.Color(Palette.Text)
    }
}

// StyleField returns the given text styled for its ColorRole.
func (th Theme) StyleField(text string, role domain.ColorRole) string {
    return lipgloss.NewStyle().Foreground(th.Resolve(role)).Render(text)
}
```

**Step 4: Run tests**
```bash
go test ./ui/theme/ -v
```
Expected: PASS.

**Step 5: Commit**
```bash
git add ui/theme/
git commit -m "feat(ui/theme): add lipgloss theme with GitHub dark palette"
```

---

## Task 4: Build ui/components/tabbar.go

**Files:**
- Create: `ui/components/tabbar.go`
- Create: `ui/components/tabbar_test.go`

**Step 1: Write failing test**
```go
// ui/components/tabbar_test.go
package components_test

import (
    "strings"
    "testing"
    "github.com/skanehira/ght/ui/components"
    "github.com/skanehira/ght/ui/theme"
)

func TestTabBarRendersActivePage(t *testing.T) {
    th := theme.Default()
    tb := components.NewTabBar(th)

    view := tb.View("issues")
    if !strings.Contains(view, "Issues") {
        t.Error("tab bar must contain 'Issues'")
    }
    if !strings.Contains(view, "Actions") {
        t.Error("tab bar must contain 'Actions'")
    }

    view2 := tb.View("actions")
    if !strings.Contains(view2, "Actions") {
        t.Error("tab bar must contain 'Actions' when actions is active")
    }
}
```

**Step 2: Run to verify failure**
```bash
go test ./ui/components/ -run TestTabBar -v
```
Expected: compile error.

**Step 3: Implement ui/components/tabbar.go**
```go
package components

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/skanehira/ght/ui/theme"
)

// TabBar renders the "● Issues  ○ Actions" header.
type TabBar struct {
    th theme.Theme
}

func NewTabBar(th theme.Theme) TabBar {
    return TabBar{th: th}
}

// View renders the tab bar. activePage is "issues" or "actions".
func (tb TabBar) View(activePage string) string {
    active := "● "
    inactive := "○ "

    issuesIndicator := inactive
    actionsIndicator := inactive
    if activePage == "issues" {
        issuesIndicator = active
    } else {
        actionsIndicator = active
    }

    issuesTab := tb.th.TabInactive.Render(issuesIndicator + "Issues")
    actionsTab := tb.th.TabInactive.Render(actionsIndicator + "Actions")
    if activePage == "issues" {
        issuesTab = tb.th.TabActive.Render(issuesIndicator + "Issues")
    } else {
        actionsTab = tb.th.TabActive.Render(actionsIndicator + "Actions")
    }

    return lipgloss.JoinHorizontal(lipgloss.Top, issuesTab, "    ", actionsTab)
}
```

**Step 4: Run tests**
```bash
go test ./ui/components/ -run TestTabBar -v
```
Expected: PASS.

**Step 5: Commit**
```bash
git add ui/components/tabbar.go ui/components/tabbar_test.go
git commit -m "feat(ui/components): add TabBar component"
```

---

## Task 5: Build ui/components/statusbar.go

**Files:**
- Create: `ui/components/statusbar.go`
- Create: `ui/components/statusbar_test.go`

**Step 1: Write failing test**
```go
// ui/components/statusbar_test.go
package components_test

import (
    "strings"
    "testing"
    "github.com/skanehira/ght/ui/components"
    "github.com/skanehira/ght/ui/theme"
)

func TestStatusBarRenderKeys(t *testing.T) {
    th := theme.Default()
    sb := components.NewStatusBar(th)
    view := sb.View(80, []components.KeyHint{
        {Key: "n", Desc: "new"},
        {Key: "/", Desc: "search"},
    }, "12 open")
    if !strings.Contains(view, "n") {
        t.Error("status bar must contain key hint 'n'")
    }
    if !strings.Contains(view, "12 open") {
        t.Error("status bar must contain context text")
    }
}
```

**Step 2: Run to verify failure**
```bash
go test ./ui/components/ -run TestStatusBar -v
```

**Step 3: Implement ui/components/statusbar.go**
```go
package components

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "github.com/skanehira/ght/ui/theme"
)

// KeyHint pairs a key label with a description for the status bar.
type KeyHint struct {
    Key  string
    Desc string
}

// StatusBar renders the bottom bar with keyhints and context info.
type StatusBar struct {
    th theme.Theme
}

func NewStatusBar(th theme.Theme) StatusBar {
    return StatusBar{th: th}
}

// View renders the status bar at the given terminal width.
// hints are rendered left-aligned; context is right-aligned.
func (sb StatusBar) View(width int, hints []KeyHint, context string) string {
    keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Accent))
    descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.TextMuted))

    parts := make([]string, 0, len(hints))
    for _, h := range hints {
        parts = append(parts, keyStyle.Render(h.Key)+":"+descStyle.Render(h.Desc))
    }
    left := strings.Join(parts, "  ")

    contextStyled := descStyle.Render(context)

    // Pad so context is right-aligned
    gap := width - lipgloss.Width(left) - lipgloss.Width(contextStyled)
    if gap < 1 {
        gap = 1
    }
    line := left + fmt.Sprintf("%*s", gap, "") + contextStyled

    return sb.th.StatusBar.Width(width).Render(line)
}
```

**Step 4: Run tests**
```bash
go test ./ui/components/ -run TestStatusBar -v
```
Expected: PASS.

**Step 5: Run all component tests**
```bash
go test ./ui/components/ -v
```

**Step 6: Commit**
```bash
git add ui/components/statusbar.go ui/components/statusbar_test.go
git commit -m "feat(ui/components): add StatusBar component"
```

---

## Task 6: Build ui/pages/issues.go — model definition and messages

**Files:**
- Create: `ui/pages/issues.go`

This task defines the model struct, the Msg types for async data, and the `Init()` method. No rendering yet.

**Step 1: Create ui/pages/issues.go with model and messages**

```go
package pages

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/shurcooL/githubv4"
    "github.com/skanehira/ght/config"
    "github.com/skanehira/ght/domain"
    "github.com/skanehira/ght/github"
    "github.com/skanehira/ght/ui/components"
    "github.com/skanehira/ght/ui/theme"
)

// focusedPanel tracks which panel has keyboard focus.
type focusedPanel int

const (
    focusFilter focusedPanel = iota
    focusIssues
    focusDetails
    focusComments
    focusIssuePreview
    focusCommentPreview
)

// Async message types

type IssuesLoadedMsg struct {
    Items    []domain.Item
    PageInfo *github.PageInfo
    Err      error
}

type IssuesMoreLoadedMsg struct {
    Items    []domain.Item
    PageInfo *github.PageInfo
    Err      error
}

type CommentsLoadedMsg struct {
    Issue *domain.Issue
}

// IssuesModel is the Bubble Tea model for the Issues page.
type IssuesModel struct {
    th         theme.Theme
    statusBar  components.StatusBar
    tabBar     components.TabBar

    // layout
    width  int
    height int

    // state
    focus      focusedPanel
    loading    bool
    query      string
    cursor     *string
    hasMore    bool
    issues     []domain.Item
    selectedIx int // index into issues

    // sub-models
    filter         textinput.Model
    issueTable     table.Model
    issuePreview   viewport.Model
    detailsView    viewport.Model
    commentsTable  table.Model
    commentPreview viewport.Model
    spinner        spinner.Model
}

// NewIssuesModel constructs the initial IssuesModel.
func NewIssuesModel(th theme.Theme) IssuesModel {
    // Filter input
    fi := textinput.New()
    fi.Placeholder = fmt.Sprintf("repo:%s/%s state:open", config.GitHub.Owner, config.GitHub.Repo)
    fi.SetValue(fmt.Sprintf("repo:%s/%s state:open", config.GitHub.Owner, config.GitHub.Repo))
    fi.Width = 80

    // Spinner
    sp := spinner.New()
    sp.Spinner = spinner.Dot
    sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Accent))

    m := IssuesModel{
        th:        th,
        statusBar: components.NewStatusBar(th),
        tabBar:    components.NewTabBar(th),
        loading:   true,
        focus:     focusIssues,
        filter:    fi,
        spinner:   sp,
    }

    m.issueTable = newTable(th)
    m.commentsTable = newTable(th)
    m.issuePreview = viewport.New(0, 0)
    m.commentPreview = viewport.New(0, 0)
    m.detailsView = viewport.New(0, 0)

    return m
}

func newTable(th theme.Theme) table.Model {
    s := table.DefaultStyles()
    s.Header = s.Header.
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color(theme.Palette.Border)).
        BorderBottom(true).
        Foreground(lipgloss.Color(theme.Palette.TextMuted)).
        Bold(true)
    s.Selected = s.Selected.
        Foreground(lipgloss.Color(theme.Palette.Text)).
        Background(lipgloss.Color("#1f2937"))
    s.Cell = s.Cell.Foreground(lipgloss.Color(theme.Palette.Text))

    t := table.New(table.WithFocused(false))
    t.SetStyles(s)
    return t
}

func (m IssuesModel) Init() tea.Cmd {
    return tea.Batch(
        m.spinner.Tick,
        fetchIssues(m.filter.Value(), nil),
    )
}
```

**Step 2: Add the async fetch commands**

Append to `ui/pages/issues.go`:

```go
// fetchIssues returns a Cmd that loads issues from GitHub.
func fetchIssues(query string, cursor *string) tea.Cmd {
    return func() tea.Msg {
        // Prepend is:issue if missing
        if !strings.Contains(query, "is:issue") {
            query = "is:issue " + query
        }

        v := map[string]interface{}{
            "query":  githubv4.String(query),
            "first":  githubv4.Int(30),
            "cursor": (*githubv4.String)(cursor),
        }

        resp, err := github.GetIssues(v)
        if err != nil {
            return IssuesLoadedMsg{Err: err}
        }

        items := make([]domain.Item, len(resp.Nodes))
        for i, node := range resp.Nodes {
            items[i] = node.Issue.ToDomain()
        }

        pi := github.PageInfo(resp.PageInfo)
        return IssuesLoadedMsg{Items: items, PageInfo: &pi}
    }
}

// fetchMoreIssues appends the next page.
func fetchMoreIssues(query string, cursor *string) tea.Cmd {
    return func() tea.Msg {
        v := map[string]interface{}{
            "query":  githubv4.String(query),
            "first":  githubv4.Int(30),
            "cursor": (*githubv4.String)(cursor),
        }
        resp, err := github.GetIssues(v)
        if err != nil {
            return IssuesMoreLoadedMsg{Err: err}
        }
        items := make([]domain.Item, len(resp.Nodes))
        for i, node := range resp.Nodes {
            items[i] = node.Issue.ToDomain()
        }
        pi := github.PageInfo(resp.PageInfo)
        return IssuesMoreLoadedMsg{Items: items, PageInfo: &pi}
    }
}
```

**Step 3: Verify it compiles**
```bash
go build ./ui/pages/
```
Expected: success (or errors only for unresolved references that will be added in later tasks).

**Step 4: Commit**
```bash
git add ui/pages/issues.go
git commit -m "feat(ui/pages/issues): add IssuesModel struct and async fetch commands"
```

---

## Task 7: Build ui/pages/issues.go — Update() handler

**Files:**
- Modify: `ui/pages/issues.go`

**Step 1: Write failing test for Update()**

Create `ui/pages/issues_test.go`:
```go
package pages_test

import (
    "testing"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/skanehira/ght/domain"
    "github.com/skanehira/ght/github"
    "github.com/skanehira/ght/ui/pages"
    "github.com/skanehira/ght/ui/theme"
)

func TestIssuesModel_LoadedMsg_SetsItems(t *testing.T) {
    th := theme.Default()
    m := pages.NewIssuesModel(th)

    msg := pages.IssuesLoadedMsg{
        Items: []domain.Item{
            &domain.Issue{ID: "1", Number: "1", Title: "Fix bug", State: "OPEN"},
        },
        PageInfo: &github.PageInfo{HasNextPage: false},
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

func TestIssuesModel_TabSwitchMsg(t *testing.T) {
    th := theme.Default()
    m := pages.NewIssuesModel(th)

    // Ctrl+A should emit a SwitchPageMsg for "actions"
    _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
    if cmd == nil {
        t.Error("Ctrl+A should return a command")
    }
}
```

**Step 2: Run to verify failure**
```bash
go test ./ui/pages/ -run TestIssuesModel -v
```
Expected: compile errors (Loading(), Issues() methods not defined, SwitchPageMsg not defined).

**Step 3: Add SwitchPageMsg and exported accessors to issues.go**

In `ui/pages/issues.go`, add:
```go
// SwitchPageMsg is sent when the user switches between main pages.
type SwitchPageMsg struct{ Page string }

// Loading returns true while the initial fetch is in progress.
func (m IssuesModel) Loading() bool { return m.loading }

// Issues returns the currently loaded issue items.
func (m IssuesModel) Issues() []domain.Item { return m.issues }
```

**Step 4: Implement Update() on IssuesModel**

Append to `ui/pages/issues.go`:
```go
func (m IssuesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.recalcLayout()

    case spinner.TickMsg:
        var cmd tea.Cmd
        m.spinner, cmd = m.spinner.Update(msg)
        cmds = append(cmds, cmd)

    case IssuesLoadedMsg:
        m.loading = false
        if msg.Err != nil {
            // TODO: surface error in status bar
            break
        }
        m.issues = msg.Items
        if msg.PageInfo != nil {
            m.hasMore = bool(msg.PageInfo.HasNextPage)
            c := string(msg.PageInfo.EndCursor)
            m.cursor = &c
        }
        m.rebuildIssueTable()
        if len(m.issues) > 0 {
            m.updateDetailsAndPreview(0)
        }

    case IssuesMoreLoadedMsg:
        m.loading = false
        if msg.Err != nil {
            break
        }
        m.issues = append(m.issues, msg.Items...)
        if msg.PageInfo != nil {
            m.hasMore = bool(msg.PageInfo.HasNextPage)
            c := string(msg.PageInfo.EndCursor)
            m.cursor = &c
        }
        m.rebuildIssueTable()

    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyCtrlA:
            return m, func() tea.Msg { return SwitchPageMsg{Page: "actions"} }
        case tea.KeyCtrlC:
            return m, tea.Quit
        case tea.KeyCtrlN:
            m.focusNext()
        case tea.KeyCtrlP:
            m.focusPrev()
        case tea.KeyEnter:
            if m.focus == focusFilter {
                m.loading = true
                m.query = m.filter.Value()
                m.cursor = nil
                cmds = append(cmds, m.spinner.Tick, fetchIssues(m.query, nil))
            }
        }

        // Delegate to focused sub-model
        switch m.focus {
        case focusFilter:
            var cmd tea.Cmd
            m.filter, cmd = m.filter.Update(msg)
            cmds = append(cmds, cmd)
        case focusIssues:
            prev := m.issueTable.Cursor()
            var cmd tea.Cmd
            m.issueTable, cmd = m.issueTable.Update(msg)
            cmds = append(cmds, cmd)
            if m.issueTable.Cursor() != prev {
                m.updateDetailsAndPreview(m.issueTable.Cursor())
            }
            // 'f' to fetch more
            if msg.Type == tea.KeyRunes && string(msg.Runes) == "f" && m.hasMore {
                m.loading = true
                cmds = append(cmds, m.spinner.Tick, fetchMoreIssues(m.query, m.cursor))
            }
        case focusComments:
            prev := m.commentsTable.Cursor()
            var cmd tea.Cmd
            m.commentsTable, cmd = m.commentsTable.Update(msg)
            cmds = append(cmds, cmd)
            if m.commentsTable.Cursor() != prev {
                m.updateCommentPreview(m.commentsTable.Cursor())
            }
        case focusIssuePreview:
            var cmd tea.Cmd
            m.issuePreview, cmd = m.issuePreview.Update(msg)
            cmds = append(cmds, cmd)
        case focusCommentPreview:
            var cmd tea.Cmd
            m.commentPreview, cmd = m.commentPreview.Update(msg)
            cmds = append(cmds, cmd)
        }
    }

    return m, tea.Batch(cmds...)
}

func (m *IssuesModel) focusNext() {
    order := []focusedPanel{focusFilter, focusIssues, focusDetails, focusComments,
        focusIssuePreview, focusCommentPreview}
    for i, p := range order {
        if p == m.focus {
            m.focus = order[(i+1)%len(order)]
            m.syncTableFocus()
            return
        }
    }
}

func (m *IssuesModel) focusPrev() {
    order := []focusedPanel{focusFilter, focusIssues, focusDetails, focusComments,
        focusIssuePreview, focusCommentPreview}
    for i, p := range order {
        if p == m.focus {
            m.focus = order[(i-1+len(order))%len(order)]
            m.syncTableFocus()
            return
        }
    }
}

func (m *IssuesModel) syncTableFocus() {
    m.issueTable.SetStyles(tableStylesForFocus(m.th, m.focus == focusIssues))
    m.commentsTable.SetStyles(tableStylesForFocus(m.th, m.focus == focusComments))
    m.filter.Blur()
    if m.focus == focusFilter {
        m.filter.Focus()
    }
}

func tableStylesForFocus(th theme.Theme, focused bool) table.Styles {
    s := table.DefaultStyles()
    s.Header = s.Header.
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color(theme.Palette.Border)).
        BorderBottom(true).
        Foreground(lipgloss.Color(theme.Palette.TextMuted)).
        Bold(true)
    if focused {
        s.Selected = s.Selected.
            Foreground(lipgloss.Color(theme.Palette.Text)).
            Background(lipgloss.Color("#1f2937"))
    } else {
        s.Selected = s.Selected.
            Foreground(lipgloss.Color(theme.Palette.TextMuted)).
            Background(lipgloss.Color(""))
    }
    return s
}

// rebuildIssueTable rebuilds the bubbles/table rows from m.issues.
func (m *IssuesModel) rebuildIssueTable() {
    cols := []table.Column{
        {Title: "#", Width: 5},
        {Title: "Title", Width: m.issueTableWidth() - 30},
        {Title: "State", Width: 8},
        {Title: "Author", Width: 14},
    }
    rows := make([]table.Row, len(m.issues))
    for i, item := range m.issues {
        issue := item.(*domain.Issue)
        rows[i] = table.Row{issue.Number, issue.Title, issue.State, issue.Author}
    }
    m.issueTable.SetColumns(cols)
    m.issueTable.SetRows(rows)
}

func (m *IssuesModel) issueTableWidth() int {
    // 45% of terminal width
    w := m.width * 45 / 100
    if w < 40 {
        w = 40
    }
    return w
}

// updateDetailsAndPreview updates the details panel and issue preview for row ix.
func (m *IssuesModel) updateDetailsAndPreview(ix int) {
    if ix >= len(m.issues) {
        return
    }
    issue := m.issues[ix].(*domain.Issue)

    // Details panel: labels, assignees, milestone, projects
    var sb strings.Builder
    writeDetail := func(label string, items []domain.Item) {
        sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.TextMuted)).Bold(true).Render(label) + "\n")
        if len(items) == 0 {
            sb.WriteString("  (none)\n")
        } else {
            for _, it := range items {
                sb.WriteString("  " + it.Fields()[0].Text + "\n")
            }
        }
        sb.WriteString("\n")
    }
    writeDetail("Labels", issue.Labels)
    writeDetail("Assignees", issue.Assignees)
    writeDetail("Milestone", issue.MileStone)
    writeDetail("Projects", issue.Projects)
    m.detailsView.SetContent(sb.String())

    // Issue preview: render markdown
    body := issue.Body
    if rendered, err := renderMarkdown(body, m.issuePreview.Width); err == nil {
        body = rendered
    }
    m.issuePreview.SetContent(body)
    m.issuePreview.GotoTop()

    // Comments table
    commentRows := make([]table.Row, len(issue.Comments))
    for i, c := range issue.Comments {
        comment := c.(*domain.Comment)
        commentRows[i] = table.Row{comment.Author, comment.UpdatedAt}
    }
    m.commentsTable.SetColumns([]table.Column{
        {Title: "Author", Width: 16},
        {Title: "Updated", Width: 14},
    })
    m.commentsTable.SetRows(commentRows)

    if len(issue.Comments) > 0 {
        m.updateCommentPreview(0)
    } else {
        m.commentPreview.SetContent("")
    }
}

func (m *IssuesModel) updateCommentPreview(ix int) {
    if ix >= len(m.issueTable.Rows()) {
        return
    }
    // Get selected issue's comments
    issueIx := m.issueTable.Cursor()
    if issueIx >= len(m.issues) {
        return
    }
    issue := m.issues[issueIx].(*domain.Issue)
    if ix >= len(issue.Comments) {
        return
    }
    comment := issue.Comments[ix].(*domain.Comment)
    body := comment.Body
    if rendered, err := renderMarkdown(body, m.commentPreview.Width); err == nil {
        body = rendered
    }
    m.commentPreview.SetContent(body)
    m.commentPreview.GotoTop()
}

func (m *IssuesModel) recalcLayout() {
    // Recalculate sub-model sizes based on terminal dimensions.
    // Heights (subtract header 3 + filter 3 + status 1 = 7):
    topH := (m.height - 7) * 55 / 100
    botH := (m.height - 7) - topH
    if topH < 5 {
        topH = 5
    }
    if botH < 3 {
        botH = 3
    }

    detailsW := m.width * 20 / 100
    issueW := m.width * 45 / 100
    previewW := m.width - detailsW - issueW

    m.issueTable.SetWidth(issueW)
    m.issueTable.SetHeight(topH)
    m.detailsView = viewport.New(detailsW-2, topH-2)
    m.issuePreview = viewport.New(previewW-2, topH-2)

    commentsW := m.width * 55 / 100
    commentPreviewW := m.width - commentsW

    m.commentsTable.SetWidth(commentsW)
    m.commentsTable.SetHeight(botH)
    m.commentPreview = viewport.New(commentPreviewW-2, botH-2)

    m.filter.Width = m.width - 4

    m.rebuildIssueTable()
}

func renderMarkdown(content string, width int) (string, error) {
    if width < 20 {
        width = 80
    }
    import_glamour "github.com/charmbracelet/glamour"
    r, err := import_glamour.NewTermRenderer(
        import_glamour.WithStandardStyle("dark"),
        import_glamour.WithWordWrap(width),
    )
    if err != nil {
        return "", err
    }
    return r.Render(content)
}
```

> **Note on renderMarkdown:** The `import_glamour` syntax above is pseudocode illustrating the import alias pattern. The actual file should have `import "github.com/charmbracelet/glamour"` at the top and call `glamour.NewTermRenderer(...)`.

**Step 5: Run tests**
```bash
go test ./ui/pages/ -run TestIssuesModel -v
```
Expected: PASS.

**Step 6: Build check**
```bash
go build ./ui/pages/
```

**Step 7: Commit**
```bash
git add ui/pages/issues.go ui/pages/issues_test.go
git commit -m "feat(ui/pages/issues): implement Update() and layout recalculation"
```

---

## Task 8: Build ui/pages/issues.go — View() rendering

**Files:**
- Modify: `ui/pages/issues.go`

**Step 1: Write failing test for View()**
```go
// In ui/pages/issues_test.go — add:
func TestIssuesModel_ViewContainsPanels(t *testing.T) {
    th := theme.Default()
    m := pages.NewIssuesModel(th)

    // Simulate a window size
    updated, _ := m.Update(tea.WindowSizeMsg{Width: 220, Height: 50})
    m2 := updated.(pages.IssuesModel)

    view := m2.View()
    if len(view) == 0 {
        t.Error("View() should not return empty string")
    }
}
```

**Step 2: Run to verify failure**
```bash
go test ./ui/pages/ -run TestIssuesModel_View -v
```

**Step 3: Implement View() on IssuesModel**

```go
func (m IssuesModel) View() string {
    if m.width == 0 {
        return "Loading..."
    }

    // Header + tab bar
    titleRight := lipgloss.NewStyle().
        Foreground(lipgloss.Color(theme.Palette.TextMuted)).
        Render(fmt.Sprintf("%s/%s", config.GitHub.Owner, config.GitHub.Repo))
    titleLeft := m.th.TitleBar.Render("github-tui")
    titleGap := m.width - lipgloss.Width(titleLeft) - lipgloss.Width(titleRight)
    if titleGap < 1 {
        titleGap = 1
    }
    titleRow := titleLeft + fmt.Sprintf("%*s", titleGap, "") + titleRight

    tabRow := "  " + m.tabBar.View("issues")
    divider := lipgloss.NewStyle().
        Foreground(lipgloss.Color(theme.Palette.Border)).
        Render(strings.Repeat("─", m.width))

    // Filter
    filterLabel := lipgloss.NewStyle().
        Foreground(lipgloss.Color(theme.Palette.TextMuted)).
        Render("Filters: ")
    filterBox := m.th.Panel.Width(m.width - 2).Render(filterLabel + m.filter.View())

    // Loading spinner or main content
    var mainContent string
    if m.loading && len(m.issues) == 0 {
        mainContent = m.renderLoading()
    } else {
        mainContent = m.renderPanels()
    }

    // Status bar
    hints := []components.KeyHint{
        {Key: "Ctrl+N", Desc: "next"},
        {Key: "Ctrl+P", Desc: "prev"},
        {Key: "n", Desc: "new"},
        {Key: "f", Desc: "fetch"},
        {Key: "/", Desc: "search"},
    }
    context := fmt.Sprintf("%d issues", len(m.issues))
    status := m.statusBar.View(m.width, hints, context)

    return lipgloss.JoinVertical(lipgloss.Left,
        titleRow,
        tabRow,
        divider,
        filterBox,
        mainContent,
        status,
    )
}

func (m IssuesModel) renderLoading() string {
    h := m.height - 10
    if h < 3 {
        h = 3
    }
    spinnerRow := fmt.Sprintf("    %s Fetching issues...", m.spinner.View())
    return lipgloss.NewStyle().Width(m.width).Height(h).
        Render(strings.Repeat("\n", h/2) + spinnerRow)
}

func (m IssuesModel) renderPanels() string {
    detailsW := m.width * 20 / 100
    issueW := m.width * 45 / 100
    previewW := m.width - detailsW - issueW

    // Top row: details | issues | preview
    detailsPanel := m.panelStyle(m.focus == focusDetails, detailsW).
        Render(m.detailsView.View())
    issuesPanel := m.panelStyle(m.focus == focusIssues, issueW).
        Render(m.issueTable.View())
    previewPanel := m.panelStyle(m.focus == focusIssuePreview, previewW).
        Render(m.issuePreview.View())
    topRow := lipgloss.JoinHorizontal(lipgloss.Top, detailsPanel, issuesPanel, previewPanel)

    // Bottom row: comments | comment preview
    commentsW := m.width * 55 / 100
    commentPreviewW := m.width - commentsW
    commentsPanel := m.panelStyle(m.focus == focusComments, commentsW).
        Render(m.commentsTable.View())
    commentPreviewPanel := m.panelStyle(m.focus == focusCommentPreview, commentPreviewW).
        Render(m.commentPreview.View())
    botRow := lipgloss.JoinHorizontal(lipgloss.Top, commentsPanel, commentPreviewPanel)

    return lipgloss.JoinVertical(lipgloss.Left, topRow, botRow)
}

func (m IssuesModel) panelStyle(focused bool, width int) lipgloss.Style {
    if focused {
        return m.th.PanelFocused.Width(width - 2)
    }
    return m.th.Panel.Width(width - 2)
}
```

**Step 4: Run tests**
```bash
go test ./ui/pages/ -v
```
Expected: all PASS.

**Step 5: Commit**
```bash
git add ui/pages/issues.go
git commit -m "feat(ui/pages/issues): implement View() rendering"
```

---

## Task 9: Build ui/pages/actions.go

**Files:**
- Create: `ui/pages/actions.go`

Mirrors the structure of `issues.go` but for workflow runs, jobs drill-down, and log viewing. Use the existing `actions.go` in the old `ui/` as a reference for keybindings and behavior (filter cycling, workflow selector, log fetching).

**Step 1: Create the model**

```go
package pages

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    gogithub "github.com/google/go-github/v68/github"
    "github.com/skanehira/ght/config"
    "github.com/skanehira/ght/domain"
    "github.com/skanehira/ght/github"
    "github.com/skanehira/ght/ui/components"
    "github.com/skanehira/ght/ui/theme"
    "github.com/skanehira/ght/utils"
)

type actionsView int

const (
    viewRuns actionsView = iota
    viewJobs
    viewLog
)

// Async message types for Actions page.
type RunsLoadedMsg struct {
    Items    []domain.Item
    PageInfo *github.PageInfo
    Err      error
}

type JobsLoadedMsg struct {
    Items []domain.Item
    Err   error
}

type LogLoadedMsg struct {
    Content   string
    Truncated bool
    Err       error
}

// statusFilterCycle is the ordered list of status filter values.
var statusFilterCycle = []string{"", "success", "failure", "in_progress", "queued"}

// ActionsModel is the Bubble Tea model for the Actions page.
type ActionsModel struct {
    th        theme.Theme
    statusBar components.StatusBar
    tabBar    components.TabBar

    width  int
    height int

    view         actionsView
    loading      bool
    statusFilter string
    statusIx     int // index into statusFilterCycle
    workflowID   int64
    workflowName string
    workflows    []*gogithub.Workflow
    cursor       *string
    hasMore      bool

    currentRunID   int64
    currentRunName string

    logCancel context.CancelFunc

    runsTable  table.Model
    jobsTable  table.Model
    logView    viewport.Model
    sp         spinner.Model
}

func NewActionsModel(th theme.Theme) ActionsModel {
    sp := spinner.New()
    sp.Spinner = spinner.Dot
    sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Accent))

    m := ActionsModel{
        th:        th,
        statusBar: components.NewStatusBar(th),
        tabBar:    components.NewTabBar(th),
        loading:   true,
        sp:        sp,
        runsTable: newTable(th),
        jobsTable: newTable(th),
    }
    m.logView = viewport.New(0, 0)
    return m
}

func (m ActionsModel) Init() tea.Cmd {
    return tea.Batch(m.sp.Tick, fetchRuns("", 0, nil))
}
```

**Step 2: Add fetch commands**

```go
func fetchRuns(statusFilter string, workflowID int64, cursor *string) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        opts := &gogithub.ListWorkflowRunsOptions{
            ListOptions: gogithub.ListOptions{PerPage: 30},
        }
        if statusFilter != "" {
            opts.Status = statusFilter
        }
        if cursor != nil {
            // cursor holds the next page number as a string
            // (parsed in ListWorkflowRuns via strconv.Atoi)
        }

        var runs *gogithub.WorkflowRuns
        var resp *gogithub.Response
        var err error

        owner, repo := config.GitHub.Owner, config.GitHub.Repo
        if workflowID > 0 {
            runs, resp, err = github.ListWorkflowRunsByWorkflowID(ctx, owner, repo, workflowID, opts)
        } else {
            runs, resp, err = github.ListWorkflowRuns(ctx, owner, repo, opts)
        }
        if err != nil {
            return RunsLoadedMsg{Err: err}
        }

        items := make([]domain.Item, len(runs.WorkflowRuns))
        for i, run := range runs.WorkflowRuns {
            items[i] = github.ConvertWorkflowRun(run)
        }
        pi := &github.PageInfo{}
        if resp != nil && resp.NextPage > 0 {
            pi.HasNextPage = true
            pi.EndCursor = githubv4.String(fmt.Sprintf("%d", resp.NextPage))
        }
        return RunsLoadedMsg{Items: items, PageInfo: pi}
    }
}

func fetchJobs(runID int64) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        jobs, err := github.ListWorkflowJobs(ctx, config.GitHub.Owner, config.GitHub.Repo, runID, nil)
        if err != nil {
            return JobsLoadedMsg{Err: err}
        }
        items := make([]domain.Item, len(jobs.Jobs))
        for i, job := range jobs.Jobs {
            items[i] = github.ConvertWorkflowJob(job)
        }
        return JobsLoadedMsg{Items: items}
    }
}

func fetchLog(jobID int64, cancel context.CancelFunc) tea.Cmd {
    return func() tea.Msg {
        ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
        // Store cancel for navigation interruption
        _ = cancel
        defer cancelFn()

        content, truncated, err := github.GetWorkflowJobLog(ctx, config.GitHub.Owner, config.GitHub.Repo, jobID)
        return LogLoadedMsg{Content: content, Truncated: truncated, Err: err}
    }
}
```

**Step 3: Implement Update() and View()**

The Update() mirrors the old `ui/actions.go` keybindings:
- `Enter` on a run → switch to jobs view, fetch jobs
- `Escape` in jobs view → back to runs view
- `Enter` on a job → fetch and show log
- `'s'` → cycle status filter
- `'r'` → refresh current view
- `'w'` → show workflow selector (modal list)
- `Ctrl+I` → emit `SwitchPageMsg{Page: "issues"}`

View() renders the tab bar + status line + table (runs or jobs) or log viewport.

Status bar context changes based on view:
- Runs: `"Status: all | Workflow: all"`
- Jobs: `"Run: #42 - CI | Esc: back"`
- Log: `"Log: job-name | 'o' close"`

Implement following the same pattern as `issues.go`. The full code is analogous — omitted here for brevity but must be complete in the actual file.

**Step 4: Write and run tests**

```go
// ui/pages/actions_test.go
package pages_test

import (
    "testing"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/skanehira/ght/ui/pages"
    "github.com/skanehira/ght/ui/theme"
)

func TestActionsModel_TabSwitchToIssues(t *testing.T) {
    th := theme.Default()
    m := pages.NewActionsModel(th)
    _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlI})
    if cmd == nil {
        t.Error("Ctrl+I should emit SwitchPageMsg")
    }
}

func TestActionsModel_RunsLoaded(t *testing.T) {
    th := theme.Default()
    m := pages.NewActionsModel(th)
    msg := pages.RunsLoadedMsg{Items: nil, PageInfo: nil}
    updated, _ := m.Update(msg)
    m2 := updated.(pages.ActionsModel)
    if m2.Loading() {
        t.Error("loading should be false after RunsLoadedMsg")
    }
}
```

```bash
go test ./ui/pages/ -run TestActionsModel -v
```
Expected: PASS.

**Step 5: Commit**
```bash
git add ui/pages/actions.go ui/pages/actions_test.go
git commit -m "feat(ui/pages/actions): implement Actions page model"
```

---

## Task 10: Build ui/pages/form.go — issue create form

**Files:**
- Create: `ui/pages/form.go`

This replaces `createIssueForm()` from the old `ui/issues.go`. Use the same GitHub API calls — fetch repo, assignees, labels, projects, milestones — and wire them into a form rendered with lipgloss.

**Step 1: Define form model**

```go
package pages

import (
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/skanehira/ght/ui/theme"
)

// FormField indices
const (
    fieldTitle = iota
    fieldAssignees
    fieldLabels
    fieldProjects
    fieldMilestone
    fieldCount
)

// CreateIssueFormModel is a simple multi-field form for creating issues.
type CreateIssueFormModel struct {
    th      theme.Theme
    inputs  []textinput.Model
    focused int
    repoID  interface{} // githubv4.ID
    owner   string
    repo    string
    err     error
    done    bool
}

// IssueCreatedMsg signals successful issue creation.
type IssueCreatedMsg struct{}

// FormDismissedMsg signals the form was cancelled.
type FormDismissedMsg struct{}

func NewCreateIssueForm(th theme.Theme, owner, repo string) CreateIssueFormModel {
    inputs := make([]textinput.Model, fieldCount)
    labels := []string{"Title", "Assignees", "Labels", "Projects", "Milestone"}
    for i, label := range labels {
        ti := textinput.New()
        ti.Placeholder = label
        inputs[i] = ti
    }
    inputs[fieldTitle].Focus()

    return CreateIssueFormModel{
        th:     th,
        inputs: inputs,
        owner:  owner,
        repo:   repo,
    }
}
```

Implement `Init()`, `Update()` (Tab/Shift+Tab navigation, Enter to submit, Esc to cancel), and `View()` (lipgloss-styled form with field labels and input boxes). Submit calls `github.CreateIssue(input)` in a `tea.Cmd` and returns `IssueCreatedMsg` or `ErrorMsg`.

**Step 2: Commit when complete**
```bash
git add ui/pages/form.go
git commit -m "feat(ui/pages/form): add CreateIssueForm model"
```

---

## Task 11: Build ui/app.go — root model

**Files:**
- Create: `ui/app.go`

**Step 1: Write failing test**
```go
// ui/app_test.go
package ui_test

import (
    "testing"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/skanehira/ght/ui"
)

func TestAppModel_DefaultPageIsIssues(t *testing.T) {
    m := ui.NewApp()
    if m.CurrentPage() != "issues" {
        t.Errorf("expected default page 'issues', got %s", m.CurrentPage())
    }
}

func TestAppModel_SwitchPageMsg(t *testing.T) {
    m := ui.NewApp()
    updated, _ := m.Update(pages.SwitchPageMsg{Page: "actions"})
    m2 := updated.(ui.AppModel)
    if m2.CurrentPage() != "actions" {
        t.Errorf("expected page 'actions', got %s", m2.CurrentPage())
    }
}
```

**Step 2: Run to verify failure**
```bash
go test ./ui/ -run TestAppModel -v
```

**Step 3: Implement ui/app.go**

```go
package ui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/skanehira/ght/ui/pages"
    "github.com/skanehira/ght/ui/theme"
)

// AppModel is the root Bubble Tea model.
type AppModel struct {
    th          theme.Theme
    currentPage string
    issues      pages.IssuesModel
    actions     pages.ActionsModel
    width       int
    height      int
}

func NewApp() AppModel {
    th := theme.Default()
    return AppModel{
        th:          th,
        currentPage: "issues",
        issues:      pages.NewIssuesModel(th),
        actions:     pages.NewActionsModel(th),
    }
}

func (m AppModel) CurrentPage() string { return m.currentPage }

func (m AppModel) Init() tea.Cmd {
    return m.issues.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        var cmd1, cmd2 tea.Cmd
        var updated tea.Model
        updated, cmd1 = m.issues.Update(msg)
        m.issues = updated.(pages.IssuesModel)
        updated, cmd2 = m.actions.Update(msg)
        m.actions = updated.(pages.ActionsModel)
        return m, tea.Batch(cmd1, cmd2)

    case pages.SwitchPageMsg:
        m.currentPage = msg.Page
        if msg.Page == "actions" {
            return m, m.actions.Init()
        }
        return m, nil

    case tea.KeyMsg:
        if msg.Type == tea.KeyCtrlC {
            return m, tea.Quit
        }
    }

    // Delegate to active page
    var cmd tea.Cmd
    var updated tea.Model
    switch m.currentPage {
    case "issues":
        updated, cmd = m.issues.Update(msg)
        m.issues = updated.(pages.IssuesModel)
    case "actions":
        updated, cmd = m.actions.Update(msg)
        m.actions = updated.(pages.ActionsModel)
    }
    return m, cmd
}

func (m AppModel) View() string {
    switch m.currentPage {
    case "actions":
        return m.actions.View()
    default:
        return m.issues.View()
    }
}

// Start runs the Bubble Tea program. Called from cmd/ght/main.go.
func Start() error {
    p := tea.NewProgram(
        NewApp(),
        tea.WithAltScreen(),
    )
    _, err := p.Run()
    return err
}
```

**Step 4: Run tests**
```bash
go test ./ui/ -run TestAppModel -v
```
Expected: PASS.

**Step 5: Commit**
```bash
git add ui/app.go ui/app_test.go
git commit -m "feat(ui/app): add root AppModel and Start() entry point"
```

---

## Task 12: Wire cmd/ght/main.go to new UI

**Files:**
- Modify: `cmd/ght/main.go`

**Step 1: Update the UI call**

Replace:
```go
if err := ui.New().Start(); err != nil {
    log.Fatal(err)
}
```

With:
```go
if err := ui.Start(); err != nil {
    log.Fatal(err)
}
```

Remove the old `ui.New()` call — `Start()` is now a package function.

**Step 2: Build**
```bash
go build ./cmd/ght
```
Expected: success (old ui/ files will now cause duplicate symbol errors — resolve by deleting them in Task 13).

**Step 3: Commit**
```bash
git add cmd/ght/main.go
git commit -m "feat(cmd): wire new bubbletea UI entry point"
```

---

## Task 13: Delete old ui/ files and clean go.mod

**Files:**
- Delete: all old tview-based files in `ui/`
- Modify: `go.mod` / `go.sum`

**Step 1: Delete old ui/ source files**

The files to delete (they all belong to the old tview UI):
```bash
rm ui/actions.go ui/assignees.go ui/comments.go ui/filter.go \
   ui/issues.go ui/labels.go ui/milestones.go ui/projects.go \
   ui/search.go ui/select.go ui/ui.go ui/view.go
```

**Step 2: Build**
```bash
go build ./...
```
Expected: success. The `ui/` package now only contains the new bubbletea files.

**Step 3: Remove unused dependencies**

Check what's now unused:
```bash
go mod tidy
```

Verify tview and tcell are removed:
```bash
grep -E "tview|tcell" go.mod
```
Expected: no output (both removed).

**Step 4: Run all tests**
```bash
go test ./...
```
Expected: all pass.

**Step 5: Build and smoke-test the binary**
```bash
go build ./cmd/ght && echo "Build OK"
```

**Step 6: Final commit**
```bash
git add -A
git commit -m "feat(ui): complete Bubble Tea migration; remove tview/tcell"
```

---

## Keybinding Reference (preserve existing, add new)

| Context | Key | Action |
|---|---|---|
| Global | `Ctrl+A` | Switch to Actions page |
| Global | `Ctrl+I` | Switch to Issues page |
| Global | `Ctrl+C` | Quit |
| Issues | `Ctrl+N` / `Ctrl+P` | Cycle panel focus |
| Issues | `Ctrl+G` | Focus issues table |
| Issues | `Ctrl+T` | Focus filter |
| Issues | `n` | New issue form |
| Issues | `e` | Edit issue body |
| Issues | `c` | Close issue |
| Issues | `o` | Reopen issue |
| Issues | `Ctrl+O` | Open in browser |
| Issues | `f` | Fetch more |
| Issues | `/` | Search / filter |
| Issues | `y` | Yank issue URL |
| Comments | `n` | New comment |
| Comments | `e` | Edit comment |
| Comments | `d` | Delete comment |
| Comments | `r` | Quote reply |
| Comments | `Ctrl+O` | Open in browser |
| Actions | `s` | Cycle status filter |
| Actions | `w` | Workflow selector |
| Actions | `r` | Refresh |
| Actions | `Enter` | Drill into run → jobs |
| Actions | `Escape` | Back from jobs to runs |
| Actions (jobs) | `Enter` | View job log |
| Actions (jobs) | `Ctrl+O` | Open in browser |

---

## Testing Strategy

- **Domain package:** unit tests in `domain/item_test.go` (already added in Task 2)
- **Theme package:** `ui/theme/theme_test.go` verifies ColorRole resolution
- **Components:** `ui/components/tabbar_test.go`, `ui/components/statusbar_test.go` verify rendering outputs contain expected strings
- **Page models:** `ui/pages/issues_test.go`, `ui/pages/actions_test.go` test `Update()` with message types — pure function, no mocks needed
- **Root model:** `ui/app_test.go` tests page switching
- **Integration:** `go build ./cmd/ght && ./ght owner/repo` to visually verify the running app

The Bubble Tea `Update()` function is pure (no I/O), making unit tests straightforward — send a `tea.Msg`, assert the returned model state.
