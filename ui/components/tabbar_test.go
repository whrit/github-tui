package components_test

import (
	"strings"
	"testing"

	"github.com/skanehira/ght/ui/components"
	"github.com/skanehira/ght/ui/theme"
)

func TestTabBarRendersActivePage(t *testing.T) {
	th := theme.Default()
	tb := components.NewTabBar(th)

	view := tb.View("issues")
	if !strings.Contains(view, "Issues") {
		t.Error("tab bar must contain 'Issues'")
	}
	if !strings.Contains(view, "Actions") {
		t.Error("tab bar must contain 'Actions'")
	}

	view2 := tb.View("actions")
	if !strings.Contains(view2, "Actions") {
		t.Error("tab bar must contain 'Actions' when actions is active")
	}
}

func TestTabBarActiveIndicator(t *testing.T) {
	th := theme.Default()
	tb := components.NewTabBar(th)

	// Both pages should always be present
	for _, page := range []string{"issues", "actions"} {
		view := tb.View(page)
		if !strings.Contains(view, "Issues") || !strings.Contains(view, "Actions") {
			t.Errorf("View(%q): must always contain both tab names", page)
		}
	}
}
