package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
)

func TestListActiveRemindersOnlyByDefault(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))
	date := "2026-04-02"

	env.seedReminder(model.Reminder{
		ID:      "active01",
		Message: "Pay rent",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   9,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: env.now,
	})
	env.seedReminder(model.Reminder{
		ID:      "done0001",
		Message: "Done task",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   10,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusCompleted,
		CreatedAt: env.now,
	})

	stdout, _, err := executeCommandResult(t, "test", "list")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(stdout, "active01") || strings.Contains(stdout, "done0001") {
		t.Fatalf("unexpected list output:\n%s", stdout)
	}
}

func TestListAllIncludesCompleted(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))
	date := "2026-04-02"

	for _, reminder := range []model.Reminder{
		{
			ID:      "active01",
			Message: "Pay rent",
			Schedule: model.Schedule{
				Type:   model.ScheduleTypeOnce,
				Hour:   9,
				Minute: 0,
				Date:   &date,
			},
			Status:    model.StatusActive,
			CreatedAt: env.now,
		},
		{
			ID:      "done0001",
			Message: "Done task",
			Schedule: model.Schedule{
				Type:   model.ScheduleTypeOnce,
				Hour:   10,
				Minute: 0,
				Date:   &date,
			},
			Status:    model.StatusCompleted,
			CreatedAt: env.now,
		},
	} {
		env.seedReminder(reminder)
	}

	stdout, _, err := executeCommandResult(t, "test", "list", "--all")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	for _, want := range []string{"active01", "done0001"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected output to contain %q\n%s", want, stdout)
		}
	}
}

func TestListEmpty(t *testing.T) {
	setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))

	stdout, _, err := executeCommandResult(t, "test", "list")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if strings.TrimSpace(stdout) != "No reminders." {
		t.Fatalf("expected empty list output, got %q", stdout)
	}
}

func TestListJSONOutput(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))
	date := "2026-04-02"

	env.seedReminder(model.Reminder{
		ID:      "active01",
		Message: "Pay rent",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   9,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: env.now,
	})

	stdout, _, err := executeCommandResult(t, "test", "list", "--json")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var reminders []model.Reminder
	if err := json.Unmarshal([]byte(stdout), &reminders); err != nil {
		t.Fatalf("failed to unmarshal list json output: %v\n%s", err, stdout)
	}
	if len(reminders) != 1 || reminders[0].ID != "active01" {
		t.Fatalf("unexpected reminders payload %#v", reminders)
	}
}
