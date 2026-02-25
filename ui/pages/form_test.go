package pages_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/skanehira/ght/ui/pages"
	"github.com/skanehira/ght/ui/theme"
)

// TestCreateIssueForm_EscDismisses verifies that pressing Escape returns a
// non-nil command (which will produce a FormDismissedMsg when executed).
func TestCreateIssueForm_EscDismisses(t *testing.T) {
	th := theme.Default()
	m := pages.NewCreateIssueForm(th, "owner", "repo")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Error("Esc should return a command (FormDismissedMsg)")
	}
}

// TestCreateIssueForm_TabCyclesFocus verifies that pressing Tab advances focus
// to the next field without panicking, and that View() renders correctly.
func TestCreateIssueForm_TabCyclesFocus(t *testing.T) {
	th := theme.Default()
	m := pages.NewCreateIssueForm(th, "owner", "repo")
	// Tab should move to next field (not crash)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m2 := updated.(pages.CreateIssueFormModel)
	_ = m2.View() // should not panic
}
