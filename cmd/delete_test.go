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
	"github.com/3000-2/nudge/internal/store"
)

func TestDeleteExistingReminder(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))
	plistPath := filepath.Join(env.agentsDir, "com.nudge.delete01.plist")
	env.writePlistFile(plistPath)

	date := "2026-04-02"
	env.seedReminder(model.Reminder{
		ID:      "delete01",
		Message: "Delete me",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   9,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: env.now,
		PlistPath: plistPath,
	})

	stdout, _, err := executeCommandResult(t, "test", "delete", "delete01")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(stdout, "Deleted reminder delete01") {
		t.Fatalf("unexpected delete output:\n%s", stdout)
	}
	if len(env.unloaded) != 1 || env.unloaded[0] != plistPath {
		t.Fatalf("expected unloadPlist to be called with %q, got %v", plistPath, env.unloaded)
	}
	if _, err := env.store().Get("delete01"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
	if _, err := os.Stat(plistPath); !os.IsNotExist(err) {
		t.Fatalf("expected plist removed, stat err=%v", err)
	}
}

func TestDeleteNonExistentReminder(t *testing.T) {
	setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))

	_, _, err := executeCommandResult(t, "test", "delete", "missing")
	if err == nil || !strings.Contains(err.Error(), `reminder "missing" not found`) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestDeleteJSONOutput(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))
	plistPath := filepath.Join(env.agentsDir, "com.nudge.delete02.plist")
	env.writePlistFile(plistPath)

	date := "2026-04-03"
	env.seedReminder(model.Reminder{
		ID:      "delete02",
		Message: "Delete me too",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   10,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: env.now,
		PlistPath: plistPath,
	})

	stdout, _, err := executeCommandResult(t, "test", "delete", "delete02", "--json")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to unmarshal delete json output: %v\n%s", err, stdout)
	}
	if payload["id"] != "delete02" {
		t.Fatalf("expected id delete02, got %#v", payload["id"])
	}
	if payload["deleted"] != true {
		t.Fatalf("expected deleted=true, got %#v", payload["deleted"])
	}
}

func TestDeleteIgnoresMissingPlistFile(t *testing.T) {
	env := setupCommandTestEnv(t, time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local))
	plistPath := filepath.Join(env.agentsDir, "com.nudge.delete03.plist")

	date := "2026-04-04"
	env.seedReminder(model.Reminder{
		ID:      "delete03",
		Message: "Missing plist",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   11,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: env.now,
		PlistPath: plistPath,
	})

	stdout, _, err := executeCommandResult(t, "test", "delete", "delete03")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(stdout, "Deleted reminder delete03") {
		t.Fatalf("unexpected delete output:\n%s", stdout)
	}
}
