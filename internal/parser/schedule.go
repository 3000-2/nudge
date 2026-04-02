package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/3000-2/nudge/internal/model"
)

var ordinalPattern = regexp.MustCompile(`^([1-9]|[12][0-9]|3[01])(st|nd|rd|th)$`)

var weekdayNames = map[string]int{
	"sun":       0,
	"sunday":    0,
	"mon":       1,
	"monday":    1,
	"tue":       2,
	"tues":      2,
	"tuesday":   2,
	"wed":       3,
	"wednesday": 3,
	"thu":       4,
	"thur":      4,
	"thurs":     4,
	"thursday":  4,
	"fri":       5,
	"friday":    5,
	"sat":       6,
	"saturday":  6,
}

func ParseSchedule(at, on, every, next string, now time.Time) (model.Schedule, error) {
	loc := now.Location()
	hour, minute, err := parseClock(at)
	if err != nil {
		return model.Schedule{}, err
	}

	exclusiveCount := 0
	for _, value := range []string{strings.TrimSpace(on), strings.TrimSpace(every), strings.TrimSpace(next)} {
		if value != "" {
			exclusiveCount++
		}
	}
	if exclusiveCount > 1 {
		return model.Schedule{}, fmt.Errorf("--on, --every, and --next are mutually exclusive")
	}

	switch {
	case strings.TrimSpace(on) != "":
		return parseOnSchedule(strings.TrimSpace(on), hour, minute, now, loc)
	case strings.TrimSpace(every) != "":
		return parseEverySchedule(strings.TrimSpace(every), hour, minute)
	case strings.TrimSpace(next) != "":
		return parseNextSchedule(strings.TrimSpace(next), hour, minute, now, loc)
	default:
		return parseImplicitOnce(hour, minute, now, loc), nil
	}
}

func parseOnSchedule(dateText string, hour, minute int, now time.Time, loc *time.Location) (model.Schedule, error) {
	date, err := time.ParseInLocation("2006-01-02", dateText, loc)
	if err != nil {
		return model.Schedule{}, fmt.Errorf("parse --on date: %w", err)
	}

	target := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, loc)
	if !target.After(now) {
		return model.Schedule{}, fmt.Errorf("scheduled time must be in the future")
	}
	if target.After(now.AddDate(1, 0, 0)) {
		return model.Schedule{}, fmt.Errorf("scheduled date must be within 1 year")
	}

	dateValue := target.Format("2006-01-02")
	return model.Schedule{
		Type:   model.ScheduleTypeOnce,
		Hour:   hour,
		Minute: minute,
		Date:   &dateValue,
	}, nil
}

func parseEverySchedule(value string, hour, minute int) (model.Schedule, error) {
	lower := strings.ToLower(strings.TrimSpace(value))
	if lower == "day" {
		return model.Schedule{
			Type:   model.ScheduleTypeDaily,
			Hour:   hour,
			Minute: minute,
		}, nil
	}

	if weekday, ok := weekdayNames[lower]; ok {
		return model.Schedule{
			Type:    model.ScheduleTypeWeekly,
			Hour:    hour,
			Minute:  minute,
			Weekday: intPtr(weekday),
		}, nil
	}

	if matches := ordinalPattern.FindStringSubmatch(lower); matches != nil {
		day, _ := strconv.Atoi(matches[1])
		return model.Schedule{
			Type:       model.ScheduleTypeMonthly,
			Hour:       hour,
			Minute:     minute,
			DayOfMonth: intPtr(day),
		}, nil
	}

	return model.Schedule{}, fmt.Errorf("invalid --every value %q", value)
}

func parseNextSchedule(value string, hour, minute int, now time.Time, loc *time.Location) (model.Schedule, error) {
	weekday, ok := weekdayNames[strings.ToLower(strings.TrimSpace(value))]
	if !ok {
		return model.Schedule{}, fmt.Errorf("invalid --next value %q", value)
	}

	today := int(now.Weekday())
	daysAhead := (weekday - today + 7) % 7
	target := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc).AddDate(0, 0, daysAhead)
	if daysAhead == 0 && !target.After(now) {
		target = target.AddDate(0, 0, 7)
	}

	dateValue := target.Format("2006-01-02")
	return model.Schedule{
		Type:   model.ScheduleTypeOnce,
		Hour:   hour,
		Minute: minute,
		Date:   &dateValue,
	}, nil
}

func parseImplicitOnce(hour, minute int, now time.Time, loc *time.Location) model.Schedule {
	target := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	if !target.After(now) {
		target = target.AddDate(0, 0, 1)
	}

	dateValue := target.Format("2006-01-02")
	return model.Schedule{
		Type:   model.ScheduleTypeOnce,
		Hour:   hour,
		Minute: minute,
		Date:   &dateValue,
	}
}

func parseClock(value string) (int, int, error) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 || len(parts[0]) != 2 || len(parts[1]) != 2 {
		return 0, 0, fmt.Errorf("--at must use HH:MM in 24h format (e.g., 09:05)")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse --at hour: %w", err)
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse --at minute: %w", err)
	}
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("--at must use HH:MM in 24h format")
	}

	return hour, minute, nil
}

func intPtr(value int) *int {
	return &value
}
