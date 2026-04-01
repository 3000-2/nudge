package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/3000-2/nudge/internal/model"
	"github.com/3000-2/nudge/internal/output"
	"github.com/3000-2/nudge/internal/parser"
	"github.com/3000-2/nudge/internal/scheduler"
	"github.com/3000-2/nudge/internal/store"
)

func newAddCmd() *cobra.Command {
	var at string
	var on string
	var every string
	var next string

	cmd := &cobra.Command{
		Use:   `add "message"`,
		Short: "Add a reminder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			message := args[0]
			now := nowFunc()

			schedule, err := parser.ParseSchedule(at, on, every, next, now)
			if err != nil {
				return err
			}

			st, err := newStore()
			if err != nil {
				return err
			}

			id, err := generateID(st.Exists)
			if err != nil {
				return err
			}

			binaryPath, err := executablePath()
			if err != nil {
				return err
			}
			agentsDir, err := launchAgentsDir()
			if err != nil {
				return err
			}

			reminder := model.Reminder{
				ID:        id,
				Message:   message,
				Schedule:  schedule,
				Status:    model.StatusActive,
				CreatedAt: now,
				PlistPath: scheduler.PlistPath(agentsDir, id),
			}

			if err := st.Add(reminder); err != nil {
				return err
			}

			rollback := func(cause error) error {
				_ = unloadPlist(reminder.PlistPath)
				_ = removeFile(reminder.PlistPath)
				_, deleteErr := st.Delete(reminder.ID)
				if deleteErr != nil && !errors.Is(deleteErr, store.ErrNotFound) {
					return fmt.Errorf("%w (rollback failed: %v)", cause, deleteErr)
				}
				return cause
			}

			content, err := renderPlist(reminder, binaryPath, st.LogsDir())
			if err != nil {
				return rollback(err)
			}
			if err := writePlist(reminder.PlistPath, content); err != nil {
				return rollback(err)
			}
			if err := loadPlist(reminder.PlistPath); err != nil {
				return rollback(err)
			}

			if jsonEnabled(cmd) {
				return output.PrintJSON(cmd.OutOrStdout(), reminder)
			}
			return output.PrintReminderDetail(cmd.OutOrStdout(), reminder)
		},
	}

	cmd.Flags().StringVar(&at, "at", "", "time in HH:MM 24h format")
	cmd.Flags().StringVar(&on, "on", "", "specific date in YYYY-MM-DD")
	cmd.Flags().StringVar(&every, "every", "", "repeat every day, weekday, or ordinal day")
	cmd.Flags().StringVar(&next, "next", "", "next weekday occurrence")
	_ = cmd.MarkFlagRequired("at")

	return cmd
}
