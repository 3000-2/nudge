package idgen_test

import (
	"strings"
	"testing"

	"github.com/3000-2/nudge/internal/idgen"
)

func TestGenerateUniqueRetriesAfterSeveralCollisions(t *testing.T) {
	calls := 0

	got, err := idgen.GenerateUnique(func(string) (bool, error) {
		calls++
		return calls <= 5, nil
	})
	if err != nil {
		t.Fatalf("GenerateUnique returned error: %v", err)
	}
	if calls != 6 {
		t.Fatalf("expected 6 existence checks, got %d", calls)
	}
	if len(got) != 8 {
		t.Fatalf("expected 8-character id, got %q", got)
	}
}

func TestGenerateUniqueExhaustsAttempts(t *testing.T) {
	calls := 0

	_, err := idgen.GenerateUnique(func(string) (bool, error) {
		calls++
		return true, nil
	})
	if err == nil || !strings.Contains(err.Error(), "exhausted 256 attempts") {
		t.Fatalf("expected exhausted attempts error, got %v", err)
	}
	if calls != 256 {
		t.Fatalf("expected 256 existence checks, got %d", calls)
	}
}
