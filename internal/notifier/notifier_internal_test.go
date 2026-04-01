package notifier

import (
	"strings"
	"testing"
)

func TestRunCommandIncludesOutputInError(t *testing.T) {
	err := runCommand("sh", "-c", "printf denied; exit 7")
	if err == nil || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("expected output to be included in error, got %v", err)
	}
}

func TestRunCommandHandlesEmptyOutputErrors(t *testing.T) {
	err := runCommand("sh", "-c", "exit 9")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), ": ") && strings.HasSuffix(err.Error(), ":") {
		t.Fatalf("unexpected formatting for empty output error: %v", err)
	}
}
