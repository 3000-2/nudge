package output_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
	"github.com/3000-2/nudge/internal/output"
)

func TestPrintReminderList(t *testing.T) {
	date := "2026-04-02"
	reminders := []model.Reminder{
		{
			ID:      "abcd1234",
			Message: "Pay rent",
			Schedule: model.Schedule{
				Type:   model.ScheduleTypeOnce,
				Hour:   9,
				Minute: 30,
				Date:   &date,
			},
			Status:    model.StatusActive,
			CreatedAt: time.Date(2026, time.April, 1, 12, 0, 0, 0, time.Local),
		},
	}

	var buf bytes.Buffer
	if err := output.PrintReminderList(&buf, reminders); err != nil {
		t.Fatalf("PrintReminderList returned error: %v", err)
	}

	text := buf.String()
	for _, part := range []string{"ID", "STATUS", "WHEN", "MESSAGE", "abcd1234", "Pay rent"} {
		if !strings.Contains(text, part) {
			t.Fatalf("expected output to contain %q\n%s", part, text)
		}
	}
}

func TestPrintReminderDetailAndJSON(t *testing.T) {
	date := "2026-04-02"
	reminder := model.Reminder{
		ID:      "abcd1234",
		Message: "Pay rent",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   9,
			Minute: 30,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: time.Date(2026, time.April, 1, 12, 0, 0, 0, time.Local),
	}

	var detail bytes.Buffer
	if err := output.PrintReminderDetail(&detail, reminder); err != nil {
		t.Fatalf("PrintReminderDetail returned error: %v", err)
	}
	if !strings.Contains(detail.String(), "Message") || !strings.Contains(detail.String(), "Schedule") {
		t.Fatalf("unexpected detail output:\n%s", detail.String())
	}

	var jsonBuf bytes.Buffer
	if err := output.PrintJSON(&jsonBuf, reminder); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}
	if !strings.Contains(jsonBuf.String(), `"id": "abcd1234"`) {
		t.Fatalf("unexpected json output:\n%s", jsonBuf.String())
	}
}
