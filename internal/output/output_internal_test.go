package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
)

func TestOrdinal(t *testing.T) {
	tests := []struct {
		day  int
		want string
	}{
		{day: 1, want: "1st"},
		{day: 2, want: "2nd"},
		{day: 3, want: "3rd"},
		{day: 4, want: "4th"},
		{day: 11, want: "11th"},
		{day: 12, want: "12th"},
		{day: 13, want: "13th"},
		{day: 21, want: "21st"},
		{day: 22, want: "22nd"},
		{day: 23, want: "23rd"},
		{day: 31, want: "31st"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := ordinal(tt.day); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestWeekdayName(t *testing.T) {
	tests := []struct {
		value int
		want  string
	}{
		{value: 0, want: "sunday"},
		{value: 1, want: "monday"},
		{value: 2, want: "tuesday"},
		{value: 3, want: "wednesday"},
		{value: 4, want: "thursday"},
		{value: 5, want: "friday"},
		{value: 6, want: "saturday"},
		{value: 99, want: "weekday-99"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := weekdayName(tt.value); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestFormatScheduleCoversNilPointerCases(t *testing.T) {
	date := "2026-04-15"
	weekday := 1
	dayOfMonth := 2

	tests := []struct {
		name     string
		schedule model.Schedule
		want     string
	}{
		{
			name: "once with date",
			schedule: model.Schedule{
				Type:   model.ScheduleTypeOnce,
				Hour:   9,
				Minute: 30,
				Date:   &date,
			},
			want: "once on 2026-04-15 at 09:30",
		},
		{
			name: "once missing date",
			schedule: model.Schedule{
				Type:   model.ScheduleTypeOnce,
				Hour:   9,
				Minute: 30,
			},
			want: "once at 09:30",
		},
		{
			name: "daily",
			schedule: model.Schedule{
				Type:   model.ScheduleTypeDaily,
				Hour:   7,
				Minute: 15,
			},
			want: "daily at 07:15",
		},
		{
			name: "weekly with weekday",
			schedule: model.Schedule{
				Type:    model.ScheduleTypeWeekly,
				Hour:    8,
				Minute:  45,
				Weekday: &weekday,
			},
			want: "every monday at 08:45",
		},
		{
			name: "weekly missing weekday",
			schedule: model.Schedule{
				Type:   model.ScheduleTypeWeekly,
				Hour:   8,
				Minute: 45,
			},
			want: "weekly at 08:45",
		},
		{
			name: "monthly with day",
			schedule: model.Schedule{
				Type:       model.ScheduleTypeMonthly,
				Hour:       10,
				Minute:     5,
				DayOfMonth: &dayOfMonth,
			},
			want: "every 2nd at 10:05",
		},
		{
			name: "monthly missing day",
			schedule: model.Schedule{
				Type:   model.ScheduleTypeMonthly,
				Hour:   10,
				Minute: 5,
			},
			want: "monthly at 10:05",
		},
		{
			name: "unknown type",
			schedule: model.Schedule{
				Type:   "custom",
				Hour:   22,
				Minute: 10,
			},
			want: "custom at 22:10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatSchedule(tt.schedule); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestFormatOptionalTimestamp(t *testing.T) {
	if got := formatOptionalTimestamp(nil); got != "-" {
		t.Fatalf("expected nil timestamp to format as -, got %q", got)
	}

	value := time.Date(2026, time.April, 1, 12, 34, 56, 0, time.Local)
	if got := formatOptionalTimestamp(&value); got != formatTimestamp(value) {
		t.Fatalf("expected %q, got %q", formatTimestamp(value), got)
	}
}

func TestPrintReminderListMultipleReminders(t *testing.T) {
	date := "2026-04-15"
	dayOfMonth := 21
	reminders := []model.Reminder{
		{
			ID:      "active01",
			Message: "Pay rent",
			Schedule: model.Schedule{
				Type:   model.ScheduleTypeOnce,
				Hour:   9,
				Minute: 30,
				Date:   &date,
			},
			Status: model.StatusActive,
		},
		{
			ID:      "done0001",
			Message: "Submit report",
			Schedule: model.Schedule{
				Type:       model.ScheduleTypeMonthly,
				Hour:       18,
				Minute:     0,
				DayOfMonth: &dayOfMonth,
			},
			Status: model.StatusCompleted,
		},
	}

	var buf bytes.Buffer
	if err := PrintReminderList(&buf, reminders); err != nil {
		t.Fatalf("PrintReminderList returned error: %v", err)
	}

	text := buf.String()
	for _, want := range []string{"ID", "STATUS", "active01", "done0001", "Pay rent", "Submit report"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected output to contain %q\n%s", want, text)
		}
	}
}

func TestPrintReminderDetailFullReminder(t *testing.T) {
	date := "2026-04-15"
	firedAt := time.Date(2026, time.April, 15, 9, 30, 0, 0, time.Local)
	reminder := model.Reminder{
		ID:      "full0001",
		Message: "Stretch",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   9,
			Minute: 30,
			Date:   &date,
		},
		Status:    model.StatusCompleted,
		CreatedAt: time.Date(2026, time.April, 1, 8, 0, 0, 0, time.Local),
		FiredAt:   &firedAt,
		PlistPath: "/tmp/full0001.plist",
	}

	var buf bytes.Buffer
	if err := PrintReminderDetail(&buf, reminder); err != nil {
		t.Fatalf("PrintReminderDetail returned error: %v", err)
	}

	text := buf.String()
	for _, want := range []string{
		"ID",
		"full0001",
		"Message",
		"Stretch",
		"Status",
		"completed",
		"Schedule",
		"once on 2026-04-15 at 09:30",
		"Created",
		"Fired",
		"/tmp/full0001.plist",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected output to contain %q\n%s", want, text)
		}
	}
}
