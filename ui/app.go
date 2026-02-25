package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/skanehira/ght/ui/pages"
	"github.com/skanehira/ght/ui/theme"
)

// AppModel is the root Bubble Tea model that routes between Issues and Actions pages.
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
