package cmd

import (
	"github.com/spf13/cobra"

	"github.com/3000-2/nudge/internal/output"
)

func newListCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List reminders",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := newStore()
			if err != nil {
				return err
			}

			reminders, err := st.List(all)
			if err != nil {
				return err
			}

			if jsonEnabled(cmd) {
				return output.PrintJSON(cmd.OutOrStdout(), reminders)
			}
			return output.PrintReminderList(cmd.OutOrStdout(), reminders)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "include completed reminders")

	return cmd
}
