package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
	"github.com/3000-2/nudge/internal/store"
)

func TestNotifyOneTimeReminderOnCorrectDate(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 12, 0, 0, 0, time.Local))
	date := "2026-04-01"
	plistPath := filepath.Join(env.agentsDir, "com.nudge.once001.plist")
	env.writePlistFile(plistPath)

	env.seedReminder(model.Reminder{
		ID:      "once001",
		Message: "Take medicine",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   12,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: env.now.Add(-time.Hour),
		PlistPath: plistPath,
	})

	_, _, err := executeCommandResult(t, "test", "notify", "once001")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	got, err := env.store().Get("once001")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if len(env.notifications) != 1 || env.notifications[0] != "Take medicine" {
		t.Fatalf("expected notification for reminder message, got %v", env.notifications)
	}
	if got.Status != model.StatusCompleted {
		t.Fatalf("expected completed status, got %q", got.Status)
	}
	if got.FiredAt == nil || !got.FiredAt.Equal(env.now) {
		t.Fatalf("expected firedAt %v, got %#v", env.now, got.FiredAt)
	}
	if len(env.unloaded) != 1 || env.unloaded[0] != plistPath {
		t.Fatalf("expected unloadPlist to be called with %q, got %v", plistPath, env.unloaded)
	}
	if _, err := os.Stat(plistPath); !os.IsNotExist(err) {
		t.Fatalf("expected plist removed, stat err=%v", err)
	}
}

func TestNotifyWrongDateCleansUpOnceReminder(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 12, 0, 0, 0, time.Local))
	date := "2026-04-02"
	plistPath := filepath.Join(env.agentsDir, "com.nudge.once002.plist")
	env.writePlistFile(plistPath)

	env.seedReminder(model.Reminder{
		ID:      "once002",
		Message: "Wrong day",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   12,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: env.now.Add(-time.Hour),
		PlistPath: plistPath,
	})

	_, _, err := executeCommandResult(t, "test", "notify", "once002")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	got, err := env.store().Get("once002")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if len(env.notifications) != 0 {
		t.Fatalf("expected no notification for stale once reminder, got %v", env.notifications)
	}
	if got.Status != model.StatusCompleted {
		t.Fatalf("expected completed status, got %q", got.Status)
	}
	if got.FiredAt == nil || !got.FiredAt.Equal(env.now) {
		t.Fatalf("expected firedAt %v, got %#v", env.now, got.FiredAt)
	}
	if _, err := os.Stat(plistPath); !os.IsNotExist(err) {
		t.Fatalf("expected plist removed, stat err=%v", err)
	}
}

func TestNotifyAlreadyCompletedReminderIsNoOp(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 12, 0, 0, 0, time.Local))
	date := "2026-04-01"
	plistPath := filepath.Join(env.agentsDir, "com.nudge.done0001.plist")
	env.writePlistFile(plistPath)

	env.seedReminder(model.Reminder{
		ID:      "done0001",
		Message: "Already done",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   12,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusCompleted,
		CreatedAt: env.now.Add(-time.Hour),
		PlistPath: plistPath,
	})

	_, _, err := executeCommandResult(t, "test", "notify", "done0001")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	got, err := env.store().Get("done0001")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.FiredAt != nil {
		t.Fatalf("expected firedAt to remain nil, got %#v", got.FiredAt)
	}
	if len(env.notifications) != 0 || len(env.unloaded) != 0 {
		t.Fatalf("expected no side effects, notifications=%v unloaded=%v", env.notifications, env.unloaded)
	}
	if _, err := os.Stat(plistPath); err != nil {
		t.Fatalf("expected plist to remain in place: %v", err)
	}
}

func TestNotifyMissingReminderIsNoOp(t *testing.T) {
	setupCommandTestEnv(t, time.Date(2026, time.April, 1, 12, 0, 0, 0, time.Local))

	_, _, err := executeCommandResult(t, "test", "notify", "missing")
	if err != nil {
		t.Fatalf("expected missing reminder notify to be no-op, got %v", err)
	}
}

func TestNotifyRecurringReminderStaysActive(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 12, 0, 0, 0, time.Local))
	weekday := 3
	plistPath := filepath.Join(env.agentsDir, "com.nudge.week0001.plist")
	env.writePlistFile(plistPath)

	env.seedReminder(model.Reminder{
		ID:      "week0001",
		Message: "Weekly review",
		Schedule: model.Schedule{
			Type:    model.ScheduleTypeWeekly,
			Hour:    12,
			Minute:  0,
			Weekday: &weekday,
		},
		Status:    model.StatusActive,
		CreatedAt: env.now.Add(-time.Hour),
		PlistPath: plistPath,
	})

	_, _, err := executeCommandResult(t, "test", "notify", "week0001")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	got, err := env.store().Get("week0001")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Status != model.StatusActive {
		t.Fatalf("expected reminder to stay active, got %q", got.Status)
	}
	if got.FiredAt == nil || !got.FiredAt.Equal(env.now) {
		t.Fatalf("expected firedAt %v, got %#v", env.now, got.FiredAt)
	}
	if len(env.notifications) != 1 || env.notifications[0] != "Weekly review" {
		t.Fatalf("expected notification, got %v", env.notifications)
	}
	if len(env.unloaded) != 0 {
		t.Fatalf("expected no plist unload for recurring reminder, got %v", env.unloaded)
	}
	if _, err := os.Stat(plistPath); err != nil {
		t.Fatalf("expected recurring plist to remain: %v", err)
	}
}

func TestCleanupOnceReminderIgnoresMissingReminder(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 12, 0, 0, 0, time.Local))
	plistPath := filepath.Join(env.agentsDir, "com.nudge.clean001.plist")

	if err := cleanupOnceReminder(env.store(), "missing", plistPath, env.now); err != nil {
		t.Fatalf("cleanupOnceReminder returned error: %v", err)
	}
}

func TestCleanupOnceReminderReturnsRemoveError(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 12, 0, 0, 0, time.Local))
	date := "2026-04-02"
	env.seedReminder(model.Reminder{
		ID:      "clean002",
		Message: "Cleanup error",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   12,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: env.now,
		PlistPath: "/tmp/clean002.plist",
	})

	removeFile = func(string) error { return errors.New("remove boom") }

	err := cleanupOnceReminder(env.store(), "clean002", "/tmp/clean002.plist", env.now)
	if err == nil || err.Error() != "remove stale plist: remove boom" {
		t.Fatalf("expected remove error, got %v", err)
	}

	got, err := env.store().Get("clean002")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Status != model.StatusCompleted {
		t.Fatalf("expected cleanup to complete reminder before remove error, got %q", got.Status)
	}
}

func TestNotifyMissingReminderStillAbsent(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 12, 0, 0, 0, time.Local))

	_, _, err := executeCommandResult(t, "test", "notify", "missing-again")
	if err != nil {
		t.Fatalf("expected missing reminder notify to be no-op, got %v", err)
	}
	if _, err := env.store().Get("missing-again"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
