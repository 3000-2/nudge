package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/3000-2/nudge/internal/idgen"
	"github.com/3000-2/nudge/internal/output"
	"github.com/3000-2/nudge/internal/store"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a reminder",
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

			if reminder.PlistPath != "" {
				if err := unloadPlist(reminder.PlistPath); err != nil {
					return err
				}
				if err := removeFile(reminder.PlistPath); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("remove plist: %w", err)
				}
			}

			deleted, err := st.Delete(id)
			if err != nil {
				return err
			}

			if jsonEnabled(cmd) {
				return output.PrintJSON(cmd.OutOrStdout(), map[string]any{
					"id":      id,
					"deleted": deleted,
				})
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Deleted reminder %s\n", id)
			return err
		},
	}
}
