package domain

import "fmt"

// WorkflowJob represents a GitHub Actions workflow job.
type WorkflowJob struct {
	ID         int64
	Name       string
	Status     string
	Conclusion string
	Duration   string
	HTMLURL    string
	RunID      int64
}

func (j *WorkflowJob) Key() string {
	return fmt.Sprintf("%d", j.ID)
}

func (j *WorkflowJob) Fields() []Field {
	statusText, role := statusDisplay(j.Status, j.Conclusion)

	return []Field{
		{Text: statusText, ColorRole: role},
		{Text: j.Name, ColorRole: ColorRoleDefault},
		{Text: j.Duration, ColorRole: ColorRoleDefault},
	}
}
