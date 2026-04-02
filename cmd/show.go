package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/3000-2/nudge/internal/idgen"
	"github.com/3000-2/nudge/internal/store"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show one reminder",
		Args:  cobra.ExactArgs(1),
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
					return fmt.Errorf("reminder %q not found", id)
				}
				return err
			}

			return writeReminder(cmd, reminder)
		},
	}
}
