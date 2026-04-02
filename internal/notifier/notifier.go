package notifier

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

var resolveNotifyApp = defaultResolveNotifyApp

func defaultResolveNotifyApp() string {
	// 1. Environment variable override
	if p := os.Getenv("NUDGE_NOTIFY_APP"); p != "" {
		if isValidApp(p) {
			return p
		}
	}

	// 2. Resolve relative to the nudge binary
	exe, err := os.Executable()
	if err == nil {
		exe, _ = filepath.EvalSymlinks(exe)
		binDir := filepath.Dir(exe)

		// ../lib/nudge/Nudge.app (standard install)
		candidate := filepath.Join(binDir, "..", "lib", "nudge", "Nudge.app")
		if isValidApp(candidate) {
			return candidate
		}

		// Same directory as binary (development)
		candidate = filepath.Join(binDir, "Nudge.app")
		if isValidApp(candidate) {
			return candidate
		}
	}

	// 3. Hardcoded fallback
	if isValidApp("/usr/local/lib/nudge/Nudge.app") {
		return "/usr/local/lib/nudge/Nudge.app"
	}

	return ""
}

func isValidApp(appPath string) bool {
	binary := filepath.Join(appPath, "Contents", "MacOS", "Nudge")
	info, err := os.Stat(binary)
	if err != nil {
		return false
	}
	return info.Mode()&0o111 != 0
}

func Notify(message string) error {
	appPath := resolveNotifyApp()

	if appPath != "" {
		binary := filepath.Join(appPath, "Contents", "MacOS", "Nudge")
		if err := runCommand(binary, message); err != nil {
			// Fallback to osascript on failure
			return notifyOsascript(message)
		}
		return nil
	}

	return notifyOsascript(message)
}

func notifyOsascript(message string) error {
	script := `on run argv
	display notification (item 1 of argv) with title "⏰ Nudge" sound name "Glass"
end run`

	if err := runCommand("osascript", "-e", script, message); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}
	return nil
}
