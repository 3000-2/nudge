package store_test

import (
	"errors"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
	"github.com/3000-2/nudge/internal/store"
)

func TestStoreCRUD(t *testing.T) {
	st, err := store.New(t.TempDir())
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
		PlistPath: "/tmp/com.nudge.abcd1234.plist",
	}

	if err := st.Add(reminder); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	exists, err := st.Exists(reminder.ID)
	if err != nil {
		t.Fatalf("Exists returned error: %v", err)
	}
	if !exists {
		t.Fatal("expected reminder to exist")
	}

	got, err := st.Get(reminder.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Message != reminder.Message {
		t.Fatalf("expected message %q, got %q", reminder.Message, got.Message)
	}

	updated, err := st.Update(reminder.ID, func(r *model.Reminder) error {
		r.Status = model.StatusCompleted
		return nil
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Status != model.StatusCompleted {
		t.Fatalf("expected completed status, got %q", updated.Status)
	}

	activeOnly, err := st.List(false)
	if err != nil {
		t.Fatalf("List(active) returned error: %v", err)
	}
	if len(activeOnly) != 0 {
		t.Fatalf("expected no active reminders, got %d", len(activeOnly))
	}

	all, err := st.List(true)
	if err != nil {
		t.Fatalf("List(all) returned error: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected one stored reminder, got %d", len(all))
	}

	deleted, err := st.Delete(reminder.ID)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if !deleted {
		t.Fatal("expected delete to report true")
	}

	_, err = st.Get(reminder.ID)
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
