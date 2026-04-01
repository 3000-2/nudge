package scheduler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/3000-2/nudge/internal/model"
)

func TestPlistTemplateDataAllScheduleTypes(t *testing.T) {
	date := "2026-04-15"
	weekday := 2
	dayOfMonth := 21

	tests := []struct {
		name     string
		reminder model.Reminder
		check    func(t *testing.T, data plistData)
	}{
		{
			name: "once",
			reminder: model.Reminder{
				ID: "once01",
				Schedule: model.Schedule{
					Type:   model.ScheduleTypeOnce,
					Hour:   8,
					Minute: 45,
					Date:   &date,
				},
			},
			check: func(t *testing.T, data plistData) {
				if data.Month != 4 || data.Day != 15 {
					t.Fatalf("expected month/day 4/15, got %d/%d", data.Month, data.Day)
				}
			},
		},
		{
			name: "daily",
			reminder: model.Reminder{
				ID: "daily01",
				Schedule: model.Schedule{
					Type:   model.ScheduleTypeDaily,
					Hour:   9,
					Minute: 15,
				},
			},
			check: func(t *testing.T, data plistData) {
				if data.Month != 0 || data.Day != 0 || data.HasWeekday {
					t.Fatalf("expected daily schedule to omit calendar extras, got %#v", data)
				}
			},
		},
		{
			name: "weekly",
			reminder: model.Reminder{
				ID: "weekly01",
				Schedule: model.Schedule{
					Type:    model.ScheduleTypeWeekly,
					Hour:    10,
					Minute:  30,
					Weekday: &weekday,
				},
			},
			check: func(t *testing.T, data plistData) {
				if !data.HasWeekday || data.Weekday != 2 {
					t.Fatalf("expected weekday 2, got %#v", data)
				}
			},
		},
		{
			name: "monthly",
			reminder: model.Reminder{
				ID: "monthly1",
				Schedule: model.Schedule{
					Type:       model.ScheduleTypeMonthly,
					Hour:       11,
					Minute:     45,
					DayOfMonth: &dayOfMonth,
				},
			},
			check: func(t *testing.T, data plistData) {
				if data.Day != 21 || data.Month != 0 || data.HasWeekday {
					t.Fatalf("expected monthly schedule day 21 only, got %#v", data)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := plistTemplateData(tt.reminder, "/tmp/remind", "/tmp/logs")
			if err != nil {
				t.Fatalf("plistTemplateData returned error: %v", err)
			}
			if data.ID != tt.reminder.ID || data.BinaryPath != "/tmp/remind" || data.LogDir != "/tmp/logs" {
				t.Fatalf("unexpected template data %#v", data)
			}
			tt.check(t, data)
		})
	}
}

func TestPlistTemplateDataErrors(t *testing.T) {
	tests := []struct {
		name     string
		reminder model.Reminder
		wantErr  string
	}{
		{
			name: "weekly missing weekday",
			reminder: model.Reminder{
				Schedule: model.Schedule{Type: model.ScheduleTypeWeekly},
			},
			wantErr: "weekly schedule missing weekday",
		},
		{
			name: "monthly missing day",
			reminder: model.Reminder{
				Schedule: model.Schedule{Type: model.ScheduleTypeMonthly},
			},
			wantErr: "monthly schedule missing day",
		},
		{
			name: "once missing date",
			reminder: model.Reminder{
				Schedule: model.Schedule{Type: model.ScheduleTypeOnce},
			},
			wantErr: "one-time schedule missing date",
		},
		{
			name: "unsupported type",
			reminder: model.Reminder{
				Schedule: model.Schedule{Type: "unknown"},
			},
			wantErr: "unsupported schedule type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := plistTemplateData(tt.reminder, "/tmp/remind", "/tmp/logs")
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestRenderPlistIncludesExpectedCalendarKeys(t *testing.T) {
	date := "2026-04-15"
	weekday := 4
	dayOfMonth := 31

	tests := []struct {
		name        string
		reminder    model.Reminder
		wantParts   []string
		absentParts []string
	}{
		{
			name: "once",
			reminder: model.Reminder{
				ID: "once01",
				Schedule: model.Schedule{
					Type:   model.ScheduleTypeOnce,
					Hour:   8,
					Minute: 5,
					Date:   &date,
				},
			},
			wantParts: []string{
				"<key>Month</key>",
				"<integer>4</integer>",
				"<key>Day</key>",
				"<integer>15</integer>",
				"<key>Hour</key>",
				"<integer>8</integer>",
				"<key>Minute</key>",
				"<integer>5</integer>",
			},
			absentParts: []string{"<key>Weekday</key>"},
		},
		{
			name: "daily",
			reminder: model.Reminder{
				ID: "daily01",
				Schedule: model.Schedule{
					Type:   model.ScheduleTypeDaily,
					Hour:   7,
					Minute: 10,
				},
			},
			wantParts: []string{
				"<key>Hour</key>",
				"<integer>7</integer>",
				"<key>Minute</key>",
				"<integer>10</integer>",
			},
			absentParts: []string{"<key>Month</key>", "<key>Day</key>", "<key>Weekday</key>"},
		},
		{
			name: "weekly",
			reminder: model.Reminder{
				ID: "weekly01",
				Schedule: model.Schedule{
					Type:    model.ScheduleTypeWeekly,
					Hour:    6,
					Minute:  20,
					Weekday: &weekday,
				},
			},
			wantParts: []string{
				"<key>Weekday</key>",
				"<integer>4</integer>",
			},
			absentParts: []string{"<key>Month</key>"},
		},
		{
			name: "monthly",
			reminder: model.Reminder{
				ID: "monthly1",
				Schedule: model.Schedule{
					Type:       model.ScheduleTypeMonthly,
					Hour:       21,
					Minute:     55,
					DayOfMonth: &dayOfMonth,
				},
			},
			wantParts: []string{
				"<key>Day</key>",
				"<integer>31</integer>",
			},
			absentParts: []string{"<key>Month</key>", "<key>Weekday</key>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := RenderPlist(tt.reminder, "/tmp/remind", "/tmp/logs")
			if err != nil {
				t.Fatalf("RenderPlist returned error: %v", err)
			}
			text := string(content)
			for _, want := range tt.wantParts {
				if !strings.Contains(text, want) {
					t.Fatalf("expected plist to contain %q\n%s", want, text)
				}
			}
			for _, absent := range tt.absentParts {
				if strings.Contains(text, absent) {
					t.Fatalf("expected plist to omit %q\n%s", absent, text)
				}
			}
		})
	}
}

func TestRenderPlistEscapesXML(t *testing.T) {
	date := "2026-04-15"
	reminder := model.Reminder{
		ID:      `id<&>"'`,
		Message: `msg<&>"'`,
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   8,
			Minute: 30,
			Date:   &date,
		},
	}

	content, err := RenderPlist(reminder, `/tmp/bin<&>"'`, `/tmp/log<&>"'`)
	if err != nil {
		t.Fatalf("RenderPlist returned error: %v", err)
	}

	text := string(content)
	for _, want := range []string{
		"com.nudge.id&lt;&amp;&gt;&#34;&#39;",
		"/tmp/bin&lt;&amp;&gt;&#34;&#39;",
		"/tmp/log&lt;&amp;&gt;&#34;&#39;/id&lt;&amp;&gt;&#34;&#39;.stdout.log",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected escaped text %q\n%s", want, text)
		}
	}
}

func TestWritePlistErrorPaths(t *testing.T) {
	t.Run("mkdir failure", func(t *testing.T) {
		parentFile := filepath.Join(t.TempDir(), "parent-file")
		if err := os.WriteFile(parentFile, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile returned error: %v", err)
		}

		err := WritePlist(filepath.Join(parentFile, "test.plist"), []byte("plist"))
		if err == nil || !strings.Contains(err.Error(), "create launch agents directory") {
			t.Fatalf("expected create directory error, got %v", err)
		}
	})

	t.Run("rename failure cleans temp file", func(t *testing.T) {
		baseDir := t.TempDir()
		targetDir := filepath.Join(baseDir, "target-dir")
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			t.Fatalf("MkdirAll returned error: %v", err)
		}

		err := WritePlist(targetDir, []byte("plist"))
		if err == nil || !strings.Contains(err.Error(), "replace plist file") {
			t.Fatalf("expected replace plist file error, got %v", err)
		}

		matches, err := filepath.Glob(filepath.Join(baseDir, "nudge-*.plist"))
		if err != nil {
			t.Fatalf("Glob returned error: %v", err)
		}
		if len(matches) != 0 {
			t.Fatalf("expected temp plist files to be cleaned up, got %v", matches)
		}
	})
}
