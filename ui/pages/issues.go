// Package pages contains full-screen Bubble Tea page models that make up the
// github-tui application.  Each page is a self-contained tea.Model that owns
// its sub-components and communicates with the application root through
// message types defined in this package.
package pages

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/shurcooL/githubv4"
	"github.com/skanehira/ght/config"
	"github.com/skanehira/ght/domain"
	github "github.com/skanehira/ght/github"
	"github.com/skanehira/ght/ui/components"
	"github.com/skanehira/ght/ui/theme"
)

// ---------------------------------------------------------------------------
// Focus constants
// ---------------------------------------------------------------------------

// focusedPanel identifies which panel currently holds keyboard focus.
type focusedPanel int

const (
	focusFilter         focusedPanel = iota // issue filter / search bar
	focusIssues                             // issues list table
	focusDetails                            // details panel (labels, assignees, etc.)
	focusComments                           // comments list table
	focusIssuePreview                       // issue body preview viewport
	focusCommentPreview                     // comment body preview viewport
)

// focusPanels is an ordered slice used to cycle focus with Ctrl+N / Ctrl+P.
var focusPanels = []focusedPanel{
	focusFilter,
	focusIssues,
	focusDetails,
	focusComments,
	focusIssuePreview,
	focusCommentPreview,
}

// ---------------------------------------------------------------------------
// Async message types
// ---------------------------------------------------------------------------

// IssuesLoadedMsg is sent by fetchIssues when the first page of issues has
// been fetched from the GitHub API.
type IssuesLoadedMsg struct {
	Items    []domain.Item
	PageInfo *github.PageInfo
	Err      error
}

// IssuesMoreLoadedMsg is sent by fetchMoreIssues when subsequent pages of
// issues are fetched.
type IssuesMoreLoadedMsg struct {
	Items    []domain.Item
	PageInfo *github.PageInfo
	Err      error
}

// CommentsLoadedMsg carries the selected issue whose comment slice has been
// refreshed.  (Reserved for future use; comments are currently embedded in
// the issue node at fetch time.)
type CommentsLoadedMsg struct {
	Issue *domain.Issue
}

// SwitchPageMsg is consumed by the application root to switch the visible
// top-level page.  Page is "issues" or "actions".
type SwitchPageMsg struct {
	Page string
}

// ---------------------------------------------------------------------------
// IssuesModel
// ---------------------------------------------------------------------------

// IssuesModel is the Bubble Tea model for the Issues page.  It owns every
// sub-component required to render the issues list, details, comments, and
// preview panes.
type IssuesModel struct {
	th        theme.Theme
	statusBar components.StatusBar
	tabBar    components.TabBar

	width  int
	height int

	focus   focusedPanel
	loading bool

	// GitHub query state
	query  string
	cursor *string
	hasMore bool

	// Data
	issues []domain.Item

	// Sub-components
	filter         textinput.Model
	issueTable     table.Model
	issuePreview   viewport.Model
	detailsView    viewport.Model
	commentsTable  table.Model
	commentPreview viewport.Model
	spinner        spinner.Model

	// Cached glamour renderer — recreated only when preview width changes.
	mdRenderer *glamour.TermRenderer
	mdWidth    int
}

// NewIssuesModel constructs an IssuesModel with themed sub-components and a
// filter input pre-populated with the configured owner/repo context.
func NewIssuesModel(th theme.Theme) IssuesModel {
	// Filter input ----------------------------------------------------------------
	fi := textinput.New()
	fi.Placeholder = fmt.Sprintf(
		"repo:%s/%s state:open",
		config.GitHub.Owner,
		config.GitHub.Repo,
	)
	fi.CharLimit = 256

	// Spinner ---------------------------------------------------------------------
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = th.Accent

	// Tables ----------------------------------------------------------------------
	issueTable := newTable(th)
	commentsTable := newTable(th)

	// Viewports -------------------------------------------------------------------
	issuePreview := viewport.New(0, 0)
	detailsView := viewport.New(0, 0)
	commentPreview := viewport.New(0, 0)

	return IssuesModel{
		th:             th,
		statusBar:      components.NewStatusBar(th),
		tabBar:         components.NewTabBar(th),
		focus:          focusIssues,
		loading:        true,
		filter:         fi,
		issueTable:     issueTable,
		commentsTable:  commentsTable,
		issuePreview:   issuePreview,
		detailsView:    detailsView,
		commentPreview: commentPreview,
		spinner:        sp,
	}
}

