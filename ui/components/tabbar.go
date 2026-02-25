package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/skanehira/ght/ui/theme"
)

// TabBar renders the "● Issues  ○ Actions" persistent header.
type TabBar struct {
	th theme.Theme
}

func NewTabBar(th theme.Theme) TabBar {
	return TabBar{th: th}
}

// View renders the tab bar. activePage is "issues" or "actions".
func (tb TabBar) View(activePage string) string {
	const activeIndicator   = "● "
	const inactiveIndicator = "○ "

	issuesIndicator  := inactiveIndicator
	actionsIndicator := inactiveIndicator
	issuesStyle      := tb.th.TabInactive
	actionsStyle     := tb.th.TabInactive

	switch activePage {
	case "issues":
		issuesIndicator = activeIndicator
		issuesStyle     = tb.th.TabActive
	case "actions":
		actionsIndicator = activeIndicator
		actionsStyle     = tb.th.TabActive
	}

	issuesTab  := issuesStyle.Render(issuesIndicator + "Issues")
	actionsTab := actionsStyle.Render(actionsIndicator + "Actions")

	return lipgloss.JoinHorizontal(lipgloss.Top, issuesTab, "    ", actionsTab)
}
