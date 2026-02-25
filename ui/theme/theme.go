package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/skanehira/ght/domain"
)

// Palette holds the raw hex color values for the GitHub dark theme.
var Palette = struct {
	Background  string
	Surface     string
	Border      string
	BorderFocus string
	Text        string
	TextMuted   string
	Success     string
	Danger      string
	Warning     string
	Accent      string
}{
	Background:  "#0d1117",
	Surface:     "#161b22",
	Border:      "#30363d",
	BorderFocus: "#58a6ff",
	Text:        "#e6edf3",
	TextMuted:   "#7d8590",
	Success:     "#3fb950",
	Danger:      "#f85149",
	Warning:     "#d29922",
	Accent:      "#58a6ff",
}

// Theme holds pre-built lipgloss styles.
type Theme struct {
	Text    lipgloss.Style
	Muted   lipgloss.Style
	Accent  lipgloss.Style
	Success lipgloss.Style
	Danger  lipgloss.Style
	Warning lipgloss.Style

	Panel        lipgloss.Style
	PanelFocused lipgloss.Style

	TableHeader   lipgloss.Style
	TableSelected lipgloss.Style

	TabActive   lipgloss.Style
	TabInactive lipgloss.Style

	StatusBar lipgloss.Style
	TitleBar  lipgloss.Style
}

// Default returns the canonical dark GitHub-palette theme.
func Default() Theme {
	p := Palette
	return Theme{
		Text:    lipgloss.NewStyle().Foreground(lipgloss.Color(p.Text)),
		Muted:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.TextMuted)),
		Accent:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.Accent)),
		Success: lipgloss.NewStyle().Foreground(lipgloss.Color(p.Success)),
		Danger:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.Danger)),
		Warning: lipgloss.NewStyle().Foreground(lipgloss.Color(p.Warning)),

		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(p.Border)),

		PanelFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(p.BorderFocus)),

		TableHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)).
			Bold(true),

		TableSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Text)).
			Background(lipgloss.Color("#1f2937")),

		TabActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Accent)).
			Bold(true),

		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.TextMuted)).
			Background(lipgloss.Color(p.Surface)).
			PaddingLeft(1).PaddingRight(1),

		TitleBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Text)).
			Background(lipgloss.Color(p.Surface)).
			Bold(true).
			PaddingLeft(1).PaddingRight(1),
	}
}

// Resolve maps a domain ColorRole to a lipgloss.Color.
func (th Theme) Resolve(role domain.ColorRole) lipgloss.Color {
	switch role {
	case domain.ColorRoleSuccess:
		return lipgloss.Color(Palette.Success)
	case domain.ColorRoleDanger:
		return lipgloss.Color(Palette.Danger)
	case domain.ColorRoleWarning:
		return lipgloss.Color(Palette.Warning)
	case domain.ColorRoleAccent:
		return lipgloss.Color(Palette.Accent)
	case domain.ColorRoleMuted:
		return lipgloss.Color(Palette.TextMuted)
	default:
		return lipgloss.Color(Palette.Text)
	}
}

// StyleField returns text styled for its ColorRole.
func (th Theme) StyleField(text string, role domain.ColorRole) string {
	return lipgloss.NewStyle().Foreground(th.Resolve(role)).Render(text)
}
