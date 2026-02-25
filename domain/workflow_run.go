package domain

import "fmt"

// WorkflowRun represents a GitHub Actions workflow run.
type WorkflowRun struct {
	ID         int64
	Name       string
	Title      string
	Status     string
	Conclusion string
	HeadBranch string
	Event      string
	RunNumber  int
	CreatedAt  string
	Duration   string
	HTMLURL    string
	RunID      int64
}

func (w *WorkflowRun) Key() string {
	return fmt.Sprintf("%d", w.ID)
}

func (w *WorkflowRun) Fields() []Field {
	statusText, role := statusDisplay(w.Status, w.Conclusion)

	return []Field{
		{Text: statusText, ColorRole: role},
		{Text: w.Name, ColorRole: ColorRoleDefault},
		{Text: w.HeadBranch, ColorRole: ColorRoleAccent},
		{Text: w.Event, ColorRole: ColorRoleWarning},
		{Text: w.Duration, ColorRole: ColorRoleDefault},
	}
}

// statusDisplay returns the display text and ColorRole for a workflow status/conclusion pair.
// For completed runs, the conclusion text is displayed; for non-completed runs, the status text.
func statusDisplay(status, conclusion string) (string, ColorRole) {
	switch status {
	case "completed":
		switch conclusion {
		case "success":
			return conclusion, ColorRoleSuccess
		case "failure":
			return conclusion, ColorRoleDanger
		default:
			// cancelled, skipped, etc.
			return conclusion, ColorRoleMuted
		}
	case "in_progress":
		return status, ColorRoleWarning
	default:
		// queued, waiting, etc.
		return status, ColorRoleMuted
	}
}
