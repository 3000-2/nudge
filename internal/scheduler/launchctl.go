package scheduler

import (
	"fmt"
	"os/exec"
	"strings"
)

var runLaunchctl = func(args ...string) ([]byte, error) {
	cmd := exec.Command("launchctl", args...)
	return cmd.CombinedOutput()
}

func Load(plistPath string) error {
	return executeLaunchctl(true, "load", "-w", plistPath)
}

func Unload(plistPath string) error {
	return executeLaunchctl(false, "unload", plistPath)
}

func executeLaunchctl(load bool, args ...string) error {
	output, err := runLaunchctl(args...)
	if err == nil {
		return nil
	}

	text := strings.TrimSpace(string(output))
	if isBenignLaunchctlError(load, text) {
		return nil
	}
	if text == "" {
		return fmt.Errorf("launchctl %s: %w", strings.Join(args, " "), err)
	}
	return fmt.Errorf("launchctl %s: %w: %s", strings.Join(args, " "), err, text)
}

func isBenignLaunchctlError(load bool, output string) bool {
	lower := strings.ToLower(output)
	if load {
		return strings.Contains(lower, "already loaded") || strings.Contains(lower, "in progress")
	}

	return strings.Contains(lower, "could not find specified service") ||
		strings.Contains(lower, "no such process") ||
		strings.Contains(lower, "not loaded") ||
		strings.Contains(lower, "no such file")
}
