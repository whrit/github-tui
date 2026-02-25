package pages

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shurcooL/githubv4"
	github "github.com/skanehira/ght/github"
	"github.com/skanehira/ght/ui/theme"
)

// ---------------------------------------------------------------------------
// FormField indices
// ---------------------------------------------------------------------------

const (
	fieldTitle     = iota
	fieldAssignees
	fieldLabels
	fieldProjects
	fieldMilestone
	fieldCount
)

// ---------------------------------------------------------------------------
// Message types
// ---------------------------------------------------------------------------

// IssueCreatedMsg signals successful issue creation.
type IssueCreatedMsg struct{}

// FormDismissedMsg signals the form was cancelled.
type FormDismissedMsg struct{}

// repoIDLoadedMsg carries the result of a fetchRepoID command.
type repoIDLoadedMsg struct {
	id  githubv4.ID
	err error
}

// issueSubmitErrMsg carries an error that occurred during issue submission.
type issueSubmitErrMsg struct{ err error }

// ---------------------------------------------------------------------------
// CreateIssueFormModel
// ---------------------------------------------------------------------------

// CreateIssueFormModel is a multi-field form for creating issues.
type CreateIssueFormModel struct {
	th      theme.Theme
	inputs  []textinput.Model
	focused int
	repoID  githubv4.ID
	owner   string
	repo    string
	err     error
	done    bool
}

// NewCreateIssueForm constructs a CreateIssueFormModel with themed inputs and
// the first field focused.
func NewCreateIssueForm(th theme.Theme, owner, repo string) CreateIssueFormModel {
	labels := []string{"Title", "Assignees (comma-separated)", "Labels (comma-separated)", "Projects", "Milestone"}
	inputs := make([]textinput.Model, fieldCount)
	for i, label := range labels {
		ti := textinput.New()
		ti.Placeholder = label
		ti.Width = 60
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

// ---------------------------------------------------------------------------
// Init
// ---------------------------------------------------------------------------

// Init starts the repoID fetch needed for the CreateIssue mutation.
func (m CreateIssueFormModel) Init() tea.Cmd {
	return fetchRepoID(m.owner, m.repo)
}

// ---------------------------------------------------------------------------
// Async commands
// ---------------------------------------------------------------------------

// fetchRepoID fetches the repository ID from the GitHub API.
func fetchRepoID(owner, repo string) tea.Cmd {
	return func() tea.Msg {
		v := map[string]interface{}{
			"owner": githubv4.String(owner),
			"name":  githubv4.String(repo),
		}
		r, err := github.GetRepo(v)
		if err != nil {
			return repoIDLoadedMsg{err: err}
		}
		return repoIDLoadedMsg{id: r.ID}
	}
}

// submitIssue submits the new issue via the GitHub GraphQL API.
func submitIssue(repoID githubv4.ID, title string) tea.Cmd {
	return func() tea.Msg {
		input := githubv4.CreateIssueInput{
			RepositoryID: repoID,
			Title:        githubv4.String(title),
		}
		if err := github.CreateIssue(input); err != nil {
			return issueSubmitErrMsg{err: err}
		}
		return IssueCreatedMsg{}
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

// Update processes incoming messages and key events.
func (m CreateIssueFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case repoIDLoadedMsg:
		if msg.err == nil {
			m.repoID = msg.id
		} else {
			m.err = msg.err
		}

	case issueSubmitErrMsg:
		m.err = msg.err

	case IssueCreatedMsg:
		m.done = true
		return m, func() tea.Msg { return IssueCreatedMsg{} }

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return m, func() tea.Msg { return FormDismissedMsg{} }

		case tea.KeyTab:
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % fieldCount
			m.inputs[m.focused].Focus()

		case tea.KeyShiftTab:
			m.inputs[m.focused].Blur()
			m.focused = (m.focused - 1 + fieldCount) % fieldCount
			m.inputs[m.focused].Focus()

		case tea.KeyEnter:
			// Submit on Enter when on last field or any field if repoID is set.
			title := m.inputs[fieldTitle].Value()
			if title != "" && m.repoID != nil {
				return m, submitIssue(m.repoID, title)
			}

		default:
			// Delegate key to focused input.
			var cmd tea.Cmd
			m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

// View renders the create-issue form.
func (m CreateIssueFormModel) View() string {
	fieldLabels := []string{"Title", "Assignees", "Labels", "Projects", "Milestone"}

	var b strings.Builder
	b.WriteString(m.th.TitleBar.Render("Create Issue") + "\n\n")

	for i, input := range m.inputs {
		label := fieldLabels[i]
		if i == m.focused {
			b.WriteString(m.th.Accent.Render("▶ "+label+": "))
		} else {
			b.WriteString(m.th.Muted.Render("  "+label+": "))
		}
		b.WriteString(input.View() + "\n\n")
	}

	if m.err != nil {
		b.WriteString(m.th.Danger.Render("Error: "+m.err.Error()) + "\n")
	}

	hints := "\n" + m.th.Muted.Render("Tab/Shift+Tab: navigate  Enter: submit  Esc: cancel")
	b.WriteString(hints)

	return m.th.Panel.Render(b.String())
}

// ---------------------------------------------------------------------------
// Exported accessors
// ---------------------------------------------------------------------------

// Done reports whether the form has completed successfully (issue created).
func (m CreateIssueFormModel) Done() bool { return m.done }
