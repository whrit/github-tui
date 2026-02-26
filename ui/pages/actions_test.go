package pages_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/skanehira/ght/ui/pages"
	"github.com/skanehira/ght/ui/theme"
)

// TestActionsModel_QEmitsRequestQuit verifies that q emits a RequestQuitMsg command.
func TestActionsModel_QEmitsRequestQuit(t *testing.T) {
	th := theme.Default()
	m := pages.NewActionsModel(th)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("q should emit RequestQuitMsg")
	}
	result := cmd()
	if _, ok := result.(pages.RequestQuitMsg); !ok {
		t.Errorf("expected RequestQuitMsg, got %T", result)
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
