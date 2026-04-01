package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
)

func TestAddCreatesReminderForSchedules(t *testing.T) {
	tests := []struct {
		name       string
		now        time.Time
		args       []string
		check      func(t *testing.T, reminder model.Reminder)
		wantOutput string
	}{
		{
			name: "at only uses today",
			now:  time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local),
			args: []string{"add", "Pay rent", "--at", "14:00"},
			check: func(t *testing.T, reminder model.Reminder) {
				if reminder.Schedule.Type != model.ScheduleTypeOnce {
					t.Fatalf("expected once schedule, got %q", reminder.Schedule.Type)
				}
				if reminder.Schedule.Date == nil || *reminder.Schedule.Date != "2026-04-01" {
					t.Fatalf("expected date 2026-04-01, got %#v", reminder.Schedule.Date)
				}
			},
			wantOutput: "Pay rent",
		},
		{
			name: "at only rolls to tomorrow",
			now:  time.Date(2026, time.April, 1, 18, 0, 0, 0, time.Local),
			args: []string{"add", "Water plants", "--at", "09:15"},
			check: func(t *testing.T, reminder model.Reminder) {
				if reminder.Schedule.Date == nil || *reminder.Schedule.Date != "2026-04-02" {
					t.Fatalf("expected date 2026-04-02, got %#v", reminder.Schedule.Date)
				}
			},
			wantOutput: "Water plants",
		},
		{
			name: "on specific date",
			now:  time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local),
			args: []string{"add", "Doctor appointment", "--at", "16:45", "--on", "2026-04-03"},
			check: func(t *testing.T, reminder model.Reminder) {
				if reminder.Schedule.Date == nil || *reminder.Schedule.Date != "2026-04-03" {
					t.Fatalf("expected date 2026-04-03, got %#v", reminder.Schedule.Date)
				}
			},
			wantOutput: "Doctor appointment",
		},
		{
			name: "every day",
			now:  time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local),
			args: []string{"add", "Stand up", "--at", "08:00", "--every", "day"},
			check: func(t *testing.T, reminder model.Reminder) {
				if reminder.Schedule.Type != model.ScheduleTypeDaily {
					t.Fatalf("expected daily schedule, got %q", reminder.Schedule.Type)
				}
			},
			wantOutput: "Stand up",
		},
		{
			name: "every monday",
			now:  time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local),
			args: []string{"add", "Weekly review", "--at", "09:00", "--every", "monday"},
			check: func(t *testing.T, reminder model.Reminder) {
				if reminder.Schedule.Type != model.ScheduleTypeWeekly {
					t.Fatalf("expected weekly schedule, got %q", reminder.Schedule.Type)
				}
				if reminder.Schedule.Weekday == nil || *reminder.Schedule.Weekday != 1 {
					t.Fatalf("expected weekday 1, got %#v", reminder.Schedule.Weekday)
				}
			},
			wantOutput: "Weekly review",
		},
		{
			name: "next wednesday",
			now:  time.Date(2026, time.April, 1, 14, 0, 0, 0, time.Local),
			args: []string{"add", "Team sync", "--at", "13:00", "--next", "wednesday"},
			check: func(t *testing.T, reminder model.Reminder) {
				if reminder.Schedule.Type != model.ScheduleTypeOnce {
					t.Fatalf("expected once schedule, got %q", reminder.Schedule.Type)
				}
				if reminder.Schedule.Date == nil || *reminder.Schedule.Date != "2026-04-08" {
					t.Fatalf("expected date 2026-04-08, got %#v", reminder.Schedule.Date)
				}
			},
			wantOutput: "Team sync",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupCommandTestEnv(t, tt.now)

			stdout, _, err := executeCommandResult(t, "test", tt.args...)
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if !strings.Contains(stdout, tt.wantOutput) {
				t.Fatalf("expected output to contain %q\n%s", tt.wantOutput, stdout)
			}

			reminders, err := env.store().List(true)
			if err != nil {
				t.Fatalf("List returned error: %v", err)
			}
			if len(reminders) != 1 {
				t.Fatalf("expected 1 reminder, got %d", len(reminders))
			}
			reminder := reminders[0]
			if reminder.ID != env.generatedID {
				t.Fatalf("expected id %q, got %q", env.generatedID, reminder.ID)
			}
			if reminder.PlistPath != filepath.Join(env.agentsDir, "com.nudge.testid01.plist") {
				t.Fatalf("unexpected plist path %q", reminder.PlistPath)
			}
			if _, err := os.Stat(reminder.PlistPath); err != nil {
				t.Fatalf("expected plist file to exist: %v", err)
			}
			if len(env.loaded) != 1 || env.loaded[0] != reminder.PlistPath {
				t.Fatalf("expected loadPlist to be called once for %q, got %v", reminder.PlistPath, env.loaded)
			}
			tt.check(t, reminder)
		})
	}
}

func TestAddJSONOutput(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))

	stdout, _, err := executeCommandResult(t, "test", "add", "Read book", "--at", "20:00", "--json")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var reminder model.Reminder
	if err := json.Unmarshal([]byte(stdout), &reminder); err != nil {
		t.Fatalf("failed to unmarshal add json output: %v\n%s", err, stdout)
	}
	if reminder.ID != env.generatedID {
		t.Fatalf("expected id %q, got %q", env.generatedID, reminder.ID)
	}
	if reminder.Message != "Read book" {
		t.Fatalf("expected message %q, got %q", "Read book", reminder.Message)
	}
}

func TestAddRollbackOnRenderWriteAndLoadFailures(t *testing.T) {
	tests := []struct {
		name          string
		override      func(env *commandTestEnv)
		wantErr       string
		expectRemoved bool
	}{
		{
			name: "render failure",
			override: func(env *commandTestEnv) {
				renderPlist = func(model.Reminder, string, string) ([]byte, error) {
					return nil, errors.New("render boom")
				}
			},
			wantErr: "render boom",
		},
		{
			name: "write failure",
			override: func(env *commandTestEnv) {
				writePlist = func(string, []byte) error {
					return errors.New("write boom")
				}
			},
			wantErr: "write boom",
		},
		{
			name: "load failure",
			override: func(env *commandTestEnv) {
				loadPlist = func(string) error {
					return errors.New("load boom")
				}
			},
			wantErr:       "load boom",
			expectRemoved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))
			tt.override(env)

			_, _, err := executeCommandResult(t, "test", "add", "Rollback me", "--at", "14:00")
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}

			reminders, err := env.store().List(true)
			if err != nil {
				t.Fatalf("List returned error: %v", err)
			}
			if len(reminders) != 0 {
				t.Fatalf("expected rollback to delete stored reminder, got %d reminders", len(reminders))
			}
			if len(env.unloaded) != 1 {
				t.Fatalf("expected unloadPlist during rollback, got %v", env.unloaded)
			}
			if tt.expectRemoved {
				plistPath := filepath.Join(env.agentsDir, "com.nudge.testid01.plist")
				if _, err := os.Stat(plistPath); !os.IsNotExist(err) {
					t.Fatalf("expected plist to be removed, stat err=%v", err)
				}
			}
		})
	}
}

func TestAddRequiresAtFlag(t *testing.T) {
	setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))

	_, _, err := executeCommandResult(t, "test", "add", "Missing time")
	if err == nil || !strings.Contains(err.Error(), `required flag(s) "at" not set`) {
		t.Fatalf("expected required flag error, got %v", err)
	}
}
