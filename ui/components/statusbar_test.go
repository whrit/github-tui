package components_test

import (
	"strings"
	"testing"

	"github.com/skanehira/ght/ui/components"
	"github.com/skanehira/ght/ui/theme"
)

func TestStatusBarContainsKeysAndContext(t *testing.T) {
	th := theme.Default()
	sb := components.NewStatusBar(th)
	view := sb.View(80, []components.KeyHint{
		{Key: "n", Desc: "new"},
		{Key: "/", Desc: "search"},
	}, "12 open")
	if !strings.Contains(view, "n") {
		t.Error("status bar must contain key hint 'n'")
	}
	if !strings.Contains(view, "12 open") {
		t.Error("status bar must contain context text '12 open'")
	}
}

func TestStatusBarEmptyHints(t *testing.T) {
	th := theme.Default()
	sb := components.NewStatusBar(th)
	view := sb.View(80, nil, "ready")
	if !strings.Contains(view, "ready") {
		t.Error("status bar must render context even with no key hints")
	}
}
