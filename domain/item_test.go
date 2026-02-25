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

func TestWorkflowRunStatusDisplay(t *testing.T) {
	cases := []struct {
		status     string
		conclusion string
		wantText   string
		wantRole   domain.ColorRole
	}{
		{"completed", "success", "success", domain.ColorRoleSuccess},
		{"completed", "failure", "failure", domain.ColorRoleDanger},
		{"completed", "cancelled", "cancelled", domain.ColorRoleMuted},
		{"in_progress", "", "in_progress", domain.ColorRoleWarning},
		{"queued", "", "queued", domain.ColorRoleMuted},
	}
	for _, c := range cases {
		run := domain.WorkflowRun{Status: c.status, Conclusion: c.conclusion}
		fields := run.Fields()
		if fields[0].Text != c.wantText {
			t.Errorf("status=%s conclusion=%s: got text %q, want %q",
				c.status, c.conclusion, fields[0].Text, c.wantText)
		}
		if fields[0].ColorRole != c.wantRole {
			t.Errorf("status=%s conclusion=%s: got role %q, want %q",
				c.status, c.conclusion, fields[0].ColorRole, c.wantRole)
		}
	}
}

func TestWorkflowJobStatusDisplay(t *testing.T) {
	cases := []struct {
		status     string
		conclusion string
		wantText   string
		wantRole   domain.ColorRole
	}{
		{"completed", "success", "success", domain.ColorRoleSuccess},
		{"completed", "failure", "failure", domain.ColorRoleDanger},
		{"completed", "cancelled", "cancelled", domain.ColorRoleMuted},
		{"in_progress", "", "in_progress", domain.ColorRoleWarning},
		{"queued", "", "queued", domain.ColorRoleMuted},
	}
	for _, c := range cases {
		job := domain.WorkflowJob{Status: c.status, Conclusion: c.conclusion}
		fields := job.Fields()
		if fields[0].Text != c.wantText {
			t.Errorf("status=%s conclusion=%s: got text %q, want %q",
				c.status, c.conclusion, fields[0].Text, c.wantText)
		}
		if fields[0].ColorRole != c.wantRole {
			t.Errorf("status=%s conclusion=%s: got role %q, want %q",
				c.status, c.conclusion, fields[0].ColorRole, c.wantRole)
		}
	}
}
