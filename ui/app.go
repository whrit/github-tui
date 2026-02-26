package ui

import (
	"github.com/charmbracelet/lipgloss"
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
	quitting    bool
	showHelp    bool
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
func (m AppModel) Quitting() bool      { return m.quitting }

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

	case pages.RequestQuitMsg:
		if m.quitting {
			return m, tea.Quit
		}
		m.quitting = true
		return m, nil

	case tea.KeyMsg:
		// Any key while help is shown closes the overlay.
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		// Any key while quit confirmation is pending cancels (q is handled above).
		if m.quitting {
			m.quitting = false
			return m, nil
		}

		switch {
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit

		case msg.Type == tea.KeyTab:
			if m.currentPage == "issues" {
				m.currentPage = "actions"
				return m, m.actions.Init()
			}
			m.currentPage = "issues"
			return m, nil

		case msg.String() == "?":
			m.showHelp = true
			return m, nil
		}
	}

	// Delegate to active page.
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
	var base string
	switch m.currentPage {
	case "actions":
		base = m.actions.View()
	default:
		base = m.issues.View()
	}

	if m.showHelp {
		return m.helpOverlay()
	}
	if m.quitting {
		return m.quitOverlay()
	}
	return base
}

func (m AppModel) quitOverlay() string {
	panel := m.th.Panel.Render(
		m.th.Warning.Render("Quit ght?") + "\n\n" +
			m.th.Text.Render("Press ") + m.th.Accent.Render("q") + m.th.Text.Render(" again to confirm") + "\n" +
			m.th.Muted.Render("any other key to cancel"),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}

func (m AppModel) helpOverlay() string {
	content := m.th.TitleBar.Render("Keybindings") + "\n\n" +
		m.th.Accent.Render("Global") + "\n" +
		m.th.Muted.Render("  Tab       switch page (Issues ↔ Actions)\n") +
		m.th.Muted.Render("  q         quit (press again to confirm)\n") +
		m.th.Muted.Render("  ?         toggle this help\n") +
		m.th.Muted.Render("  Ctrl+C    force quit\n\n") +
		m.th.Accent.Render("Issues") + "\n" +
		m.th.Muted.Render("  [ / ]     cycle focus panel\n") +
		m.th.Muted.Render("  Enter     search (when filter focused)\n") +
		m.th.Muted.Render("  r         refresh\n") +
		m.th.Muted.Render("  o         open issue in browser\n") +
		m.th.Muted.Render("  f         load next page of results\n") +
		m.th.Muted.Render("  n         new issue (coming soon)\n\n") +
		m.th.Accent.Render("Actions") + "\n" +
		m.th.Muted.Render("  s         cycle status filter\n") +
		m.th.Muted.Render("  r         refresh\n") +
		m.th.Muted.Render("  o         open in browser\n") +
		m.th.Muted.Render("  Enter     drill in (runs → jobs → log)\n") +
		m.th.Muted.Render("  Esc       go back\n\n") +
		m.th.Muted.Render("Press any key to close")
	panel := m.th.Panel.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
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
