package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/3000-2/nudge/internal/idgen"
	"github.com/3000-2/nudge/internal/model"
	"github.com/3000-2/nudge/internal/notifier"
	"github.com/3000-2/nudge/internal/output"
	"github.com/3000-2/nudge/internal/scheduler"
	"github.com/3000-2/nudge/internal/store"
)

var (
	nowFunc = func() time.Time {
		return time.Now().In(time.Local)
	}
	newStore = store.NewDefault

	generateID      = idgen.GenerateUnique
	executablePath  = scheduler.ExecutablePath
	launchAgentsDir = scheduler.DefaultLaunchAgentsDir
	renderPlist     = scheduler.RenderPlist
	writePlist      = scheduler.WritePlist
	loadPlist       = scheduler.Load
	unloadPlist     = scheduler.Unload
	notifyReminder  = notifier.Notify
	removeFile      = os.Remove
)

func NewRootCmd(version string) *cobra.Command {
	if version == "" {
		version = "dev"
	}

	rootCmd := &cobra.Command{
		Use:           "nudge",
		Short:         "Manage macOS reminders with launchd and native notifications",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().Bool("json", false, "output JSON")

	rootCmd.AddCommand(newAddCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newShowCmd())
	rootCmd.AddCommand(newDeleteCmd())
	rootCmd.AddCommand(newNotifyCmd())
	rootCmd.AddCommand(newVersionCmd(version))

	return rootCmd
}

func jsonEnabled(cmd *cobra.Command) bool {
	value, err := cmd.Flags().GetBool("json")
	if err != nil {
		return false
	}
	return value
}

func writeReminder(cmd *cobra.Command, reminder model.Reminder) error {
	if jsonEnabled(cmd) {
		return output.PrintJSON(cmd.OutOrStdout(), reminder)
	}
	return output.PrintReminderDetail(cmd.OutOrStdout(), reminder)
}