// newTable creates a themed bubbles table.Model with header and selected row
// styles pulled from the provided theme.
func newTable(th theme.Theme) table.Model {
	s := table.DefaultStyles()
	s.Header = th.TableHeader
	s.Selected = th.TableSelected
	t := table.New(
		table.WithStyles(s),
		table.WithFocused(false),
	)
	return t
}

// ---------------------------------------------------------------------------
// Exported accessors (needed by tests)
// ---------------------------------------------------------------------------

// Loading reports whether the model is still fetching the first page.
func (m IssuesModel) Loading() bool { return m.loading }

// Issues returns the current slice of fetched issue items.
func (m IssuesModel) Issues() []domain.Item { return m.issues }

// ---------------------------------------------------------------------------
// Init
// ---------------------------------------------------------------------------

// Init starts the spinner tick and triggers the initial issues fetch.
func (m IssuesModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchIssues(m.filter.Value(), nil),
	)
}

// ---------------------------------------------------------------------------
// Async fetch commands
// ---------------------------------------------------------------------------

// fetchIssues builds and dispatches a GitHub issue search for the first page.
// It prepends "is:issue" to the query when not already present.
func fetchIssues(query string, cursor *string) tea.Cmd {
	return func() tea.Msg {
		q := ensureIsIssue(query)
		var ghCursor *githubv4.String
		if cursor != nil {
			v := githubv4.String(*cursor)
			ghCursor = &v
		}
		variables := map[string]interface{}{
			"query":  githubv4.String(q),
			"first":  githubv4.Int(30),
			"cursor": ghCursor,
		}
		result, err := github.GetIssues(variables)
		if err != nil {
			return IssuesLoadedMsg{Err: err}
		}
		items, pi := convertIssuesResult(result)
		return IssuesLoadedMsg{Items: items, PageInfo: pi}
	}
}

// fetchMoreIssues fetches the next page of issues and returns IssuesMoreLoadedMsg.
func fetchMoreIssues(query string, cursor *string) tea.Cmd {
	return func() tea.Msg {
		q := ensureIsIssue(query)
		var ghCursor *githubv4.String
		if cursor != nil {
			v := githubv4.String(*cursor)
			ghCursor = &v
		}
		variables := map[string]interface{}{
			"query":  githubv4.String(q),
			"first":  githubv4.Int(30),
			"cursor": ghCursor,
		}
		result, err := github.GetIssues(variables)
		if err != nil {
			return IssuesMoreLoadedMsg{Err: err}
		}
		items, pi := convertIssuesResult(result)
		return IssuesMoreLoadedMsg{Items: items, PageInfo: pi}
	}
}

// ensureIsIssue prepends "is:issue " to the query when it is not already
// present, so searches always target issues rather than PRs.
func ensureIsIssue(query string) string {
	if !strings.Contains(query, "is:issue") {
		if query == "" {
			return "is:issue"
		}
		return "is:issue " + query
	}
	return query
}

