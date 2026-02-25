package github

import (
	"testing"
	"time"

	gogithub "github.com/google/go-github/v68/github"

	"github.com/skanehira/ght/domain"
)

func TestCleanLog(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strips ANSI escape codes",
			input: "\x1b[32mPassing\x1b[0m test \x1b[1;31mfailed\x1b[0m",
			want:  "Passing test failed",
		},
		{
			name:  "strips GitHub timestamp prefixes",
			input: "2024-01-15T10:30:45.1234567Z Run started\n2024-01-15T10:30:46.9876543Z Step completed",
			want:  "Run started\nStep completed",
		},
		{
			name:  "strips both ANSI and timestamps",
			input: "2024-01-15T10:30:45.1234567Z \x1b[32mPassing\x1b[0m test",
			want:  "Passing test",
		},
		{
			name:  "no special chars unchanged",
			input: "Hello world\nSecond line",
			want:  "Hello world\nSecond line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanLog(tt.input)
			if got != tt.want {
				t.Errorf("CleanLog() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{
			name: "seconds only",
			d:    42 * time.Second,
			want: "42s",
		},
		{
			name: "minutes and seconds",
			d:    3*time.Minute + 15*time.Second,
			want: "3m 15s",
		},
		{
			name: "hours and minutes",
			d:    2*time.Hour + 30*time.Minute,
			want: "2h 30m",
		},
		{
			name: "zero",
			d:    0,
			want: "0s",
		},
		{
			name: "exactly one minute",
			d:    1 * time.Minute,
			want: "1m 0s",
		},
		{
			name: "exactly one hour",
			d:    1 * time.Hour,
			want: "1h 0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertWorkflowRun(t *testing.T) {
	now := time.Now()
	startedAt := gogithub.Timestamp{Time: now.Add(-5 * time.Minute)}
	updatedAt := gogithub.Timestamp{Time: now}
	createdAt := gogithub.Timestamp{Time: now}

	tests := []struct {
		name       string
		run        *gogithub.WorkflowRun
		wantName   string
		wantStatus string
		wantConc   string
		wantBranch string
		wantEvent  string
	}{
		{
			name: "completed success run",
			run: &gogithub.WorkflowRun{
				ID:           gogithub.Ptr(int64(12345)),
				Name:         gogithub.Ptr("CI"),
				DisplayTitle: gogithub.Ptr("fix: update deps"),
				Status:       gogithub.Ptr("completed"),
				Conclusion:   gogithub.Ptr("success"),
				HeadBranch:   gogithub.Ptr("main"),
				Event:        gogithub.Ptr("push"),
				RunNumber:    gogithub.Ptr(42),
				RunStartedAt: &startedAt,
				UpdatedAt:    &updatedAt,
				CreatedAt:    &createdAt,
				HTMLURL:      gogithub.Ptr("https://github.com/org/repo/actions/runs/12345"),
			},
			wantName:   "CI",
			wantStatus: "completed",
			wantConc:   "success",
			wantBranch: "main",
			wantEvent:  "push",
		},
		{
			name: "in_progress run",
			run: &gogithub.WorkflowRun{
				ID:           gogithub.Ptr(int64(12346)),
				Name:         gogithub.Ptr("Deploy"),
				DisplayTitle: gogithub.Ptr("feat: new feature"),
				Status:       gogithub.Ptr("in_progress"),
				Conclusion:   gogithub.Ptr(""),
				HeadBranch:   gogithub.Ptr("feature-branch"),
				Event:        gogithub.Ptr("pull_request"),
				RunNumber:    gogithub.Ptr(43),
				RunStartedAt: &startedAt,
				UpdatedAt:    &updatedAt,
				CreatedAt:    &createdAt,
				HTMLURL:      gogithub.Ptr("https://github.com/org/repo/actions/runs/12346"),
			},
			wantName:   "Deploy",
			wantStatus: "in_progress",
			wantConc:   "",
			wantBranch: "feature-branch",
			wantEvent:  "pull_request",
		},
		{
			name: "completed failure run",
			run: &gogithub.WorkflowRun{
				ID:           gogithub.Ptr(int64(12347)),
				Name:         gogithub.Ptr("Tests"),
				DisplayTitle: gogithub.Ptr("test: add coverage"),
				Status:       gogithub.Ptr("completed"),
				Conclusion:   gogithub.Ptr("failure"),
				HeadBranch:   gogithub.Ptr("dev"),
				Event:        gogithub.Ptr("push"),
				RunNumber:    gogithub.Ptr(44),
				RunStartedAt: &startedAt,
				UpdatedAt:    &updatedAt,
				CreatedAt:    &createdAt,
				HTMLURL:      gogithub.Ptr("https://github.com/org/repo/actions/runs/12347"),
			},
			wantName:   "Tests",
			wantStatus: "completed",
			wantConc:   "failure",
			wantBranch: "dev",
			wantEvent:  "push",
		},
		{
			name: "nil RunStartedAt gives empty duration",
			run: &gogithub.WorkflowRun{
				ID:           gogithub.Ptr(int64(12348)),
				Name:         gogithub.Ptr("Build"),
				DisplayTitle: gogithub.Ptr("chore: bump version"),
				Status:       gogithub.Ptr("queued"),
				Conclusion:   gogithub.Ptr(""),
				HeadBranch:   gogithub.Ptr("main"),
				Event:        gogithub.Ptr("workflow_dispatch"),
				RunNumber:    gogithub.Ptr(45),
				RunStartedAt: nil,
				UpdatedAt:    &updatedAt,
				CreatedAt:    &createdAt,
				HTMLURL:      gogithub.Ptr("https://github.com/org/repo/actions/runs/12348"),
			},
			wantName:   "Build",
			wantStatus: "queued",
			wantConc:   "",
			wantBranch: "main",
			wantEvent:  "workflow_dispatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertWorkflowRun(tt.run)

			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.Conclusion != tt.wantConc {
				t.Errorf("Conclusion = %q, want %q", got.Conclusion, tt.wantConc)
			}
			if got.HeadBranch != tt.wantBranch {
				t.Errorf("HeadBranch = %q, want %q", got.HeadBranch, tt.wantBranch)
			}
			if got.Event != tt.wantEvent {
				t.Errorf("Event = %q, want %q", got.Event, tt.wantEvent)
			}
			if got.ID != tt.run.GetID() {
				t.Errorf("ID = %d, want %d", got.ID, tt.run.GetID())
			}

			// Verify nil RunStartedAt results in empty duration
			if tt.run.RunStartedAt == nil && got.Duration != "" {
				t.Errorf("Duration = %q, want empty for nil RunStartedAt", got.Duration)
			}

			// Verify non-nil RunStartedAt results in non-empty duration
			if tt.run.RunStartedAt != nil && got.Duration == "" {
				t.Errorf("Duration is empty, want non-empty for non-nil RunStartedAt")
			}
		})
	}
}

func TestConvertWorkflowJob(t *testing.T) {
	now := time.Now()
	startedAt := gogithub.Timestamp{Time: now.Add(-2 * time.Minute)}
	completedAt := gogithub.Timestamp{Time: now}

	tests := []struct {
		name       string
		job        *gogithub.WorkflowJob
		wantName   string
		wantStatus string
		wantConc   string
	}{
		{
			name: "completed success job",
			job: &gogithub.WorkflowJob{
				ID:          gogithub.Ptr(int64(99001)),
				Name:        gogithub.Ptr("build"),
				Status:      gogithub.Ptr("completed"),
				Conclusion:  gogithub.Ptr("success"),
				RunID:       gogithub.Ptr(int64(12345)),
				HTMLURL:     gogithub.Ptr("https://github.com/org/repo/actions/runs/12345/jobs/99001"),
				StartedAt:   &startedAt,
				CompletedAt: &completedAt,
			},
			wantName:   "build",
			wantStatus: "completed",
			wantConc:   "success",
		},
		{
			name: "in_progress job",
			job: &gogithub.WorkflowJob{
				ID:          gogithub.Ptr(int64(99002)),
				Name:        gogithub.Ptr("test"),
				Status:      gogithub.Ptr("in_progress"),
				Conclusion:  gogithub.Ptr(""),
				RunID:       gogithub.Ptr(int64(12345)),
				HTMLURL:     gogithub.Ptr("https://github.com/org/repo/actions/runs/12345/jobs/99002"),
				StartedAt:   &startedAt,
				CompletedAt: nil,
			},
			wantName:   "test",
			wantStatus: "in_progress",
			wantConc:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertWorkflowJob(tt.job)

			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.Conclusion != tt.wantConc {
				t.Errorf("Conclusion = %q, want %q", got.Conclusion, tt.wantConc)
			}
			if got.ID != tt.job.GetID() {
				t.Errorf("ID = %d, want %d", got.ID, tt.job.GetID())
			}
			if got.RunID != tt.job.GetRunID() {
				t.Errorf("RunID = %d, want %d", got.RunID, tt.job.GetRunID())
			}
		})
	}
}

