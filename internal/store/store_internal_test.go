package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
)

func TestNewDefaultAndAccessors(t *testing.T) {
	home := t.TempDir()
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("Setenv returned error: %v", err)
	}
	defer os.Setenv("HOME", originalHome)

	st, err := NewDefault()
	if err != nil {
		t.Fatalf("NewDefault returned error: %v", err)
	}

	wantFile := filepath.Join(home, ".nudge", "reminders.json")
	wantLogs := filepath.Join(home, ".nudge", "logs")
	if got := st.FilePath(); got != wantFile {
		t.Fatalf("expected file path %q, got %q", wantFile, got)
	}
	if got := st.LogsDir(); got != wantLogs {
		t.Fatalf("expected logs dir %q, got %q", wantLogs, got)
	}
	if _, err := os.Stat(wantLogs); err != nil {
		t.Fatalf("expected logs dir to exist: %v", err)
	}
}

func TestReadStorageLockedHandlesEmptyAndCorruptFiles(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "empty file",
			content: " \n\t",
		},
		{
			name:    "corrupt json",
			content: "{",
			wantErr: "decode storage file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(st.filePath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("WriteFile returned error: %v", err)
			}

			storage, err := st.readStorageLocked()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("readStorageLocked returned error: %v", err)
			}
			if storage.Version != storageVersion {
				t.Fatalf("expected version %d, got %d", storageVersion, storage.Version)
			}
			if len(storage.Reminders) != 0 {
				t.Fatalf("expected empty reminders, got %d", len(storage.Reminders))
			}
		})
	}
}

func TestWriteStorageLockedNormalizesNilReminders(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if err := st.writeStorageLocked(model.StorageFile{}); err != nil {
		t.Fatalf("writeStorageLocked returned error: %v", err)
	}

	data, err := os.ReadFile(st.filePath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "\"version\": 1") || !strings.Contains(text, "\"reminders\": []") {
		t.Fatalf("unexpected storage file contents:\n%s", text)
	}
}

func TestWriteStorageLockedErrors(t *testing.T) {
	t.Run("create temp error", func(t *testing.T) {
		baseFile := filepath.Join(t.TempDir(), "base-file")
		if err := os.WriteFile(baseFile, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile returned error: %v", err)
		}

		st := &Store{
			baseDir:  baseFile,
			filePath: filepath.Join(baseFile, "reminders.json"),
		}

		if err := st.writeStorageLocked(model.StorageFile{}); err == nil || !strings.Contains(err.Error(), "create temp storage file") {
			t.Fatalf("expected create temp error, got %v", err)
		}
	})

	t.Run("replace file error cleans temp file", func(t *testing.T) {
		baseDir := t.TempDir()
		targetDir := filepath.Join(baseDir, "target-dir")
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			t.Fatalf("MkdirAll returned error: %v", err)
		}

		st := &Store{
			baseDir:  baseDir,
			filePath: targetDir,
		}

		if err := st.writeStorageLocked(model.StorageFile{}); err == nil || !strings.Contains(err.Error(), "replace storage file") {
			t.Fatalf("expected replace storage file error, got %v", err)
		}

		matches, err := filepath.Glob(filepath.Join(baseDir, "nudge-*.tmp"))
		if err != nil {
			t.Fatalf("Glob returned error: %v", err)
		}
		if len(matches) != 0 {
			t.Fatalf("expected temp files to be cleaned up, got %v", matches)
		}
	})
}

func TestCloneScheduleAndCloneReminderDeepCopy(t *testing.T) {
	date := "2026-04-02"
	weekday := 3
	dayOfMonth := 21
	firedAt := time.Date(2026, time.April, 1, 11, 0, 0, 0, time.Local)
	originalFiredAt := firedAt

	schedule := model.Schedule{
		Type:       model.ScheduleTypeOnce,
		Hour:       9,
		Minute:     30,
		Date:       &date,
		Weekday:    &weekday,
		DayOfMonth: &dayOfMonth,
	}
	clonedSchedule := cloneSchedule(schedule)

	if clonedSchedule.Date == schedule.Date || clonedSchedule.Weekday == schedule.Weekday || clonedSchedule.DayOfMonth == schedule.DayOfMonth {
		t.Fatal("expected cloneSchedule to deep copy pointer fields")
	}

	*schedule.Date = "2026-05-01"
	*schedule.Weekday = 5
	*schedule.DayOfMonth = 31

	if *clonedSchedule.Date != "2026-04-02" || *clonedSchedule.Weekday != 3 || *clonedSchedule.DayOfMonth != 21 {
		t.Fatalf("cloneSchedule was mutated: %#v", clonedSchedule)
	}

	reminder := model.Reminder{
		ID:        "abcd1234",
		Message:   "Stretch",
		Schedule:  schedule,
		Status:    model.StatusActive,
		CreatedAt: time.Date(2026, time.April, 1, 9, 0, 0, 0, time.Local),
		FiredAt:   &firedAt,
	}
	clonedReminder := cloneReminder(reminder)
	if clonedReminder.FiredAt == reminder.FiredAt {
		t.Fatal("expected cloneReminder to deep copy firedAt")
	}

	*reminder.FiredAt = reminder.FiredAt.Add(2 * time.Hour)
	if clonedReminder.FiredAt == nil || !clonedReminder.FiredAt.Equal(originalFiredAt) {
		t.Fatalf("expected cloned firedAt to remain %v, got %#v", originalFiredAt, clonedReminder.FiredAt)
	}
}

func TestStoreDuplicateAndMissingOperations(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

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

	if err := st.Add(reminder); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if err := st.Add(reminder); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected duplicate add error, got %v", err)
	}

	if _, err := st.Update("missing", func(*model.Reminder) error { return nil }); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	deleted, err := st.Delete("missing")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if deleted {
		t.Fatal("expected deleted to be false for missing reminder")
	}
}