// convertIssuesResult translates a *github.Issues API response into the
// domain.Item slice and PageInfo pointer used by the model.
func convertIssuesResult(result *github.Issues) ([]domain.Item, *github.PageInfo) {
	items := make([]domain.Item, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		issue := node.Issue.ToDomain()
		items = append(items, issue)
	}
	pi := result.PageInfo
	return items, &pi
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

// Update processes incoming messages and key events, returning an updated
// model and any follow-up commands.
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
		if msg.Err == nil {
			m.issues = msg.Items
			m = m.applyPageInfo(msg.PageInfo)
			m.rebuildIssueTable()
			if len(m.issues) > 0 {
				m.updateDetailsAndPreview(0)
			}
		}

	case IssuesMoreLoadedMsg:
		if msg.Err == nil {
			m.issues = append(m.issues, msg.Items...)
			m = m.applyPageInfo(msg.PageInfo)
			m.rebuildIssueTable()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+a":
			return m, func() tea.Msg { return SwitchPageMsg{Page: "actions"} }

		case "ctrl+c":
			return m, tea.Quit

		case "ctrl+n":
			m.focusNext()
			m.syncTableFocus()

		case "ctrl+p":
			m.focusPrev()
			m.syncTableFocus()

		case "enter":
			if m.focus == focusFilter {
				// Start a fresh search with the current filter value.
				m.loading = true
				m.issues = nil
				m.cursor = nil
				m.hasMore = false
				m.query = m.filter.Value()
				cmds = append(cmds, fetchIssues(m.query, nil))
			}

		case "f":
			if m.focus == focusIssues && m.hasMore && m.cursor != nil {
				cmds = append(cmds, fetchMoreIssues(m.query, m.cursor))
			}

		default:
			// Delegate key events to the active sub-component.
			var cmd tea.Cmd
			m, cmd = m.delegateKey(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// applyPageInfo stores cursor and hasMore from a PageInfo response.
func (m IssuesModel) applyPageInfo(pi *github.PageInfo) IssuesModel {
	if pi == nil {
		return m
	}
	m.hasMore = bool(pi.HasNextPage)
	if m.hasMore {
		s := string(pi.EndCursor)
		m.cursor = &s
	} else {
		m.cursor = nil
	}
	return m
}

// delegateKey forwards key messages to the focused sub-component and handles
// cursor-movement side effects (updating details/preview panes).
func (m IssuesModel) delegateKey(msg tea.KeyMsg) (IssuesModel, tea.Cmd) {
	var cmd tea.Cmd
	prevRow := m.issueTable.Cursor()
	prevCommentRow := m.commentsTable.Cursor()

	switch m.focus {
	case focusFilter:
		m.filter, cmd = m.filter.Update(msg)

	case focusIssues:
		m.issueTable, cmd = m.issueTable.Update(msg)
		// When the selected row changes, refresh the details and preview panes.
		if m.issueTable.Cursor() != prevRow && len(m.issues) > 0 {
			m.updateDetailsAndPreview(m.issueTable.Cursor())
		}

	case focusComments:
		m.commentsTable, cmd = m.commentsTable.Update(msg)
		if m.commentsTable.Cursor() != prevCommentRow {
			m.updateCommentPreview(m.commentsTable.Cursor())
		}

	case focusIssuePreview:
		m.issuePreview, cmd = m.issuePreview.Update(msg)

	case focusCommentPreview:
		m.commentPreview, cmd = m.commentPreview.Update(msg)

	case focusDetails:
		m.detailsView, cmd = m.detailsView.Update(msg)
	}

	return m, cmd
}

// ---------------------------------------------------------------------------
// Focus cycling helpers
// ---------------------------------------------------------------------------

func (m *IssuesModel) focusNext() {
	for i, p := range focusPanels {
		if p == m.focus {
			m.focus = focusPanels[(i+1)%len(focusPanels)]
			return
		}
	}
	m.focus = focusPanels[0]
}

func (m *IssuesModel) focusPrev() {
	for i, p := range focusPanels {
		if p == m.focus {
			idx := (i - 1 + len(focusPanels)) % len(focusPanels)
			m.focus = focusPanels[idx]
			return
		}
	}
	m.focus = focusPanels[len(focusPanels)-1]
}

// syncTableFocus propagates the current focus state to sub-components that
// need to know whether they are focused (tables, filter input).
func (m *IssuesModel) syncTableFocus() {
	// Filter input
	if m.focus == focusFilter {
		m.filter.Focus()
	} else {
		m.filter.Blur()
	}

	// Issue table
	issueStyles := tableStylesForFocus(m.th, m.focus == focusIssues)
	m.issueTable.SetStyles(issueStyles)

	// Comments table
	commentStyles := tableStylesForFocus(m.th, m.focus == focusComments)
	m.commentsTable.SetStyles(commentStyles)
}

// tableStylesForFocus returns table.Styles with the selected-row highlight
// active only when the table is focused.
func tableStylesForFocus(th theme.Theme, focused bool) table.Styles {
	s := table.DefaultStyles()
	s.Header = th.TableHeader
	if focused {
		s.Selected = th.TableSelected
	} else {
		s.Selected = lipgloss.NewStyle() // unfocused: no highlight
	}
	return s
}

// ---------------------------------------------------------------------------
// Table / pane population helpers
// ---------------------------------------------------------------------------

// rebuildIssueTable reconstructs the issue table columns and rows from the
// current m.issues slice.
func (m *IssuesModel) rebuildIssueTable() {
	// Calculate available inner width (subtract 2 for rounded border chars).
	innerW := m.issueTableWidth()
	if innerW < 1 {
		innerW = 80
	}

	cols := []table.Column{
		{Title: "Repo",   Width: 20},
		{Title: "#",      Width: 6},
		{Title: "State",  Width: 7},
		{Title: "Author", Width: 14},
		{Title: "Title",  Width: max(innerW-20-6-7-14-4, 20)},
	}

	rows := make([]table.Row, 0, len(m.issues))
	for _, item := range m.issues {
		issue, ok := item.(*domain.Issue)
		if !ok {
			continue
		}
		repoStr := fmt.Sprintf("%s/%s", issue.RepoOwner, issue.Repo)
		rows = append(rows, table.Row{
			repoStr,
			issue.Number,
			issue.State,
			issue.Author,
			issue.Title,
		})
	}

	m.issueTable.SetColumns(cols)
	m.issueTable.SetRows(rows)
}

// updateDetailsAndPreview populates the details viewport and issue preview
// for the issue at index ix.
func (m *IssuesModel) updateDetailsAndPreview(ix int) {
	if ix < 0 || ix >= len(m.issues) {
		return
	}
	issue, ok := m.issues[ix].(*domain.Issue)
	if !ok {
		return
	}

	// Build details text.
	var sb strings.Builder

	// Labels
	sb.WriteString("Labels: ")
	if len(issue.Labels) == 0 {
		sb.WriteString("(none)")
	} else {
		parts := make([]string, 0, len(issue.Labels))
		for _, l := range issue.Labels {
			parts = append(parts, l.Key())
		}
		sb.WriteString(strings.Join(parts, ", "))
	}
	sb.WriteString("\n")

	// Assignees
	sb.WriteString("Assignees: ")
	if len(issue.Assignees) == 0 {
		sb.WriteString("(none)")
	} else {
		parts := make([]string, 0, len(issue.Assignees))
		for _, a := range issue.Assignees {
			parts = append(parts, a.Key())
		}
		sb.WriteString(strings.Join(parts, ", "))
	}
	sb.WriteString("\n")

	// Milestone
	sb.WriteString("Milestone: ")
	if len(issue.MileStone) == 0 {
		sb.WriteString("(none)")
	} else {
		parts := make([]string, 0, len(issue.MileStone))
		for _, ms := range issue.MileStone {
			parts = append(parts, ms.Key())
		}
		sb.WriteString(strings.Join(parts, ", "))
	}
	sb.WriteString("\n")

	// Projects
	sb.WriteString("Projects: ")
	if len(issue.Projects) == 0 {
		sb.WriteString("(none)")
	} else {
		parts := make([]string, 0, len(issue.Projects))
		for _, p := range issue.Projects {
			parts = append(parts, p.Key())
		}
		sb.WriteString(strings.Join(parts, ", "))
	}

	m.detailsView.SetContent(sb.String())

	// Render issue body as markdown for the preview pane.
	m.issuePreview.SetContent(m.renderMarkdown(issue.Body, m.issuePreview.Width))

	// Rebuild the comments table for the selected issue.
	m.rebuildCommentsTable(issue)
}

// rebuildCommentsTable populates the comments table for a given issue.
func (m *IssuesModel) rebuildCommentsTable(issue *domain.Issue) {
	innerW := m.commentsTableWidth()
	if innerW < 1 {
		innerW = 60
	}

	cols := []table.Column{
		{Title: "Author",  Width: 16},
		{Title: "Updated", Width: max(innerW-16-2, 20)},
	}

	rows := make([]table.Row, 0, len(issue.Comments))
	for _, item := range issue.Comments {
		f := item.Fields()
		author := ""
		updated := ""
		if len(f) > 0 {
			author = f[0].Text
		}
		if len(f) > 1 {
			updated = f[1].Text
		}
		rows = append(rows, table.Row{author, updated})
	}

	m.commentsTable.SetColumns(cols)
	m.commentsTable.SetRows(rows)

	// Reset comment preview.
	m.commentPreview.SetContent("")
}

// updateCommentPreview renders the selected comment's body into the preview
// viewport.
func (m *IssuesModel) updateCommentPreview(ix int) {
	if len(m.issues) == 0 {
		return
	}
	cursorIx := m.issueTable.Cursor()
	if cursorIx < 0 || cursorIx >= len(m.issues) {
		return
	}
	issue, ok := m.issues[cursorIx].(*domain.Issue)
	if !ok {
		return
	}
	if ix < 0 || ix >= len(issue.Comments) {
		return
	}
	comment := issue.Comments[ix]

	// domain.Comment stores body in the struct, not in Fields() — type-assert.
	body := ""
	if dc, ok := comment.(*domain.Comment); ok {
		body = dc.Body
	}

	m.commentPreview.SetContent(m.renderMarkdown(body, m.commentPreview.Width))
}

// ---------------------------------------------------------------------------
// Layout recalculation
// ---------------------------------------------------------------------------

// recalcLayout distributes available terminal space among all sub-components.
// Called whenever the window size changes.
func (m *IssuesModel) recalcLayout() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	// Reserve rows for: title (1) + tabbar (1) + divider (1) + filter (3) + statusbar (1) = 7
	const reservedRows = 7
	contentH := m.height - reservedRows
	if contentH < 4 {
		contentH = 4
	}

	// Top row: 60% of content height; bottom row: 40%.
	topH := contentH * 6 / 10
	bottomH := contentH - topH
	if topH < 2 {
		topH = 2
	}
	if bottomH < 2 {
		bottomH = 2
	}

	// Inner height for tables/viewports (subtract 2 for border top+bottom).
	topInner := topH - 2
	if topInner < 1 {
		topInner = 1
	}
	bottomInner := bottomH - 2
	if bottomInner < 1 {
		bottomInner = 1
	}

	// Horizontal split — widths (as percentages of total terminal width):
	//   Top row:    details 20% | issues 45% | issuePreview 35%
	//   Bottom row: comments 55% | commentPreview 45%
	totalW := m.width

	detailsW := totalW * 20 / 100
	issuesW := totalW * 45 / 100
	issuePreviewW := totalW - detailsW - issuesW

	commentsW := totalW * 55 / 100
	commentPreviewW := totalW - commentsW

	// Inner widths (subtract 2 for border sides).
	detailsInner := max(detailsW-2, 1)
	issuesInner := max(issuesW-2, 1)
	issuePreviewInner := max(issuePreviewW-2, 1)
	commentsInner := max(commentsW-2, 1)
	commentPreviewInner := max(commentPreviewW-2, 1)

	// Apply sizes.
	m.detailsView.Width = detailsInner
	m.detailsView.Height = topInner

	m.issueTable.SetWidth(issuesInner)
	m.issueTable.SetHeight(topInner)

	m.issuePreview.Width = issuePreviewInner
	m.issuePreview.Height = topInner

	m.commentsTable.SetWidth(commentsInner)
	m.commentsTable.SetHeight(bottomInner)

	m.commentPreview.Width = commentPreviewInner
	m.commentPreview.Height = bottomInner

	m.filter.Width = totalW - 4 // 4 = border sides + padding

	// After resizing, re-render content at the new widths.
	if len(m.issues) > 0 {
		m.rebuildIssueTable()
		m.updateDetailsAndPreview(m.issueTable.Cursor())
	}
}

// issueTableWidth returns the current inner width for the issue table panel.
func (m *IssuesModel) issueTableWidth() int {
	w := m.width * 45 / 100
	return max(w-2, 1)
}

// commentsTableWidth returns the current inner width for the comments table panel.
func (m *IssuesModel) commentsTableWidth() int {
	w := m.width * 55 / 100
	return max(w-2, 1)
}

// ---------------------------------------------------------------------------
// Markdown rendering
// ---------------------------------------------------------------------------

// markdownRenderer returns a cached glamour TermRenderer for the given width.
// The renderer is recreated only when the width changes, avoiding expensive
// allocations on every cursor movement.
func (m *IssuesModel) markdownRenderer(width int) *glamour.TermRenderer {
	if width <= 0 {
		width = 80
	}
	if m.mdRenderer != nil && m.mdWidth == width {
		return m.mdRenderer
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	m.mdRenderer = r
	m.mdWidth = width
	return r
}

// renderMarkdown renders markdown content using the cached renderer.
func (m *IssuesModel) renderMarkdown(content string, width int) string {
	r := m.markdownRenderer(width)
	if r == nil {
		return content
	}
	rendered, err := r.Render(content)
	if err != nil {
		return content
	}
	return rendered
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

// View renders the complete Issues page to a string.
func (m IssuesModel) View() string {
	owner := config.GitHub.Owner
	repo := config.GitHub.Repo

	// --- Title row -----------------------------------------------------------
	titleLeft := m.th.TitleBar.Render("github-tui")
	ownerRepo := m.th.Muted.Render(owner + "/" + repo)
	titleW := m.width
	if titleW <= 0 {
		titleW = 80
	}
	gap := titleW - lipgloss.Width(titleLeft) - lipgloss.Width(ownerRepo)
	if gap < 1 {
		gap = 1
	}
	titleRow := titleLeft + strings.Repeat(" ", gap) + ownerRepo

	// --- Tab bar row ---------------------------------------------------------
	tabRow := "  " + m.tabBar.View("issues")

	// --- Horizontal divider --------------------------------------------------
	divider := m.th.Muted.Render(strings.Repeat("─", titleW))

	// --- Filter box ----------------------------------------------------------
	filterStyle := panelStyle(m.focus == focusFilter, titleW)
	filterBox := filterStyle.Render(m.filter.View())

	// --- Main content --------------------------------------------------------
	var content string
	if m.loading && len(m.issues) == 0 {
		content = m.renderLoading()
	} else {
		content = m.renderPanels()
	}

	// --- Status bar ----------------------------------------------------------
	hints := []components.KeyHint{
		{Key: "Ctrl+N", Desc: "next"},
		{Key: "Ctrl+P", Desc: "prev"},
		{Key: "n", Desc: "new"},
		{Key: "f", Desc: "fetch"},
		{Key: "/", Desc: "search"},
	}
	statusBarView := m.statusBar.View(titleW, hints, "")

	return strings.Join([]string{
		titleRow,
		tabRow,
		divider,
		filterBox,
		content,
		statusBarView,
	}, "\n")
}

// renderLoading returns a centered spinner line for the initial loading state.
func (m IssuesModel) renderLoading() string {
	msg := m.spinner.View() + " Fetching issues..."
	w := m.width
	if w <= 0 {
		w = 80
	}
	msgW := lipgloss.Width(msg)
	pad := (w - msgW) / 2
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) + msg
}

// renderPanels lays out the two-row panel grid:
//
//	Top row:    [details 20%] [issues 45%] [issue preview 35%]
//	Bottom row: [comments 55%] [comment preview 45%]
func (m IssuesModel) renderPanels() string {
	// Top row
	detailsPanel := panelStyle(m.focus == focusDetails, m.width*20/100).
		Height(m.detailsView.Height).
		Render(m.detailsView.View())

	issuesPanel := panelStyle(m.focus == focusIssues, m.width*45/100).
		Height(m.issueTable.Height()).
		Render(m.issueTable.View())

	issuePreviewPanel := panelStyle(m.focus == focusIssuePreview, m.width-m.width*20/100-m.width*45/100).
		Height(m.issuePreview.Height).
		Render(m.issuePreview.View())

	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		detailsPanel,
		issuesPanel,
		issuePreviewPanel,
	)

	// Bottom row
	commentsW := m.width * 55 / 100
	commentPreviewW := m.width - commentsW

	commentsPanel := panelStyle(m.focus == focusComments, commentsW).
		Height(m.commentsTable.Height()).
		Render(m.commentsTable.View())

	commentPreviewPanel := panelStyle(m.focus == focusCommentPreview, commentPreviewW).
		Height(m.commentPreview.Height).
		Render(m.commentPreview.View())

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top,
		commentsPanel,
		commentPreviewPanel,
	)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)
}

// panelStyle returns a themed rounded-border style sized to width.
// When focused is true it uses the accent-colored border.
func panelStyle(focused bool, width int) lipgloss.Style {
	var base lipgloss.Style
	if focused {
		base = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Palette.BorderFocus))
	} else {
		base = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Palette.Border))
	}
	if width > 2 {
		base = base.Width(width - 2)
	}
	return base
}

