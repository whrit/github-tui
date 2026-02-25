# ght — Usage Guide

`ght` is a keyboard-driven terminal UI client for GitHub. It provides issue browsing, comment reading, and GitHub Actions workflow inspection directly in your terminal, without leaving the keyboard.

---

## Table of Contents

1. [Installation and Configuration](#1-installation-and-configuration)
2. [Starting ght](#2-starting-ght)
3. [Interface Layout](#3-interface-layout)
4. [Global Keys](#4-global-keys)
5. [Issues Page](#5-issues-page)
6. [Actions Page](#6-actions-page)
7. [Create Issue Form](#7-create-issue-form)
8. [GitHub Search Syntax](#8-github-search-syntax)
9. [Tips and Constraints](#9-tips-and-constraints)

---

## 1. Installation and Configuration

### Building

```bash
# Build the binary into the current directory
go build ./cmd/ght

# Or install it into your Go bin path
go install ./cmd/ght
```

### First-run token setup

On the first run, if no config file is found, `ght` prompts you interactively for a GitHub Personal Access Token and writes it to disk:

```
Welcome to ght! No config file found.
You need a GitHub Personal Access Token (classic) with repo, workflow, and read:org scopes.
Create one at: https://github.com/settings/tokens

Enter your GitHub token: <your token here>
```

The token is written to `config.yaml` in the platform-specific config directory:

| Platform | Location |
|----------|----------|
| macOS / Linux | `$XDG_CONFIG_HOME/ght/config.yaml` (or `~/.config/ght/config.yaml`) |
| Windows | `%AppData%\ght\config.yaml` |

The config file format is:

```yaml
github:
  token: ghp_your_token_here
```

### Required token scopes (classic PAT)

At startup, `ght` validates your token's scopes before launching the UI. If required scopes are missing, the program exits immediately with a clear error message.

| Scope | Purpose |
|-------|---------|
| `repo` | Read issues, comments, labels, milestones, projects; create issues; access Actions |
| `project` or `read:org` | Read GitHub Projects associated with issues |

Fine-grained PATs are accepted but cannot be scope-validated via API headers. If you use a fine-grained PAT and encounter permission errors, verify that the token covers `Contents`, `Issues`, `Actions`, and `Projects` read permissions for the target repository.

A debug log is written to `<config-dir>/ght/debug.log` on every run.

---

## 2. Starting ght

### Auto-detect repository from git remote

Run `ght` from inside a cloned repository. It reads `git remote get-url origin` and parses the owner and repo name from HTTPS, SSH, and git-protocol remote URLs:

```bash
cd ~/projects/my-repo
ght
```

### Specify a repository explicitly

Pass `owner/repo` as the first positional argument to override auto-detection:

```bash
ght skanehira/ght
ght torvalds/linux
```

If neither a valid git remote nor a `owner/repo` argument is provided, `ght` exits with an error.

---

## 3. Interface Layout

### Issues page layout

```
┌ github-tui                               owner/repo ─────────────────┐
│  ● Issues    ○ Actions                                                │
├───────────────────────────────────────────────────────────────────────┤
│ [ repo:owner/repo state:open                                        ] │
├──────────────────┬────────────────────────────┬───────────────────────┤
│  Details  (20%)  │  Issues table  (45%)        │  Issue Preview  (35%) │
│                  │ ──────────────────────────  │                       │
│  Labels: bug     │  Repo    #   State  Author  │  ## Fix login bug     │
│  Assignees: alice│  ow/re  42  open   alice    │                       │
│  Milestone: v2   │  ow/re  41  open   bob      │  Body text...         │
│  Projects: (none)│                             │                       │
├──────────────────┴────────────────┬───────────┴───────────────────────┤
│  Comments table  (55%)            │  Comment Preview  (45%)           │
│ ────────────────────────────────  │                                   │
│  Author            Updated        │  Comment body text...             │
│  alice             Jan 10 14:22   │                                   │
├───────────────────────────────────┴───────────────────────────────────┤
│  Ctrl+N:next  Ctrl+P:prev  n:new  f:fetch  /:search     (context)    │
└───────────────────────────────────────────────────────────────────────┘
```

The title row always shows `github-tui` on the left and `owner/repo` on the right. The tab bar below it indicates the active page with a filled circle (`●`) and inactive pages with an open circle (`○`). Focused panels are highlighted with an accent-colored rounded border; unfocused panels use a dimmed border.

### Actions page layout

```
┌ github-tui                               owner/repo ─────────────────┐
│  ○ Issues    ● Actions                                                │
├───────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  Status   Workflow          Branch     Event    Duration              │
│  success  CI                main       push     1m 32s               │
│  failure  Deploy            main       push     4m 07s               │
│  ...                                                                  │
│                                                                       │
├───────────────────────────────────────────────────────────────────────┤
│  s:status  r:refresh  Enter:jobs          Status: all | Workflow: all │
└───────────────────────────────────────────────────────────────────────┘
```

The Actions page uses a single full-width table that transitions between three views (runs, jobs, log) as you drill down. The status bar context area on the right updates to reflect the current view and active filters.

---

## 4. Global Keys

These keys work regardless of which page is active.

| Key | Action |
|-----|--------|
| `Ctrl+C` | Quit ght |
| `Ctrl+A` | Switch to the Actions page |
| `Ctrl+I` | Switch to the Issues page |

`Ctrl+C` is handled at the application root level and will quit from any view, including mid-drill-down in Actions.

---

## 5. Issues Page

The Issues page is the default view on startup. It consists of six panels arranged in two rows.

### Panel focus order

Focus cycles through panels in a fixed order using `Ctrl+N` (forward) and `Ctrl+P` (backward). The focused panel is visually indicated by an accent-colored border. Only the focused panel receives keyboard input.

| Order | Panel | Description |
|-------|-------|-------------|
| 1 | Filter bar | GitHub search query input |
| 2 | Issues table | Scrollable list of issues |
| 3 | Details panel | Labels, assignees, milestone, and projects for the selected issue |
| 4 | Comments table | List of comments for the selected issue |
| 5 | Issue preview | Rendered markdown body of the selected issue |
| 6 | Comment preview | Rendered markdown body of the selected comment |

After the last panel (`Comment preview`), `Ctrl+N` wraps back to the Filter bar. Similarly, `Ctrl+P` from the Filter bar wraps to `Comment preview`.

### Panel focus keys

| Key | Action |
|-----|--------|
| `Ctrl+N` | Move focus to the next panel |
| `Ctrl+P` | Move focus to the previous panel |

### Filter bar

The filter bar is a text input pre-populated with a placeholder showing the default search:

```
repo:owner/repo state:open
```

The placeholder text reflects the owner and repo detected at startup. It is not sent automatically — you must explicitly enter a query and press `Enter`.

| Key | Action |
|-----|--------|
| `Enter` | Submit the current query and reload the issues table |
| Any printable character | Edit the query string |

When `Enter` is pressed, the issues table clears and a loading spinner appears while the first 30 results are fetched. The query string is internally prefixed with `is:issue` if it is not already present, so results always target issues rather than pull requests.

The filter bar accepts any GitHub issues search syntax. See [Section 8](#8-github-search-syntax) for examples.

### Issues table

The issues table displays up to 30 issues per page, fetched from the GitHub GraphQL API. Columns shown:

| Column | Content |
|--------|---------|
| Repo | `owner/repo` of the issue |
| # | Issue number |
| State | `open` or `closed` |
| Author | Username of the issue author |
| Title | Issue title (fills remaining width) |

When focus is on the Issues table, all standard table navigation keys from the `charmbracelet/bubbles` table component apply:

| Key | Action |
|-----|--------|
| `Up` / `k` | Move selection up one row |
| `Down` / `j` | Move selection down one row |
| `f` | Fetch the next page of results (only active when more pages exist) |

Changing the selected row automatically updates the Details panel, Issue preview, and Comments table to reflect the newly selected issue.

The `f` key only triggers a fetch if there are more pages available (indicated by the GitHub API cursor). If the current page is the last, `f` has no effect.

### Details panel

The Details panel is a read-only viewport showing metadata for the currently selected issue:

```
Labels: bug, enhancement
Assignees: alice, bob
Milestone: v2.0
Projects: Q1 Roadmap
```

Any field with no data shows `(none)`. When this panel has focus, standard viewport scroll keys (arrow keys, Page Up/Down) apply if the content exceeds the panel height.

### Issue preview

Displays the body of the selected issue rendered as markdown using the `glamour` library with a dark terminal style. When this panel has focus, it is scrollable with standard viewport navigation keys.

### Comments table

Displays comments for the selected issue with two columns:

| Column | Content |
|--------|---------|
| Author | Username of the comment author |
| Updated | Date and time of the last update |

Changing the selected row in the Comments table automatically updates the Comment preview panel.

### Comment preview

Displays the body of the selected comment rendered as markdown. When focused, standard viewport scroll keys apply.

### Issues page status bar

The status bar at the bottom of the Issues page always shows the following key hints:

```
Ctrl+N:next  Ctrl+P:prev  n:new  f:fetch  /:search
```

Note: The `n:new` hint refers to the Create Issue form, which is implemented but not yet wired to the `n` key in the current build. See [Section 7](#7-create-issue-form).

The `/` hint is displayed but search is handled through the Filter bar by focusing it with `Ctrl+N`/`Ctrl+P` and typing directly.

---

## 6. Actions Page

The Actions page is accessed via `Ctrl+A` from anywhere. It presents GitHub Actions workflow run data in a three-level drill-down: runs, then jobs for a selected run, then the log for a selected job.

Each level is rendered in a single full-width table or viewport. The status bar at the bottom changes to reflect the current level and available keys.

### Level 1 — Runs view (default)

Lists the most recent workflow runs for the repository (30 per page via the GitHub REST API).

**Table columns:**

| Column | Content |
|--------|---------|
| Status | Run outcome: `success`, `failure`, `in_progress`, `queued`, or other conclusion/status values |
| Workflow | Name of the workflow |
| Branch | Head branch the run was triggered on |
| Event | Trigger event (e.g., `push`, `pull_request`, `workflow_dispatch`) |
| Duration | Elapsed time formatted as `Xs`, `Xm Ys`, or `Xh Ym` |

**Status values displayed:**

For completed runs, the `Status` column shows the conclusion (`success`, `failure`, `cancelled`, `skipped`). For in-progress runs it shows `in_progress`. For queued runs it shows `queued`.

**Keys in the Runs view:**

| Key | Action |
|-----|--------|
| `Up` / `Down` | Navigate rows |
| `Enter` | Drill into the Jobs view for the selected run |
| `s` | Cycle the status filter |
| `r` | Refresh the runs list |
| `Ctrl+I` | Switch to the Issues page |
| `Ctrl+C` | Quit |

**Status filter cycle:**

Pressing `s` advances through this cycle and immediately re-fetches:

```
all → success → failure → in_progress → queued → all → ...
```

The current filter is shown in the status bar context area on the right:

```
Status: success | Workflow: all
```

### Level 2 — Jobs view

After pressing `Enter` on a run, the table transitions to show all jobs for that run. The status bar context updates to show the run identifier.

**Table columns:**

| Column | Content |
|--------|---------|
| Status | Job outcome (same values as runs) |
| Job | Job name |
| Duration | Elapsed time |

**Keys in the Jobs view:**

| Key | Action |
|-----|--------|
| `Up` / `Down` | Navigate rows |
| `Enter` | Open the Log view for the selected job |
| `r` | Refresh the jobs list |
| `Esc` | Return to the Runs view |
| `Ctrl+I` | Switch to the Issues page |
| `Ctrl+C` | Quit |

The status bar while in Jobs view:

```
Esc:back  r:refresh  Enter:log          Run: #42 - CI
```

The run identifier format is `#<run-number> - <workflow-name>`.

### Level 3 — Log view

After pressing `Enter` on a job, the log for that job is fetched and displayed in a scrollable viewport. Log content is cleaned of ANSI escape sequences and GitHub's timestamp prefixes before display.

**Keys in the Log view:**

| Key | Action |
|-----|--------|
| `Up` / `Down` | Scroll one line |
| `Page Up` / `Page Down` | Scroll one page |
| `Esc` | Return to the Jobs view |
| `Ctrl+C` | Quit |

The status bar while in Log view:

```
Esc:close                                Log view
```

The log viewport starts at the top of the output when first opened.

**Log size limit:**

Log downloads are capped at 10MB. If the log exceeds this limit, the following notice is appended to the visible content:

```
--- Log truncated at 10MB. Press Ctrl+O on the job to view full log in browser. ---
```

Log fetching has a 30-second timeout. If the download times out or fails, `ght` remains in the Jobs view and the error is written to the debug log.

---

## 7. Create Issue Form

The Create Issue form is a fully implemented component (`CreateIssueFormModel`) that can be used to open a new issue against the current repository via the GitHub GraphQL API.

**Current status:** The form implementation is complete and tested in isolation, but the `n` key binding that would launch it from the Issues page is not yet wired in the current build. The `n:new` hint in the status bar is a forward declaration — the feature is coming. The form can be referenced in `ui/pages/form.go`.

### Form fields

When integrated, the form will present five input fields navigated top to bottom:

| Field | Placeholder | Notes |
|-------|-------------|-------|
| Title | `Title` | Required. The issue will not be submitted without a non-empty title. |
| Assignees | `Assignees (comma-separated)` | Comma-separated GitHub usernames |
| Labels | `Labels (comma-separated)` | Comma-separated label names |
| Projects | `Projects` | Project name |
| Milestone | `Milestone` | Milestone title |

The currently focused field is indicated by an accent-colored arrow prefix (`▶ FieldName:`). Unfocused fields use a muted style (`  FieldName:`).

### Form keys

| Key | Action |
|-----|--------|
| `Tab` | Move focus to the next field |
| `Shift+Tab` | Move focus to the previous field |
| `Enter` | Submit the form (requires a non-empty Title field and a resolved repository ID) |
| `Esc` | Cancel and dismiss the form |

Focus wraps around: `Tab` from the last field (Milestone) moves to the first (Title), and `Shift+Tab` from Title moves to Milestone.

### Submission behavior

On submit, `ght` sends a `CreateIssue` GraphQL mutation containing at minimum the repository node ID and the issue title. If submission succeeds, the form closes. If an error occurs (network failure, insufficient permissions, etc.), the error message is displayed in red below the fields.

The repository ID is fetched asynchronously when the form initializes; submission is gated on its arrival.

---

## 8. GitHub Search Syntax

The filter bar on the Issues page accepts the full GitHub search query syntax for issues. The string `is:issue` is automatically prepended if absent, so your queries only need to include the filtering predicates beyond that.

### Common filters

```bash
# All open issues in the repo (default placeholder)
repo:owner/repo state:open

# Filter by label
is:open label:bug

# Filter by multiple labels (both must match)
is:open label:bug label:priority:high

# Filter by assignee
is:open assignee:alice

# Issues assigned to yourself
is:open assignee:@me

# Filter by author
is:open author:bob

# Filter by milestone
is:open milestone:"v2.0"

# Issues mentioning a user
is:open mentions:alice

# Closed issues
is:closed label:bug

# Issues with no assignee
is:open no:assignee

# Text search in title and body
is:open login bug in:title

# Combine filters
is:open label:bug assignee:alice milestone:"v2.0"
```

### How queries are sent

Each time you press `Enter` in the filter bar, `ght` clears the current issue list, resets pagination, and issues a fresh GraphQL search query. Results are returned 30 at a time. Use `f` in the Issues table to fetch additional pages when more are available.

The `is:issue` qualifier is silently prepended if it is not already part of your query string, ensuring that pull requests are never included in results.

---

## 9. Tips and Constraints

### Focus and keyboard routing

Key events are routed exclusively to the focused panel. If pressing `j`, `k`, or arrow keys is not moving the expected list, check which panel currently has an accent-colored border. Use `Ctrl+N` or `Ctrl+P` to move focus to the correct panel.

### Panel sizing

Panels are sized as fixed percentages of the terminal width and height:

| Panel | Width | Row |
|-------|-------|-----|
| Details | 20% | Top |
| Issues table | 45% | Top |
| Issue preview | 35% (remainder) | Top |
| Comments table | 55% | Bottom |
| Comment preview | 45% (remainder) | Bottom |

The top row takes 60% of the available content height; the bottom row takes the remaining 40%. Resize your terminal window and the layout recalculates immediately.

### Markdown rendering

Issue and comment bodies are rendered using `glamour` with the `dark` terminal style. The renderer is cached and only rebuilt when the panel width changes, so resizing is efficient. If glamour encounters a rendering error, the raw markdown source is displayed as a fallback.

### Rate limiting

All GitHub API calls share a single rate limiter (`RateLimiter` in `github/rate_limiter.go`) that uses a token bucket to respect GitHub's API limits and a concurrency semaphore to prevent request storms. You are unlikely to hit rate limits during normal interactive use, but bulk operations (fetching many pages in rapid succession) may cause brief pauses.

### Actions log cleaning

Log content downloaded from GitHub Actions has two layers of noise removed before display:

1. **ANSI escape sequences** — color codes and cursor movement sequences are stripped entirely.
2. **Timestamp prefixes** — GitHub prepends each log line with an ISO 8601 timestamp (e.g., `2024-01-15T10:30:45.1234567Z `). These are removed, leaving only the raw command output.

### Token type differences

| Token type | Scope validation | Behavior on insufficient permissions |
|------------|-----------------|--------------------------------------|
| Classic PAT | Full — checked at startup via `X-OAuth-Scopes` header | `ght` exits with a clear error listing missing scopes |
| Fine-grained PAT | None — header not present | `ght` starts with a warning in the debug log; API calls may fail at runtime |

Use a classic PAT with `repo`, `workflow`, and `project` (or `read:org`) scopes for the most predictable experience.

### Debug log

Every run writes a debug log to `<config-dir>/ght/debug.log`. If `ght` exits unexpectedly or an API call silently fails, check this file for the underlying error message.