func TestWorkflowRunFields(t *testing.T) {
	tests := []struct {
		name       string
		run        domain.WorkflowRun
		wantRole   domain.ColorRole
		wantStatus string
	}{
		{
			name: "completed success is green",
			run: domain.WorkflowRun{
				ID: 1, Status: "completed", Conclusion: "success",
				Name: "CI", HeadBranch: "main", Event: "push", Duration: "5m 0s",
			},
			wantRole:   domain.ColorRoleSuccess,
			wantStatus: "success",
		},
		{
			name: "completed failure is red",
			run: domain.WorkflowRun{
				ID: 2, Status: "completed", Conclusion: "failure",
				Name: "CI", HeadBranch: "main", Event: "push", Duration: "3m 15s",
			},
			wantRole:   domain.ColorRoleDanger,
			wantStatus: "failure",
		},
		{
			name: "completed cancelled is muted",
			run: domain.WorkflowRun{
				ID: 3, Status: "completed", Conclusion: "cancelled",
				Name: "CI", HeadBranch: "main", Event: "push", Duration: "1m 0s",
			},
			wantRole:   domain.ColorRoleMuted,
			wantStatus: "cancelled",
		},
		{
			name: "in_progress is warning",
			run: domain.WorkflowRun{
				ID: 4, Status: "in_progress", Conclusion: "",
				Name: "CI", HeadBranch: "main", Event: "push", Duration: "2m 30s",
			},
			wantRole:   domain.ColorRoleWarning,
			wantStatus: "in_progress",
		},
		{
			name: "queued is muted",
			run: domain.WorkflowRun{
				ID: 5, Status: "queued", Conclusion: "",
				Name: "CI", HeadBranch: "main", Event: "push", Duration: "",
			},
			wantRole:   domain.ColorRoleMuted,
			wantStatus: "queued",
		},
		{
			name: "waiting is muted",
			run: domain.WorkflowRun{
				ID: 6, Status: "waiting", Conclusion: "",
				Name: "CI", HeadBranch: "main", Event: "push", Duration: "",
			},
			wantRole:   domain.ColorRoleMuted,
			wantStatus: "waiting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := tt.run.Fields()

			if len(fields) != 5 {
				t.Fatalf("Fields() returned %d fields, want 5", len(fields))
			}

			// First field is status
			statusField := fields[0]
			if statusField.ColorRole != tt.wantRole {
				t.Errorf("status color role = %v, want %v", statusField.ColorRole, tt.wantRole)
			}
			if statusField.Text != tt.wantStatus {
				t.Errorf("status text = %q, want %q", statusField.Text, tt.wantStatus)
			}
		})
	}
}

