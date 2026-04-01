package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMainRunsVersionCommand(t *testing.T) {
	originalArgs := os.Args
	originalStdout := os.Stdout
	originalVersion := version
	defer func() {
		os.Args = originalArgs
		os.Stdout = originalStdout
		version = originalVersion
	}()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe returned error: %v", err)
	}

	os.Args = []string{"nudge", "version"}
	os.Stdout = writer
	version = "1.2.3"

	main()

	if err := writer.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if strings.TrimSpace(string(output)) != "1.2.3" {
		t.Fatalf("expected version output, got %q", string(output))
	}
}

func TestMainExitsOnCommandError(t *testing.T) {
	if os.Getenv("REMIND_MAIN_ERROR_HELPER") == "1" {
		os.Args = []string{"nudge", "show"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMainExitsOnCommandError")
	cmd.Env = append(os.Environ(), "REMIND_MAIN_ERROR_HELPER=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
	}
	if !strings.Contains(stderr.String(), `accepts 1 arg(s), received 0`) {
		t.Fatalf("expected cobra error in stderr, got %q", stderr.String())
	}
}
