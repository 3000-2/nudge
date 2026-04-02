package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
)

func TestShowExistingReminder(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))
	date := "2026-04-02"
	env.seedReminder(model.Reminder{
		ID:      "show0001",
		Message: "Stretch",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   12,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: env.now,
	})

	stdout, _, err := executeCommandResult(t, "test", "show", "show0001")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	for _, want := range []string{"show0001", "Stretch", "once on 2026-04-02 at 12:00"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected output to contain %q\n%s", want, stdout)
		}
	}
}

func TestShowNonExistentReminder(t *testing.T) {
	setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))

	_, _, err := executeCommandResult(t, "test", "show", "miss0000")
	if err == nil || !strings.Contains(err.Error(), `reminder "miss0000" not found`) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestShowJSONOutput(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))
	date := "2026-04-02"
	env.seedReminder(model.Reminder{
		ID:      "showjson",
		Message: "Read book",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   18,
			Minute: 30,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: env.now,
	})

	stdout, _, err := executeCommandResult(t, "test", "show", "showjson", "--json")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var reminder model.Reminder
	if err := json.Unmarshal([]byte(stdout), &reminder); err != nil {
		t.Fatalf("failed to unmarshal show json output: %v\n%s", err, stdout)
	}
	if reminder.ID != "showjson" || reminder.Message != "Read book" {
		t.Fatalf("unexpected reminder payload %#v", reminder)
	}
}
