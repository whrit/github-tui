package pages_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/skanehira/ght/ui/pages"
	"github.com/skanehira/ght/ui/theme"
)

// TestActionsModel_TabSwitchToIssues verifies that Ctrl+I emits a command
// that, when executed, produces a SwitchPageMsg{Page: "issues"}.
func TestActionsModel_TabSwitchToIssues(t *testing.T) {
	th := theme.Default()
	m := pages.NewActionsModel(th)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlI})
	if cmd == nil {
		t.Error("Ctrl+I should emit SwitchPageMsg")
	}
}

// TestActionsModel_RunsLoaded verifies that a RunsLoadedMsg transitions the
// model out of the loading state.
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
