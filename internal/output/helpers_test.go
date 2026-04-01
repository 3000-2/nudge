package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/3000-2/nudge/internal/model"
)

func TestFormatScheduleVariants(t *testing.T) {
	weekday := 1
	dayOfMonth := 21
	date := "2026-04-02"

	cases := []struct {
		name     string
		schedule model.Schedule
		want     string
	}{
		{
			name: "daily",
			schedule: model.Schedule{
				Type:   model.ScheduleTypeDaily,
				Hour:   7,
				Minute: 0,
			},
			want: "daily at 07:00",
		},
		{
			name: "weekly",
			schedule: model.Schedule{
				Type:    model.ScheduleTypeWeekly,
				Hour:    9,
				Minute:  15,
				Weekday: &weekday,
			},
			want: "every monday at 09:15",
		},
		{
			name: "monthly",
			schedule: model.Schedule{
				Type:       model.ScheduleTypeMonthly,
				Hour:       18,
				Minute:     45,
				DayOfMonth: &dayOfMonth,
			},
			want: "every 21st at 18:45",
		},
		{
			name: "once",
			schedule: model.Schedule{
				Type:   model.ScheduleTypeOnce,
				Hour:   12,
				Minute: 30,
				Date:   &date,
			},
			want: "once on 2026-04-02 at 12:30",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatSchedule(tc.schedule); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestPrintReminderListEmptyAndSingleLine(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintReminderList(&buf, nil); err != nil {
		t.Fatalf("PrintReminderList returned error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "No reminders." {
		t.Fatalf("unexpected empty list output %q", buf.String())
	}

	if got := SingleLine("  hello\nworld  "); got != "hello world" {
		t.Fatalf("expected single line output, got %q", got)
	}
}
