package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/3000-2/nudge/internal/idgen"
	"github.com/3000-2/nudge/internal/model"
	"github.com/3000-2/nudge/internal/store"
)

func newNotifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "notify <id>",
		Short:  "Send a reminder notification",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if !idgen.IsValid(id) {
				return fmt.Errorf("invalid reminder id %q", id)
			}

			st, err := newStore()
			if err != nil {
				return err
			}

			reminder, err := st.Get(id)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					return nil
				}
				return err
			}
			if reminder.Status != model.StatusActive {
				return nil
			}

			now := nowFunc()
			if reminder.Schedule.Type == model.ScheduleTypeOnce {
				if reminder.Schedule.Date == nil {
					return fmt.Errorf("one-time reminder %q has no scheduled date", id)
				}
				if now.Format("2006-01-02") != *reminder.Schedule.Date {
					// Wrong year/date — this is a stale plist firing annually.
					// Clean up without sending notification.
					return cleanupOnceReminder(st, id, reminder.PlistPath, now)
				}
			}

			if err := notifyReminder(reminder.Message); err != nil {
				return err
			}

			firedAt := now
			updated, err := st.Update(id, func(current *model.Reminder) error {
				if current.Status != model.StatusActive {
					return nil
				}
				current.FiredAt = &firedAt
				if current.Schedule.Type == model.ScheduleTypeOnce {
					current.Status = model.StatusCompleted
				}
				return nil
			})
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					return nil
				}
				return err
			}

			if updated.Schedule.Type != model.ScheduleTypeOnce {
				return nil
			}

			if updated.PlistPath != "" {
				if err := unloadPlist(updated.PlistPath); err != nil {
					return err
				}
				if err := removeFile(updated.PlistPath); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("remove plist: %w", err)
				}
			}

			return nil
		},
	}
}

// cleanupOnceReminder silently marks a stale one-time reminder as completed
// and removes its plist. This handles the case where launchd re-fires the
// plist in a subsequent year (StartCalendarInterval has no Year field).
func cleanupOnceReminder(st *store.Store, id, plistPath string, now time.Time) error {
	_, err := st.Update(id, func(r *model.Reminder) error {
		r.Status = model.StatusCompleted
		r.FiredAt = &now
		return nil
	})
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return err
	}

	if plistPath != "" {
		_ = unloadPlist(plistPath)
		if err := removeFile(plistPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove stale plist: %w", err)
		}
	}
	return nil
}
