package notifier

import (
	"errors"
	"testing"
)

func TestNotifyPassesMessageAsArgv(t *testing.T) {
	original := runCommand
	defer func() { runCommand = original }()

	var gotName string
	var gotArgs []string
	runCommand = func(name string, args ...string) error {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return nil
	}

	message := `He said "hello" & goodbye`
	if err := Notify(message); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}

	if gotName != "osascript" {
		t.Fatalf("expected osascript, got %q", gotName)
	}
	// Should be: -e <script> <message>
	if len(gotArgs) != 3 || gotArgs[0] != "-e" {
		t.Fatalf("expected [-e, <script>, <message>], got %#v", gotArgs)
	}
	// Last arg should be the raw message — no escaping needed with argv approach
	if gotArgs[2] != message {
		t.Fatalf("expected raw message %q as last arg, got %q", message, gotArgs[2])
	}
}

func TestNotifyReturnsRunnerError(t *testing.T) {
	original := runCommand
	defer func() { runCommand = original }()

	runCommand = func(string, ...string) error {
		return errors.New("boom")
	}

	if err := Notify("test"); err == nil {
		t.Fatal("expected error")
	}
}
