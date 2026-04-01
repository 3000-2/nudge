package parser_test

import (
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
	"github.com/3000-2/nudge/internal/parser"
)

func TestParseScheduleAtOnlyUsesTodayWhenFuture(t *testing.T) {
	now := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local)

	got, err := parser.ParseSchedule("15:30", "", "", "", now)
	if err != nil {
		t.Fatalf("ParseSchedule returned error: %v", err)
	}

	wantDate := "2026-04-01"
	if got.Type != model.ScheduleTypeOnce {
		t.Fatalf("expected once schedule, got %q", got.Type)
	}
	if got.Date == nil || *got.Date != wantDate {
		t.Fatalf("expected date %q, got %#v", wantDate, got.Date)
	}
	if got.Hour != 15 || got.Minute != 30 {
		t.Fatalf("expected 15:30, got %02d:%02d", got.Hour, got.Minute)
	}
}

func TestParseScheduleAtOnlyUsesTomorrowWhenPast(t *testing.T) {
	now := time.Date(2026, time.April, 1, 18, 0, 0, 0, time.Local)

	got, err := parser.ParseSchedule("09:15", "", "", "", now)
	if err != nil {
		t.Fatalf("ParseSchedule returned error: %v", err)
	}

	wantDate := "2026-04-02"
	if got.Date == nil || *got.Date != wantDate {
		t.Fatalf("expected date %q, got %#v", wantDate, got.Date)
	}
}

func TestParseScheduleEveryWeekly(t *testing.T) {
	now := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local)

	got, err := parser.ParseSchedule("08:45", "", "monday", "", now)
	if err != nil {
		t.Fatalf("ParseSchedule returned error: %v", err)
	}

	if got.Type != model.ScheduleTypeWeekly {
		t.Fatalf("expected weekly schedule, got %q", got.Type)
	}
	if got.Weekday == nil || *got.Weekday != 1 {
		t.Fatalf("expected weekday 1, got %#v", got.Weekday)
	}
}

func TestParseScheduleEveryMonthly(t *testing.T) {
	now := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local)

	got, err := parser.ParseSchedule("08:45", "", "21st", "", now)
	if err != nil {
		t.Fatalf("ParseSchedule returned error: %v", err)
	}

	if got.Type != model.ScheduleTypeMonthly {
		t.Fatalf("expected monthly schedule, got %q", got.Type)
	}
	if got.DayOfMonth == nil || *got.DayOfMonth != 21 {
		t.Fatalf("expected day of month 21, got %#v", got.DayOfMonth)
	}
}

func TestParseScheduleNextUsesSameDayWhenTimeIsFuture(t *testing.T) {
	now := time.Date(2026, time.April, 6, 8, 0, 0, 0, time.Local)

	got, err := parser.ParseSchedule("09:00", "", "", "mon", now)
	if err != nil {
		t.Fatalf("ParseSchedule returned error: %v", err)
	}

	wantDate := "2026-04-06"
	if got.Date == nil || *got.Date != wantDate {
		t.Fatalf("expected date %q, got %#v", wantDate, got.Date)
	}
}

func TestParseScheduleNextSkipsToNextWeekWhenTimePassed(t *testing.T) {
	now := time.Date(2026, time.April, 6, 10, 0, 0, 0, time.Local)

	got, err := parser.ParseSchedule("09:00", "", "", "monday", now)
	if err != nil {
		t.Fatalf("ParseSchedule returned error: %v", err)
	}

	wantDate := "2026-04-13"
	if got.Date == nil || *got.Date != wantDate {
		t.Fatalf("expected date %q, got %#v", wantDate, got.Date)
	}
}

func TestParseScheduleRejectsMutuallyExclusiveFlags(t *testing.T) {
	now := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local)

	if _, err := parser.ParseSchedule("09:00", "2026-04-02", "day", "", now); err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
}
