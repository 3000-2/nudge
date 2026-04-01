package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/3000-2/nudge/internal/model"
)

func PrintJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func PrintReminderList(w io.Writer, reminders []model.Reminder) error {
	if len(reminders) == 0 {
		_, err := fmt.Fprintln(w, "No reminders.")
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tSTATUS\tWHEN\tMESSAGE"); err != nil {
		return err
	}
	for _, reminder := range reminders {
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\n",
			reminder.ID,
			reminder.Status,
			FormatSchedule(reminder.Schedule),
			reminder.Message,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func PrintReminderDetail(w io.Writer, reminder model.Reminder) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	rows := [][2]string{
		{"ID", reminder.ID},
		{"Message", reminder.Message},
		{"Status", reminder.Status},
		{"Schedule", FormatSchedule(reminder.Schedule)},
		{"Created", formatTimestamp(reminder.CreatedAt)},
		{"Fired", formatOptionalTimestamp(reminder.FiredAt)},
		{"PlistPath", reminder.PlistPath},
	}

	for _, row := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1]); err != nil {
			return err
		}
	}

	return tw.Flush()
}

func FormatSchedule(schedule model.Schedule) string {
	timeText := fmt.Sprintf("%02d:%02d", schedule.Hour, schedule.Minute)

	switch schedule.Type {
	case model.ScheduleTypeDaily:
		return fmt.Sprintf("daily at %s", timeText)
	case model.ScheduleTypeWeekly:
		if schedule.Weekday == nil {
			return fmt.Sprintf("weekly at %s", timeText)
		}
		return fmt.Sprintf("every %s at %s", weekdayName(*schedule.Weekday), timeText)
	case model.ScheduleTypeMonthly:
		if schedule.DayOfMonth == nil {
			return fmt.Sprintf("monthly at %s", timeText)
		}
		return fmt.Sprintf("every %s at %s", ordinal(*schedule.DayOfMonth), timeText)
	case model.ScheduleTypeOnce:
		if schedule.Date == nil {
			return fmt.Sprintf("once at %s", timeText)
		}
		return fmt.Sprintf("once on %s at %s", *schedule.Date, timeText)
	default:
		return fmt.Sprintf("%s at %s", schedule.Type, timeText)
	}
}

func formatTimestamp(value time.Time) string {
	return value.In(time.Local).Format("2006-01-02 15:04:05 MST")
}

func formatOptionalTimestamp(value *time.Time) string {
	if value == nil {
		return "-"
	}
	return formatTimestamp(*value)
}

func weekdayName(weekday int) string {
	names := []string{
		"sunday",
		"monday",
		"tuesday",
		"wednesday",
		"thursday",
		"friday",
		"saturday",
	}
	if weekday < 0 || weekday >= len(names) {
		return fmt.Sprintf("weekday-%d", weekday)
	}
	return names[weekday]
}

func ordinal(day int) string {
	switch {
	case day%100 >= 11 && day%100 <= 13:
		return fmt.Sprintf("%dth", day)
	case day%10 == 1:
		return fmt.Sprintf("%dst", day)
	case day%10 == 2:
		return fmt.Sprintf("%dnd", day)
	case day%10 == 3:
		return fmt.Sprintf("%drd", day)
	default:
		return fmt.Sprintf("%dth", day)
	}
}

func SingleLine(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "\n", " ")
}
