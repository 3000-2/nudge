package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/3000-2/nudge/internal/store"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show one reminder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := newStore()
			if err != nil {
				return err
			}

			reminder, err := st.Get(args[0])
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					return fmt.Errorf("reminder %q not found", args[0])
				}
				return err
			}

			return writeReminder(cmd, reminder)
		},
	}
}
