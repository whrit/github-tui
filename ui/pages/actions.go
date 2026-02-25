package pages

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gogithub "github.com/google/go-github/v68/github"
	"github.com/shurcooL/githubv4"
	"github.com/skanehira/ght/config"
	"github.com/skanehira/ght/domain"
	github "github.com/skanehira/ght/github"
	"github.com/skanehira/ght/ui/components"
	"github.com/skanehira/ght/ui/theme"
)

// ---------------------------------------------------------------------------
// View state
// ---------------------------------------------------------------------------

type actionsView int

const (
	viewRuns actionsView = iota
	viewJobs
	viewLog
)

var statusFilterCycle = []string{"", "success", "failure", "in_progress", "queued"}

// ---------------------------------------------------------------------------
// Async message types
// ---------------------------------------------------------------------------

// RunsLoadedMsg is sent when workflow runs have been fetched.
type RunsLoadedMsg struct {
	Items    []domain.Item
	PageInfo *github.PageInfo
	Err      error
}

// JobsLoadedMsg is sent when workflow jobs have been fetched for a run.
type JobsLoadedMsg struct {
	Items []domain.Item
	Err   error
}

// LogLoadedMsg is sent when a job log has been fetched.
type LogLoadedMsg struct {
	Content   string
	Truncated bool
	Err       error
}

// ---------------------------------------------------------------------------
// ActionsModel
// ---------------------------------------------------------------------------

// ActionsModel is the Bubble Tea model for the Actions page. It owns every
// sub-component required to render workflow runs, jobs, and log viewing.
type ActionsModel struct {
	th        theme.Theme
	statusBar components.StatusBar
	tabBar    components.TabBar

	width  int
	height int

	view         actionsView
	loading      bool
	statusFilter string
	statusIx     int
	workflowID   int64
	workflowName string
	workflows    []*gogithub.Workflow
	cursor       *string
	hasMore      bool

	currentRunID   int64
	currentRunName string

	runs []domain.Item
	jobs []domain.Item

	runsTable table.Model
	jobsTable table.Model
	logView   viewport.Model
	sp        spinner.Model
}

// NewActionsModel constructs an ActionsModel with themed sub-components.
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

// ---------------------------------------------------------------------------
// Exported accessors (needed by tests)
// ---------------------------------------------------------------------------

// Loading reports whether the model is still fetching data.
func (m ActionsModel) Loading() bool { return m.loading }

// ---------------------------------------------------------------------------
// Init
// ---------------------------------------------------------------------------

// Init starts the spinner tick and triggers the initial workflow runs fetch.
func (m ActionsModel) Init() tea.Cmd {
	return tea.Batch(m.sp.Tick, fetchRuns("", 0, nil))
}

// ---------------------------------------------------------------------------
// Async fetch commands
// ---------------------------------------------------------------------------

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
			page, err := strconv.Atoi(*cursor)
			if err == nil {
				opts.ListOptions.Page = page
			}
		}

		owner, repo := config.GitHub.Owner, config.GitHub.Repo
		var runs *gogithub.WorkflowRuns
		var resp *gogithub.Response
		var err error
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

