package scheduler_test

import (
	"strings"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
	"github.com/3000-2/nudge/internal/scheduler"
)

func TestRenderPlistForOnceSchedule(t *testing.T) {
	date := "2027-12-25"
	reminder := model.Reminder{
		ID:      "abc12345",
		Message: `call "mom"`,
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   7,
			Minute: 30,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local),
	}

	content, err := scheduler.RenderPlist(reminder, `/tmp/A&B/"nudge"`, `/tmp/log&dir`)
	if err != nil {
		t.Fatalf("RenderPlist returned error: %v", err)
	}

	text := string(content)
	for _, part := range []string{
		"<string>com.nudge.abc12345</string>",
		"<key>Month</key>",
		"<integer>12</integer>",
		"<key>Day</key>",
		"<integer>25</integer>",
		`<string>/tmp/A&amp;B/&#34;nudge&#34;</string>`,
		`<string>/tmp/log&amp;dir/abc12345.stdout.log</string>`,
	} {
		if !strings.Contains(text, part) {
			t.Fatalf("expected plist to contain %q\n%s", part, text)
		}
	}
}

func TestRenderPlistForWeeklySchedule(t *testing.T) {
	weekday := 1
	reminder := model.Reminder{
		ID: "wxyz9876",
		Schedule: model.Schedule{
			Type:    model.ScheduleTypeWeekly,
			Hour:    9,
			Minute:  15,
			Weekday: &weekday,
		},
	}

	content, err := scheduler.RenderPlist(reminder, "/tmp/remind", "/tmp/logs")
	if err != nil {
		t.Fatalf("RenderPlist returned error: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "<key>Weekday</key>") || !strings.Contains(text, "<integer>1</integer>") {
		t.Fatalf("expected weekly plist to include weekday, got\n%s", text)
	}
}
