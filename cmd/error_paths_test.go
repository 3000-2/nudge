package cmd

import (
	"strings"
	"testing"
)

func TestListEmptyAndShowMissing(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	restore := stubCommandDeps(t)
	defer restore()

	output := executeCommand(t, "test", "list")
	if strings.TrimSpace(output) != "No reminders." {
		t.Fatalf("expected empty list output, got %q", output)
	}

	root := NewRootCmd("test")
	root.SetArgs([]string{"show", "missing"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), `reminder "missing" not found`) {
		t.Fatalf("expected not found error, got %v", err)
	}
}
