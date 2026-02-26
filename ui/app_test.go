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

func TestAppModel_TabSwitchesToActions(t *testing.T) {
	m := ui.NewApp()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m2 := updated.(ui.AppModel)
	if m2.CurrentPage() != "actions" {
		t.Errorf("expected page 'actions' after Tab, got %s", m2.CurrentPage())
	}
}

func TestAppModel_TabWrapsBackToIssues(t *testing.T) {
	m := ui.NewApp()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m2 := updated.(ui.AppModel)
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
	m3 := updated2.(ui.AppModel)
	if m3.CurrentPage() != "issues" {
		t.Errorf("expected page 'issues' after second Tab, got %s", m3.CurrentPage())
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

func TestAppModel_QuitConfirmation_FirstPressArms(t *testing.T) {
	m := ui.NewApp()
	updated, cmd := m.Update(pages.RequestQuitMsg{})
	m2 := updated.(ui.AppModel)
	if cmd != nil {
		t.Error("first RequestQuitMsg should not produce a command")
	}
	if !m2.Quitting() {
		t.Error("expected quitting=true after first RequestQuitMsg")
	}
}

func TestAppModel_QuitConfirmation_SecondPressQuits(t *testing.T) {
	m := ui.NewApp()
	updated, _ := m.Update(pages.RequestQuitMsg{})
	m2 := updated.(ui.AppModel)
	_, cmd2 := m2.Update(pages.RequestQuitMsg{})
	if cmd2 == nil {
		t.Error("second RequestQuitMsg should return tea.Quit command")
	}
}

func TestAppModel_QuitConfirmation_OtherKeyCancels(t *testing.T) {
	m := ui.NewApp()
	updated, _ := m.Update(pages.RequestQuitMsg{})
	m2 := updated.(ui.AppModel)
	// Press any non-q key — should cancel quitting
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m3 := updated2.(ui.AppModel)
	if m3.Quitting() {
		t.Error("expected quitting=false after pressing a non-q key")
	}
}
