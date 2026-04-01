package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
	"github.com/3000-2/nudge/internal/store"
)

func TestAddListShowDeleteFlow(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	restore := stubCommandDeps(t)
	defer restore()

	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local)
	}
	generateID = func(func(string) (bool, error)) (string, error) {
		return "abcd1234", nil
	}
	loadPlist = func(string) error { return nil }
	unloadPlist = func(string) error { return nil }
	notifyReminder = func(string) error { return nil }

	addOut := executeCommand(t, "test", "add", "Pay rent", "--at", "15:30", "--json")
	var added model.Reminder
	if err := json.Unmarshal([]byte(addOut), &added); err != nil {
		t.Fatalf("failed to parse add output: %v\n%s", err, addOut)
	}
	if added.ID != "abcd1234" {
		t.Fatalf("expected reminder id abcd1234, got %q", added.ID)
	}

	listOut := executeCommand(t, "test", "list", "--json")
	var reminders []model.Reminder
	if err := json.Unmarshal([]byte(listOut), &reminders); err != nil {
		t.Fatalf("failed to parse list output: %v\n%s", err, listOut)
	}
	if len(reminders) != 1 {
		t.Fatalf("expected 1 reminder, got %d", len(reminders))
	}

	showOut := executeCommand(t, "test", "show", "abcd1234")
	if !strings.Contains(showOut, "Pay rent") {
		t.Fatalf("expected show output to include message, got:\n%s", showOut)
	}

	deleteOut := executeCommand(t, "test", "delete", "abcd1234", "--json")
	if !strings.Contains(deleteOut, `"deleted": true`) {
		t.Fatalf("expected delete output to confirm deletion, got:\n%s", deleteOut)
	}

	st, err := store.NewDefault()
	if err != nil {
		t.Fatalf("store.NewDefault returned error: %v", err)
	}
	all, err := st.List(true)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(all) != 0 {
		t.Fatalf("expected store to be empty, got %d reminders", len(all))
	}
	if _, err := os.Stat(filepath.Join(home, "Library", "LaunchAgents", "com.nudge.abcd1234.plist")); !os.IsNotExist(err) {
		t.Fatalf("expected plist removed, stat err=%v", err)
	}
}

func TestNotifyCompletesOneTimeReminder(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	restore := stubCommandDeps(t)
	defer restore()

	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 1, 12, 0, 0, 0, time.Local)
	}
	loadPlist = func(string) error { return nil }
	unloadPlist = func(string) error { return nil }
	notifyReminder = func(string) error { return nil }

	st, err := store.NewDefault()
	if err != nil {
		t.Fatalf("store.NewDefault returned error: %v", err)
	}

	date := "2026-04-01"
	plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.nudge.abcd1234.plist")
	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(plistPath, []byte("plist"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	reminder := model.Reminder{
		ID:      "abcd1234",
		Message: "Stretch",
		Schedule: model.Schedule{
			Type:   model.ScheduleTypeOnce,
			Hour:   12,
			Minute: 0,
			Date:   &date,
		},
		Status:    model.StatusActive,
		CreatedAt: time.Date(2026, time.April, 1, 9, 0, 0, 0, time.Local),
		PlistPath: plistPath,
	}
	if err := st.Add(reminder); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	executeCommand(t, "test", "notify", "abcd1234")

	got, err := st.Get("abcd1234")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Status != model.StatusCompleted {
		t.Fatalf("expected completed status, got %q", got.Status)
	}
	if got.FiredAt == nil {
		t.Fatal("expected fired_at to be set")
	}
	if _, err := os.Stat(plistPath); !os.IsNotExist(err) {
		t.Fatalf("expected plist removed, stat err=%v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	output := executeCommand(t, "1.2.3", "version")
	if strings.TrimSpace(output) != "1.2.3" {
		t.Fatalf("expected version output, got %q", output)
	}
}

func executeCommand(t *testing.T, version string, args ...string) string {
	t.Helper()

	root := NewRootCmd(version)
	var out bytes.Buffer
	var errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v\nstderr:\n%s", err, errBuf.String())
	}

	return out.String()
}

func setHome(t *testing.T, home string) {
	t.Helper()

	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("Setenv returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Setenv("HOME", originalHome)
	})
}

func stubCommandDeps(t *testing.T) func() {
	t.Helper()

	originalNowFunc := nowFunc
	originalGenerateID := generateID
	originalLoadPlist := loadPlist
	originalUnloadPlist := unloadPlist
	originalNotifyReminder := notifyReminder
	originalExecutablePath := executablePath

	executablePath = func() (string, error) {
		return "/tmp/nudge", nil
	}

	return func() {
		nowFunc = originalNowFunc
		generateID = originalGenerateID
		loadPlist = originalLoadPlist
		unloadPlist = originalUnloadPlist
		notifyReminder = originalNotifyReminder
		executablePath = originalExecutablePath
	}
}
