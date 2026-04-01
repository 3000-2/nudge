package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/3000-2/nudge/internal/model"
	"github.com/3000-2/nudge/internal/store"
)

type commandTestEnv struct {
	t             *testing.T
	storeDir      string
	agentsDir     string
	now           time.Time
	generatedID   string
	notifications []string
	loaded        []string
	unloaded      []string
	removed       []string
	written       map[string][]byte
}

func setupCommandTestEnv(t *testing.T, now time.Time) *commandTestEnv {
	t.Helper()

	env := &commandTestEnv{
		t:           t,
		storeDir:    t.TempDir(),
		now:         now,
		generatedID: "testid01",
		written:     map[string][]byte{},
	}
	env.agentsDir = filepath.Join(env.storeDir, "LaunchAgents")

	originalNowFunc := nowFunc
	originalNewStore := newStore
	originalGenerateID := generateID
	originalExecutablePath := executablePath
	originalLaunchAgentsDir := launchAgentsDir
	originalRenderPlist := renderPlist
	originalWritePlist := writePlist
	originalLoadPlist := loadPlist
	originalUnloadPlist := unloadPlist
	originalNotifyReminder := notifyReminder
	originalRemoveFile := removeFile

	nowFunc = func() time.Time { return env.now }
	newStore = func() (*store.Store, error) { return store.New(env.storeDir) }
	generateID = func(func(string) (bool, error)) (string, error) { return env.generatedID, nil }
	executablePath = func() (string, error) { return "/usr/local/bin/nudge", nil }
	launchAgentsDir = func() (string, error) { return env.agentsDir, nil }
	renderPlist = func(model.Reminder, string, string) ([]byte, error) { return []byte("<plist/>"), nil }
	writePlist = func(path string, content []byte) error {
		env.written[path] = append([]byte(nil), content...)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return os.WriteFile(path, content, 0o644)
	}
	loadPlist = func(path string) error {
		env.loaded = append(env.loaded, path)
		return nil
	}
	unloadPlist = func(path string) error {
		env.unloaded = append(env.unloaded, path)
		return nil
	}
	notifyReminder = func(message string) error {
		env.notifications = append(env.notifications, message)
		return nil
	}
	removeFile = func(path string) error {
		env.removed = append(env.removed, path)
		return os.Remove(path)
	}

	t.Cleanup(func() {
		nowFunc = originalNowFunc
		newStore = originalNewStore
		generateID = originalGenerateID
		executablePath = originalExecutablePath
		launchAgentsDir = originalLaunchAgentsDir
		renderPlist = originalRenderPlist
		writePlist = originalWritePlist
		loadPlist = originalLoadPlist
		unloadPlist = originalUnloadPlist
		notifyReminder = originalNotifyReminder
		removeFile = originalRemoveFile
	})

	return env
}

func (env *commandTestEnv) store() *store.Store {
	env.t.Helper()

	st, err := store.New(env.storeDir)
	if err != nil {
		env.t.Fatalf("store.New returned error: %v", err)
	}
	return st
}

func (env *commandTestEnv) seedReminder(reminder model.Reminder) {
	env.t.Helper()

	if err := env.store().Add(reminder); err != nil {
		env.t.Fatalf("Add returned error: %v", err)
	}
}

func (env *commandTestEnv) writePlistFile(path string) {
	env.t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		env.t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(path, []byte("plist"), 0o644); err != nil {
		env.t.Fatalf("WriteFile returned error: %v", err)
	}
}

func executeCommandResult(t *testing.T, version string, args ...string) (string, string, error) {
	t.Helper()

	root := NewRootCmd(version)
	var out bytes.Buffer
	var errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs(args)

	err := root.Execute()
	return out.String(), errBuf.String(), err
}
