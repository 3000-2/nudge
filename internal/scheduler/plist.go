package scheduler

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/3000-2/nudge/internal/model"
)

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.nudge.{{xml .ID}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{xml .BinaryPath}}</string>
        <string>notify</string>
        <string>{{xml .ID}}</string>
    </array>
    <key>StartCalendarInterval</key>
    <dict>
        {{- if .Month }}
        <key>Month</key>
        <integer>{{.Month}}</integer>
        {{- end }}
        {{- if .Day }}
        <key>Day</key>
        <integer>{{.Day}}</integer>
        {{- end }}
        {{- if .HasWeekday }}
        <key>Weekday</key>
        <integer>{{.Weekday}}</integer>
        {{- end }}
        <key>Hour</key>
        <integer>{{.Hour}}</integer>
        <key>Minute</key>
        <integer>{{.Minute}}</integer>
    </dict>
    <key>StandardOutPath</key>
    <string>{{xml .LogDir}}/{{xml .ID}}.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>{{xml .LogDir}}/{{xml .ID}}.stderr.log</string>
</dict>
</plist>
`

type plistData struct {
	ID         string
	BinaryPath string
	LogDir     string
	Hour       int
	Minute     int
	Month      int
	Day        int
	HasWeekday bool
	Weekday    int
}

func ExecutablePath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}

	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve executable symlink: %w", err)
	}

	return realPath, nil
}

func DefaultLaunchAgentsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents"), nil
}

func PlistPath(launchAgentsDir, id string) string {
	return filepath.Join(launchAgentsDir, fmt.Sprintf("com.nudge.%s.plist", id))
}

var plistTmpl = template.Must(
	template.New("launchd").Funcs(template.FuncMap{
		"xml": escapeXML,
	}).Parse(plistTemplate),
)

func RenderPlist(reminder model.Reminder, binaryPath, logDir string) ([]byte, error) {
	data, err := plistTemplateData(reminder, binaryPath, logDir)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := plistTmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render plist template: %w", err)
	}

	return buf.Bytes(), nil
}

func WritePlist(plistPath string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return fmt.Errorf("create launch agents directory: %w", err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(plistPath), "nudge-*.plist")
	if err != nil {
		return fmt.Errorf("create temp plist file: %w", err)
	}

	tempPath := tempFile.Name()
	committed := false
	defer func() {
		_ = tempFile.Close()
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(content); err != nil {
		return fmt.Errorf("write plist file: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("sync plist file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close plist file: %w", err)
	}
	if err := os.Chmod(tempPath, 0o644); err != nil {
		return fmt.Errorf("chmod plist file: %w", err)
	}
	if err := os.Rename(tempPath, plistPath); err != nil {
		return fmt.Errorf("replace plist file: %w", err)
	}
	committed = true
	return nil
}

func plistTemplateData(reminder model.Reminder, binaryPath, logDir string) (plistData, error) {
	data := plistData{
		ID:         reminder.ID,
		BinaryPath: binaryPath,
		LogDir:     logDir,
		Hour:       reminder.Schedule.Hour,
		Minute:     reminder.Schedule.Minute,
	}

	switch reminder.Schedule.Type {
	case model.ScheduleTypeDaily:
		return data, nil
	case model.ScheduleTypeWeekly:
		if reminder.Schedule.Weekday == nil {
			return plistData{}, fmt.Errorf("weekly schedule missing weekday")
		}
		data.HasWeekday = true
		data.Weekday = *reminder.Schedule.Weekday
		return data, nil
	case model.ScheduleTypeMonthly:
		if reminder.Schedule.DayOfMonth == nil {
			return plistData{}, fmt.Errorf("monthly schedule missing day")
		}
		data.Day = *reminder.Schedule.DayOfMonth
		return data, nil
	case model.ScheduleTypeOnce:
		if reminder.Schedule.Date == nil {
			return plistData{}, fmt.Errorf("one-time schedule missing date")
		}
		date, err := time.ParseInLocation("2006-01-02", *reminder.Schedule.Date, time.Local)
		if err != nil {
			return plistData{}, fmt.Errorf("parse one-time schedule date: %w", err)
		}
		data.Month = int(date.Month())
		data.Day = date.Day()
		return data, nil
	default:
		return plistData{}, fmt.Errorf("unsupported schedule type %q", reminder.Schedule.Type)
	}
}

var xmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&#34;",
	"'", "&#39;",
)

func escapeXML(value string) string {
	return xmlReplacer.Replace(value)
}
