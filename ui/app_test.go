package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/skanehira/ght/ui"
	"github.com/skanehira/ght/ui/pages"
)

func TestAppModel_DefaultPageIsIssues(t *testing.T) {
	m := ui.NewApp()
	if m.CurrentPage() != "issues" {
		t.Errorf("expected default page 'issues', got %s", m.CurrentPage())
	}
}

func TestAppModel_SwitchPageMsg(t *testing.T) {
	m := ui.NewApp()
	updated, _ := m.Update(pages.SwitchPageMsg{Page: "actions"})
	m2 := updated.(ui.AppModel)
	if m2.CurrentPage() != "actions" {
		t.Errorf("expected page 'actions', got %s", m2.CurrentPage())
	}
}
