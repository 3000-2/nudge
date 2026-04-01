package scheduler

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestExecuteLaunchctlIgnoresBenignErrors(t *testing.T) {
	original := runLaunchctl
	defer func() { runLaunchctl = original }()

	runLaunchctl = func(args ...string) ([]byte, error) {
		return []byte("service already loaded"), errors.New("boom")
	}

	if err := Load("/tmp/test.plist"); err != nil {
		t.Fatalf("Load returned error for benign message: %v", err)
	}

	runLaunchctl = func(args ...string) ([]byte, error) {
		return []byte("Could not find specified service"), errors.New("boom")
	}

	if err := Unload("/tmp/test.plist"); err != nil {
		t.Fatalf("Unload returned error for benign message: %v", err)
	}
}

func TestExecuteLaunchctlReturnsDetailedError(t *testing.T) {
	original := runLaunchctl
	defer func() { runLaunchctl = original }()

	runLaunchctl = func(args ...string) ([]byte, error) {
		return []byte("permission denied"), errors.New("boom")
	}

	err := Load("/tmp/test.plist")
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected detailed error, got %v", err)
	}
}

func TestWritePlistAndHelperPaths(t *testing.T) {
	home := t.TempDir()
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("Setenv returned error: %v", err)
	}
	defer os.Setenv("HOME", originalHome)

	agentsDir, err := DefaultLaunchAgentsDir()
	if err != nil {
		t.Fatalf("DefaultLaunchAgentsDir returned error: %v", err)
	}

	plistPath := PlistPath(agentsDir, "abcd1234")
	if !strings.Contains(plistPath, "com.nudge.abcd1234.plist") {
		t.Fatalf("unexpected plist path %q", plistPath)
	}

	if err := WritePlist(plistPath, []byte("plist")); err != nil {
		t.Fatalf("WritePlist returned error: %v", err)
	}

	data, err := os.ReadFile(plistPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(data) != "plist" {
		t.Fatalf("expected plist content, got %q", string(data))
	}

	path, err := ExecutablePath()
	if err != nil {
		t.Fatalf("ExecutablePath returned error: %v", err)
	}
	if path == "" {
		t.Fatal("expected executable path")
	}
}
