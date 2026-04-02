package notifier

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNotifyUsesSwiftAppWhenAvailable(t *testing.T) {
	origRun := runCommand
	origResolve := resolveNotifyApp
	defer func() { runCommand = origRun; resolveNotifyApp = origResolve }()

	// Create a fake Nudge.app
	appDir := filepath.Join(t.TempDir(), "Nudge.app", "Contents", "MacOS")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fakeBinary := filepath.Join(appDir, "Nudge")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	resolveNotifyApp = func() string {
		return filepath.Join(t.TempDir(), "Nudge.app")
	}
	// Re-resolve with the correct temp dir
	nudgeApp := filepath.Dir(filepath.Dir(appDir))
	resolveNotifyApp = func() string { return nudgeApp }

	var gotName string
	var gotArgs []string
	runCommand = func(name string, args ...string) error {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return nil
	}

	if err := Notify("hello"); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}

	if gotName != fakeBinary {
		t.Fatalf("expected Swift binary %q, got %q", fakeBinary, gotName)
	}
	if len(gotArgs) != 1 || gotArgs[0] != "hello" {
		t.Fatalf("expected [hello], got %#v", gotArgs)
	}
}

func TestNotifyFallsBackToOsascript(t *testing.T) {
	origRun := runCommand
	origResolve := resolveNotifyApp
	defer func() { runCommand = origRun; resolveNotifyApp = origResolve }()

	resolveNotifyApp = func() string { return "" }

	var gotName string
	var gotArgs []string
	runCommand = func(name string, args ...string) error {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return nil
	}

	if err := Notify("test"); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}

	if gotName != "osascript" {
		t.Fatalf("expected osascript fallback, got %q", gotName)
	}
	if len(gotArgs) != 3 || gotArgs[0] != "-e" {
		t.Fatalf("expected osascript args, got %#v", gotArgs)
	}
}

func TestNotifyFallsBackOnSwiftFailure(t *testing.T) {
	origRun := runCommand
	origResolve := resolveNotifyApp
	defer func() { runCommand = origRun; resolveNotifyApp = origResolve }()

	resolveNotifyApp = func() string { return "/fake/Nudge.app" }

	callCount := 0
	runCommand = func(name string, args ...string) error {
		callCount++
		if callCount == 1 {
			return errors.New("swift failed")
		}
		return nil
	}

	if err := Notify("test"); err != nil {
		t.Fatalf("expected fallback to succeed, got %v", err)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 calls (swift + osascript), got %d", callCount)
	}
}

func TestNotifyReturnsErrorWhenBothFail(t *testing.T) {
	origRun := runCommand
	origResolve := resolveNotifyApp
	defer func() { runCommand = origRun; resolveNotifyApp = origResolve }()

	resolveNotifyApp = func() string { return "/fake/Nudge.app" }
	runCommand = func(string, ...string) error { return errors.New("boom") }

	if err := Notify("test"); err == nil {
		t.Fatal("expected error")
	}
}

func TestIsValidApp(t *testing.T) {
	// Valid app
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "Test.app", "Contents", "MacOS")
	os.MkdirAll(appDir, 0o755)
	os.WriteFile(filepath.Join(appDir, "Nudge"), []byte("x"), 0o755)
	if !isValidApp(filepath.Join(tmpDir, "Test.app")) {
		t.Fatal("expected valid app")
	}

	// Missing binary
	if isValidApp("/nonexistent/Nudge.app") {
		t.Fatal("expected invalid app")
	}
}

func TestDefaultResolveNotifyAppEnvOverride(t *testing.T) {
	// Create fake app for env var test
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "Nudge.app", "Contents", "MacOS")
	os.MkdirAll(appDir, 0o755)
	os.WriteFile(filepath.Join(appDir, "Nudge"), []byte("x"), 0o755)
	envApp := filepath.Join(tmpDir, "Nudge.app")

	t.Setenv("NUDGE_NOTIFY_APP", envApp)
	got := defaultResolveNotifyApp()
	if got != envApp {
		t.Fatalf("expected %q from env, got %q", envApp, got)
	}
}
