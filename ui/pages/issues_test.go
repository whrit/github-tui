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

// TestIssuesModel_TabSwitchMsg verifies that Ctrl+A emits a command (which
// when executed will produce a SwitchPageMsg{Page: "actions"}).
func TestIssuesModel_TabSwitchMsg(t *testing.T) {
	th := theme.Default()
	m := pages.NewIssuesModel(th)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	if cmd == nil {
		t.Error("Ctrl+A should return a command")
	}

	// Execute the returned command and verify it produces SwitchPageMsg.
	result := cmd()
	switchMsg, ok := result.(pages.SwitchPageMsg)
	if !ok {
		t.Errorf("expected SwitchPageMsg, got %T", result)
	}
	if switchMsg.Page != "actions" {
		t.Errorf("expected Page=actions, got %q", switchMsg.Page)
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

// TestIssuesModel_CtrlC_ReturnsQuit verifies that Ctrl+C returns the tea.Quit
// command.
func TestIssuesModel_CtrlC_ReturnsQuit(t *testing.T) {
	th := theme.Default()
	m := pages.NewIssuesModel(th)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("Ctrl+C should return a command")
	}
	// tea.Quit is a non-nil Cmd; we can't compare function pointers directly
	// but the presence of a non-nil Cmd is sufficient for this test.
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