func TestWorkflowJobFields(t *testing.T) {
	tests := []struct {
		name       string
		job        domain.WorkflowJob
		wantRole   domain.ColorRole
		wantStatus string
	}{
		{
			name: "completed success is success role",
			job: domain.WorkflowJob{
				ID: 1, Status: "completed", Conclusion: "success",
				Name: "build", Duration: "2m 0s",
			},
			wantRole:   domain.ColorRoleSuccess,
			wantStatus: "success",
		},
		{
			name: "completed failure is danger role",
			job: domain.WorkflowJob{
				ID: 2, Status: "completed", Conclusion: "failure",
				Name: "test", Duration: "1m 30s",
			},
			wantRole:   domain.ColorRoleDanger,
			wantStatus: "failure",
		},
		{
			name: "in_progress is warning role",
			job: domain.WorkflowJob{
				ID: 3, Status: "in_progress", Conclusion: "",
				Name: "deploy", Duration: "30s",
			},
			wantRole:   domain.ColorRoleWarning,
			wantStatus: "in_progress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := tt.job.Fields()

			if len(fields) != 3 {
				t.Fatalf("Fields() returned %d fields, want 3", len(fields))
			}

			// First field is status
			statusField := fields[0]
			if statusField.ColorRole != tt.wantRole {
				t.Errorf("status color role = %v, want %v", statusField.ColorRole, tt.wantRole)
			}
			if statusField.Text != tt.wantStatus {
				t.Errorf("status text = %q, want %q", statusField.Text, tt.wantStatus)
			}
		})
	}
}
