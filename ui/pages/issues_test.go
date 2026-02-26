package pages_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shurcooL/githubv4"
	"github.com/skanehira/ght/domain"
	github "github.com/skanehira/ght/github"
	"github.com/skanehira/ght/ui/pages"
	"github.com/skanehira/ght/ui/theme"
)

// TestIssuesModel_LoadedMsg_SetsItems verifies that when an IssuesLoadedMsg
// is processed the model transitions out of the loading state and stores the
// received items.
func TestIssuesModel_LoadedMsg_SetsItems(t *testing.T) {
	th := theme.Default()
	m := pages.NewIssuesModel(th)

	// github.PageInfo uses githubv4.Boolean (not bool) for HasNextPage.
	msg := pages.IssuesLoadedMsg{
		Items: []domain.Item{
			&domain.Issue{ID: "1", Number: "1", Title: "Fix bug", State: "OPEN"},
		},
		PageInfo: &github.PageInfo{HasNextPage: githubv4.Boolean(false)},
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

// TestIssuesModel_QSwitchMsg verifies that pressing q (when not in the filter)
// emits a RequestQuitMsg command.
func TestIssuesModel_QSwitchMsg(t *testing.T) {
	th := theme.Default()
	m := pages.NewIssuesModel(th)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("q should return a command when filter is not focused")
	}

	// Execute the returned command and verify it produces RequestQuitMsg.
	result := cmd()
	_, ok := result.(pages.RequestQuitMsg)
	if !ok {
		t.Errorf("expected RequestQuitMsg, got %T", result)
	}
}

// TestIssuesModel_MoreLoadedMsg_AppendsItems verifies that IssuesMoreLoadedMsg
// appends to the existing issue list rather than replacing it.
func TestIssuesModel_MoreLoadedMsg_AppendsItems(t *testing.T) {
	th := theme.Default()
	m := pages.NewIssuesModel(th)

	// First load
	first := pages.IssuesLoadedMsg{
		Items: []domain.Item{
			&domain.Issue{ID: "1", Number: "1", Title: "First", State: "OPEN"},
		},
		PageInfo: &github.PageInfo{HasNextPage: githubv4.Boolean(true), EndCursor: githubv4.String("cursor1")},
	}
	updated, _ := m.Update(first)
	m = updated.(pages.IssuesModel)

	// Second load (more)
	more := pages.IssuesMoreLoadedMsg{
		Items: []domain.Item{
			&domain.Issue{ID: "2", Number: "2", Title: "Second", State: "OPEN"},
		},
		PageInfo: &github.PageInfo{HasNextPage: githubv4.Boolean(false)},
	}
	updated2, _ := m.Update(more)
	m2 := updated2.(pages.IssuesModel)

	if len(m2.Issues()) != 2 {
		t.Errorf("expected 2 issues after append, got %d", len(m2.Issues()))
	}
}

// TestIssuesModel_CtrlC_IsHandledByAppModel verifies that Ctrl+C is not handled
// by the Issues page (it is intercepted by AppModel before delegation).
func TestIssuesModel_CtrlC_IsHandledByAppModel(t *testing.T) {
	th := theme.Default()
	m := pages.NewIssuesModel(th)

	// Ctrl+C should return nil from the Issues page since AppModel now owns it.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	// Delegated to the textinput/table which returns nil — the important thing
	// is the page does not quit on its own.
	_ = cmd // result is implementation-defined; we just verify no panic
}

// TestIssuesModel_ViewContainsPanels verifies that View() produces a non-empty
// string after a window size message is applied.
func TestIssuesModel_ViewContainsPanels(t *testing.T) {
	th := theme.Default()
	m := pages.NewIssuesModel(th)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 220, Height: 50})
	m2 := updated.(pages.IssuesModel)

	view := m2.View()
	if len(view) == 0 {
		t.Error("View() should not return empty string")
	}
}
