package theme_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/skanehira/ght/domain"
	"github.com/skanehira/ght/ui/theme"
)

func TestThemeResolveColorRole(t *testing.T) {
	th := theme.Default()
	cases := []struct {
		role domain.ColorRole
	}{
		{domain.ColorRoleDefault},
		{domain.ColorRoleMuted},
		{domain.ColorRoleAccent},
		{domain.ColorRoleSuccess},
		{domain.ColorRoleDanger},
		{domain.ColorRoleWarning},
	}
	for _, c := range cases {
		color := th.Resolve(c.role)
		if color == (lipgloss.Color("")) {
			t.Errorf("Resolve(%s) returned empty color", c.role)
		}
	}
}

func TestThemeStylesNonZero(t *testing.T) {
	th := theme.Default()
	if th.Panel.GetBorderStyle() == (lipgloss.Border{}) {
		t.Error("Panel style must have a border")
	}
}
