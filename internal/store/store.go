package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/3000-2/nudge/internal/model"
)

const storageVersion = 1

var ErrNotFound = errors.New("reminder not found")

type Store struct {
	baseDir  string
	filePath string
	lockPath string
	logsDir  string
}

func NewDefault() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	return New(filepath.Join(home, ".nudge"))
}

func New(baseDir string) (*Store, error) {
	s := &Store{
		baseDir:  baseDir,
		filePath: filepath.Join(baseDir, "reminders.json"),
		lockPath: filepath.Join(baseDir, ".lock"),
		logsDir:  filepath.Join(baseDir, "logs"),
	}
	if err := s.ensureDirs(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) FilePath() string {
	return s.filePath
}

func (s *Store) LogsDir() string {
	return s.logsDir
}

func (s *Store) Exists(id string) (bool, error) {
	var exists bool
	err := s.withLock(func() error {
		storage, err := s.readStorageLocked()
		if err != nil {
			return err
		}
		exists = findReminderIndex(storage.Reminders, id) >= 0
		return nil
	})
	return exists, err
}

func (s *Store) Add(reminder model.Reminder) error {
	return s.withLock(func() error {
		storage, err := s.readStorageLocked()
		if err != nil {
			return err
		}
		if findReminderIndex(storage.Reminders, reminder.ID) >= 0 {
			return fmt.Errorf("reminder %q already exists", reminder.ID)
		}

		storage.Reminders = append(storage.Reminders, cloneReminder(reminder))
		return s.writeStorageLocked(storage)
	})
}

func (s *Store) List(includeCompleted bool) ([]model.Reminder, error) {
	var reminders []model.Reminder
	err := s.withLock(func() error {
		storage, err := s.readStorageLocked()
		if err != nil {
			return err
		}

		for _, reminder := range storage.Reminders {
			if !includeCompleted && reminder.Status != model.StatusActive {
				continue
			}
			reminders = append(reminders, cloneReminder(reminder))
		}

		return nil
	})
	return reminders, err
}

func (s *Store) Get(id string) (model.Reminder, error) {
	var reminder model.Reminder
	err := s.withLock(func() error {
		storage, err := s.readStorageLocked()
		if err != nil {
			return err
		}

		index := findReminderIndex(storage.Reminders, id)
		if index < 0 {
			return ErrNotFound
		}

		reminder = cloneReminder(storage.Reminders[index])
		return nil
	})
	return reminder, err
}

func (s *Store) Update(id string, mutate func(*model.Reminder) error) (model.Reminder, error) {
	var updated model.Reminder
	err := s.withLock(func() error {
		storage, err := s.readStorageLocked()
		if err != nil {
			return err
		}

		index := findReminderIndex(storage.Reminders, id)
		if index < 0 {
			return ErrNotFound
		}

		working := cloneReminder(storage.Reminders[index])
		if err := mutate(&working); err != nil {
			return err
		}

		storage.Reminders[index] = cloneReminder(working)
		if err := s.writeStorageLocked(storage); err != nil {
			return err
		}

		updated = cloneReminder(working)
		return nil
	})
	return updated, err
}

func (s *Store) Delete(id string) (bool, error) {
	var deleted bool
	err := s.withLock(func() error {
		storage, err := s.readStorageLocked()
		if err != nil {
			return err
		}

		index := findReminderIndex(storage.Reminders, id)
		if index < 0 {
			return nil
		}

		storage.Reminders = append(storage.Reminders[:index], storage.Reminders[index+1:]...)
		deleted = true
		return s.writeStorageLocked(storage)
	})
	return deleted, err
}

func (s *Store) ensureDirs() error {
	for _, dir := range []string{s.baseDir, s.logsDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}
	return nil
}

func (s *Store) withLock(fn func() error) error {
	if err := s.ensureDirs(); err != nil {
		return err
	}

	lockFile, err := os.OpenFile(s.lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("open lock file: %w", err)
	}
	defer lockFile.Close()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("acquire store lock: %w", err)
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

	return fn()
}

func (s *Store) readStorageLocked() (model.StorageFile, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return model.StorageFile{Version: storageVersion, Reminders: []model.Reminder{}}, nil
		}
		return model.StorageFile{}, fmt.Errorf("read storage file: %w", err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return model.StorageFile{Version: storageVersion, Reminders: []model.Reminder{}}, nil
	}

	var storage model.StorageFile
	if err := json.Unmarshal(data, &storage); err != nil {
		return model.StorageFile{}, fmt.Errorf("decode storage file: %w", err)
	}

	if storage.Version == 0 {
		storage.Version = storageVersion
	}
	if storage.Reminders == nil {
		storage.Reminders = []model.Reminder{}
	}

	return storage, nil
}

func (s *Store) writeStorageLocked(storage model.StorageFile) error {
	storage.Version = storageVersion
	if storage.Reminders == nil {
		storage.Reminders = []model.Reminder{}
	}

	tempFile, err := os.CreateTemp(s.baseDir, "nudge-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp storage file: %w", err)
	}

	tempPath := tempFile.Name()
	committed := false
	defer func() {
		_ = tempFile.Close()
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(storage); err != nil {
		return fmt.Errorf("encode storage file: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("sync storage file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp storage file: %w", err)
	}
	if err := os.Chmod(tempPath, 0o600); err != nil {
		return fmt.Errorf("chmod storage file: %w", err)
	}

	if err := os.Rename(tempPath, s.filePath); err != nil {
		return fmt.Errorf("replace storage file: %w", err)
	}
	committed = true
	return nil
}

func findReminderIndex(reminders []model.Reminder, id string) int {
	for index, reminder := range reminders {
		if reminder.ID == id {
			return index
		}
	}
	return -1
}

func cloneReminder(reminder model.Reminder) model.Reminder {
	cloned := reminder
	cloned.Schedule = cloneSchedule(reminder.Schedule)
	if reminder.FiredAt != nil {
		t := *reminder.FiredAt
		cloned.FiredAt = &t
	}
	return cloned
}

func cloneSchedule(schedule model.Schedule) model.Schedule {
	cloned := schedule
	if schedule.Date != nil {
		value := *schedule.Date
		cloned.Date = &value
	}
	if schedule.Weekday != nil {
		value := *schedule.Weekday
		cloned.Weekday = &value
	}
	if schedule.DayOfMonth != nil {
		value := *schedule.DayOfMonth
		cloned.DayOfMonth = &value
	}
	return cloned
}
