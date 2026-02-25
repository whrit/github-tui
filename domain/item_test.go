package domain_test

import (
	"testing"

	"github.com/skanehira/ght/domain"
)

func TestColorRoleConstants(t *testing.T) {
	roles := []domain.ColorRole{
		domain.ColorRoleDefault,
		domain.ColorRoleMuted,
		domain.ColorRoleAccent,
		domain.ColorRoleSuccess,
		domain.ColorRoleDanger,
		domain.ColorRoleWarning,
	}
	for _, r := range roles {
		if string(r) == "" {
			t.Errorf("ColorRole constant must not be empty string")
		}
	}
}

func TestFieldHasColorRole(t *testing.T) {
	f := domain.Field{Text: "hello", ColorRole: domain.ColorRoleSuccess}
	if f.ColorRole != domain.ColorRoleSuccess {
		t.Errorf("expected ColorRoleSuccess, got %s", f.ColorRole)
	}
}
