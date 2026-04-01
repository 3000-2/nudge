package notifier

import (
	"fmt"
	"os/exec"
	"strings"
)

var runCommand = func(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			return fmt.Errorf("execute %s: %w", name, err)
		}
		return fmt.Errorf("execute %s: %w: %s", name, err, text)
	}
	return nil
}

func Notify(message string) error {
	script := `on run argv
	display notification (item 1 of argv) with title "⏰ Remind" sound name "Glass"
end run`

	if err := runCommand("osascript", "-e", script, message); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}
	return nil
}
