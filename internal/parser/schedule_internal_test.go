package parser

import (
	"strings"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
)

func TestParseOnSchedule(t *testing.T) {
	now := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local)

	tests := []struct {
		name     string
		dateText string
		wantDate string
		wantErr  string
	}{
		{
			name:     "future date",
			dateText: "2026-04-02",
			wantDate: "2026-04-02",
		},
		{
			name:     "past date",
			dateText: "2026-03-31",
			wantErr:  "scheduled time must be in the future",
		},
		{
			name:     "invalid format",
			dateText: "04/02/2026",
			wantErr:  "parse --on date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOnSchedule(tt.dateText, 14, 30, now, time.Local)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseOnSchedule returned error: %v", err)
			}
			if got.Type != model.ScheduleTypeOnce {
				t.Fatalf("expected once schedule, got %q", got.Type)
			}
			if got.Date == nil || *got.Date != tt.wantDate {
				t.Fatalf("expected date %q, got %#v", tt.wantDate, got.Date)
			}
			if got.Hour != 14 || got.Minute != 30 {
				t.Fatalf("expected 14:30, got %02d:%02d", got.Hour, got.Minute)
			}
		})
	}
}

func TestParseClockInvalid(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "invalid hour", value: "25:00"},
		{name: "invalid minute", value: "12:60"},
		{name: "text", value: "abc"},
		{name: "missing minute", value: "12"},
		{name: "too many parts", value: "12:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := parseClock(tt.value); err == nil {
				t.Fatalf("expected error for %q", tt.value)
			}
		})
	}
}

func TestParseScheduleRejectsAdditionalMutuallyExclusiveFlags(t *testing.T) {
	now := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local)

	tests := []struct {
		name  string
		on    string
		every string
		next  string
	}{
		{
			name:  "on and every",
			on:    "2026-04-02",
			every: "day",
		},
		{
			name: "on and next",
			on:   "2026-04-02",
			next: "wednesday",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSchedule("09:00", tt.on, tt.every, tt.next, now)
			if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
				t.Fatalf("expected mutually exclusive error, got %v", err)
			}
		})
	}
}

func TestParseScheduleRejectsInvalidEveryValues(t *testing.T) {
	now := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.Local)

	for _, value := range []string{"invalid", "32nd", "0th"} {
		t.Run(value, func(t *testing.T) {
			_, err := ParseSchedule("09:00", "", value, "", now)
			if err == nil || !strings.Contains(err.Error(), "invalid --every value") {
				t.Fatalf("expected invalid --every error for %q, got %v", value, err)
			}
		})
	}
}
