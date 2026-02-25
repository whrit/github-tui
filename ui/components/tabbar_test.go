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

func TestTabBarActiveIndicatorPlacement(t *testing.T) {
	th := theme.Default()
	tb := components.NewTabBar(th)

	view := tb.View("issues")
	issuesIdx := strings.Index(view, "Issues")
	filledIdx := strings.Index(view, "●")
	if filledIdx < 0 || filledIdx > issuesIdx {
		t.Error("active indicator must appear before 'Issues' when page is issues")
	}

	view2 := tb.View("actions")
	actionsIdx := strings.Index(view2, "Actions")
	filledIdx2 := strings.Index(view2, "●")
	if filledIdx2 < 0 || filledIdx2 > actionsIdx {
		t.Error("active indicator must appear before 'Actions' when page is actions")
	}
}
