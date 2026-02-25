package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/skanehira/ght/ui/theme"
)

// KeyHint pairs a key label with a description for the status bar.
type KeyHint struct {
	Key  string
	Desc string
}

// StatusBar renders the bottom context bar.
type StatusBar struct {
	th theme.Theme
}

func NewStatusBar(th theme.Theme) StatusBar {
	return StatusBar{th: th}
}

// View renders the status bar at the given terminal width.
// Key hints are left-aligned; context string is right-aligned.
func (sb StatusBar) View(width int, hints []KeyHint, context string) string {
	keyStyle  := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.Accent))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.TextMuted))

	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, keyStyle.Render(h.Key)+":"+descStyle.Render(h.Desc))
	}
	left := strings.Join(parts, "  ")

	contextStyled := descStyle.Render(context)

	// Calculate visible widths (lipgloss.Width strips ANSI codes).
	leftWidth    := lipgloss.Width(left)
	contextWidth := lipgloss.Width(contextStyled)
	// Account for the StatusBar style's left/right padding (1 each).
	const sidePadding = 2
	gap := width - sidePadding - leftWidth - contextWidth
	if gap < 1 {
		gap = 1
	}
	line := left + fmt.Sprintf("%*s", gap, "") + contextStyled

	return sb.th.StatusBar.Width(width).Render(line)
}
