package idgen_test

import (
	"regexp"
	"testing"

	"github.com/3000-2/nudge/internal/idgen"
)

func TestGenerateUniqueReturnsEightCharacterID(t *testing.T) {
	got, err := idgen.GenerateUnique(func(string) (bool, error) {
		return false, nil
	})
	if err != nil {
		t.Fatalf("GenerateUnique returned error: %v", err)
	}

	if !regexp.MustCompile(`^[a-z0-9]{8}$`).MatchString(got) {
		t.Fatalf("expected 8-char alphanumeric id, got %q", got)
	}
}

func TestGenerateUniqueRetriesWhenExistsReportsCollision(t *testing.T) {
	calls := 0

	_, err := idgen.GenerateUnique(func(string) (bool, error) {
		calls++
		return calls == 1, nil
	})
	if err != nil {
		t.Fatalf("GenerateUnique returned error: %v", err)
	}

	if calls < 2 {
		t.Fatalf("expected collision retry, got %d calls", calls)
	}
}