func fetchLog(jobID int64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		content, truncated, err := github.GetWorkflowJobLog(ctx, config.GitHub.Owner, config.GitHub.Repo, jobID)
		return LogLoadedMsg{Content: content, Truncated: truncated, Err: err}
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

// Update processes incoming messages and key events.
func (m ActionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcLayout()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		cmds = append(cmds, cmd)

	case RunsLoadedMsg:
		m.loading = false
		if msg.Err == nil {
			m.runs = msg.Items
			m.rebuildRunsTable()
			if msg.PageInfo != nil {
				m.hasMore = bool(msg.PageInfo.HasNextPage)
				if m.hasMore {
					s := string(msg.PageInfo.EndCursor)
					m.cursor = &s
				} else {
					m.cursor = nil
				}
			}
		}

	case JobsLoadedMsg:
		m.loading = false
		if msg.Err == nil {
			m.jobs = msg.Items
			m.rebuildJobsTable()
		}

	case LogLoadedMsg:
		m.loading = false
		if msg.Err == nil {
			content := msg.Content
			if msg.Truncated {
				content += "\n\n--- Log truncated at 10MB. Press Ctrl+O on the job to view full log in browser. ---"
			}
			m.logView.SetContent(content)
			m.logView.GotoTop()
			m.view = viewLog
		}

	case tea.KeyMsg:
		switch m.view {
		case viewRuns:
			switch msg.Type {
			case tea.KeyCtrlI:
				return m, func() tea.Msg { return SwitchPageMsg{Page: "issues"} }
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEnter:
				ix := m.runsTable.Cursor()
				if ix < len(m.runs) {
					run, ok := m.runs[ix].(*domain.WorkflowRun)
					if ok {
						m.currentRunID = run.ID
						m.currentRunName = fmt.Sprintf("#%d - %s", run.RunNumber, run.Name)
						m.view = viewJobs
						m.loading = true
						cmds = append(cmds, m.sp.Tick, fetchJobs(m.currentRunID))
					}
				}
			default:
				if msg.Type == tea.KeyRunes {
					switch string(msg.Runes) {
					case "s":
						m.statusIx = (m.statusIx + 1) % len(statusFilterCycle)
						m.statusFilter = statusFilterCycle[m.statusIx]
						m.loading = true
						m.cursor = nil
						cmds = append(cmds, m.sp.Tick, fetchRuns(m.statusFilter, m.workflowID, nil))
					case "r":
						m.loading = true
						cmds = append(cmds, m.sp.Tick, fetchRuns(m.statusFilter, m.workflowID, m.cursor))
					}
				}
				// Delegate to runsTable for navigation.
				var cmd tea.Cmd
				m.runsTable, cmd = m.runsTable.Update(msg)
				cmds = append(cmds, cmd)
			}

		case viewJobs:
			switch msg.Type {
			case tea.KeyEscape:
				m.view = viewRuns
				m.loading = false
			case tea.KeyCtrlI:
				return m, func() tea.Msg { return SwitchPageMsg{Page: "issues"} }
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEnter:
				ix := m.jobsTable.Cursor()
				if ix < len(m.jobs) {
					job, ok := m.jobs[ix].(*domain.WorkflowJob)
					if ok {
						m.loading = true
						cmds = append(cmds, m.sp.Tick, fetchLog(job.ID))
					}
				}
			case tea.KeyRunes:
				if string(msg.Runes) == "r" {
					m.loading = true
					cmds = append(cmds, m.sp.Tick, fetchJobs(m.currentRunID))
				}
				var cmd tea.Cmd
				m.jobsTable, cmd = m.jobsTable.Update(msg)
				cmds = append(cmds, cmd)
			default:
				var cmd tea.Cmd
				m.jobsTable, cmd = m.jobsTable.Update(msg)
				cmds = append(cmds, cmd)
			}

		case viewLog:
			switch msg.Type {
			case tea.KeyEscape:
				m.view = viewJobs
			case tea.KeyCtrlC:
				return m, tea.Quit
			default:
				var cmd tea.Cmd
				m.logView, cmd = m.logView.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// ---------------------------------------------------------------------------
// Table / pane population helpers
// ---------------------------------------------------------------------------

// rebuildRunsTable reconstructs the runs table columns and rows.
func (m *ActionsModel) rebuildRunsTable() {
	cols := []table.Column{
		{Title: "Status", Width: 12},
		{Title: "Workflow", Width: 30},
		{Title: "Branch", Width: 20},
		{Title: "Event", Width: 12},
		{Title: "Duration", Width: 10},
	}

	rows := make([]table.Row, 0, len(m.runs))
	for _, item := range m.runs {
		run, ok := item.(*domain.WorkflowRun)
		if !ok {
			continue
		}
		fields := run.Fields()
		statusText := ""
		if len(fields) > 0 {
			statusText = fields[0].Text
		}
		rows = append(rows, table.Row{
			statusText,
			run.Name,
			run.HeadBranch,
			run.Event,
			run.Duration,
		})
	}

	m.runsTable.SetColumns(cols)
	m.runsTable.SetRows(rows)
}

// rebuildJobsTable reconstructs the jobs table columns and rows.
func (m *ActionsModel) rebuildJobsTable() {
	cols := []table.Column{
		{Title: "Status", Width: 12},
		{Title: "Job", Width: 40},
		{Title: "Duration", Width: 10},
	}

	rows := make([]table.Row, 0, len(m.jobs))
	for _, item := range m.jobs {
		job, ok := item.(*domain.WorkflowJob)
		if !ok {
			continue
		}
		fields := job.Fields()
		statusText := ""
		if len(fields) > 0 {
			statusText = fields[0].Text
		}
		rows = append(rows, table.Row{
			statusText,
			job.Name,
			job.Duration,
		})
	}

	m.jobsTable.SetColumns(cols)
	m.jobsTable.SetRows(rows)
}

// ---------------------------------------------------------------------------
// Layout recalculation
// ---------------------------------------------------------------------------

func (m *ActionsModel) recalcLayout() {
	contentH := m.height - 4 // header + tab + divider + statusbar
	if contentH < 3 {
		contentH = 3
	}
	m.runsTable.SetWidth(m.width)
	m.runsTable.SetHeight(contentH)
	m.jobsTable.SetWidth(m.width)
	m.jobsTable.SetHeight(contentH)
	m.logView.Width = m.width
	m.logView.Height = contentH
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

// View renders the complete Actions page to a string.
func (m ActionsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Title row
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Border))
	titleRight := m.th.Muted.Render(fmt.Sprintf("%s/%s", config.GitHub.Owner, config.GitHub.Repo))
	titleLeft := m.th.TitleBar.Render("github-tui")
	gap := m.width - lipgloss.Width(titleLeft) - lipgloss.Width(titleRight)
	if gap < 1 {
		gap = 1
	}
	titleRow := titleLeft + fmt.Sprintf("%*s", gap, "") + titleRight

	tabRow := "  " + m.tabBar.View("actions")
	divider := dividerStyle.Render(strings.Repeat("─", m.width))

	// Content
	var content string
	if m.loading {
		content = m.renderLoading()
	} else {
		switch m.view {
		case viewRuns:
			content = m.th.Panel.Width(m.width - 2).Render(m.runsTable.View())
		case viewJobs:
			content = m.th.Panel.Width(m.width - 2).Render(m.jobsTable.View())
		case viewLog:
			content = m.th.Panel.Width(m.width - 2).Render(m.logView.View())
		}
	}

	// Status bar
	hints, ctx := m.statusHints()
	status := m.statusBar.View(m.width, hints, ctx)

	return lipgloss.JoinVertical(lipgloss.Left, titleRow, tabRow, divider, content, status)
}

func (m ActionsModel) renderLoading() string {
	h := m.height - 8
	if h < 3 {
		h = 3
	}
	return m.th.Muted.Width(m.width).Height(h).
		Render(strings.Repeat("\n", h/2) + fmt.Sprintf("    %s Loading...", m.sp.View()))
}

func (m ActionsModel) statusHints() ([]components.KeyHint, string) {
	switch m.view {
	case viewRuns:
		statusLabel := m.statusFilter
		if statusLabel == "" {
			statusLabel = "all"
		}
		wfLabel := m.workflowName
		if wfLabel == "" {
			wfLabel = "all"
		}
		hints := []components.KeyHint{
			{Key: "s", Desc: "status"},
			{Key: "r", Desc: "refresh"},
			{Key: "Enter", Desc: "jobs"},
		}
		ctx := fmt.Sprintf("Status: %s | Workflow: %s", statusLabel, wfLabel)
		return hints, ctx
	case viewJobs:
		hints := []components.KeyHint{
			{Key: "Esc", Desc: "back"},
			{Key: "r", Desc: "refresh"},
			{Key: "Enter", Desc: "log"},
		}
		return hints, fmt.Sprintf("Run: %s", m.currentRunName)
	case viewLog:
		return []components.KeyHint{{Key: "Esc", Desc: "close"}}, "Log view"
	}
	return nil, ""
}

