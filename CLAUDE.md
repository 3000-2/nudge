# nudge — macOS CLI Reminder Tool

## Overview
Go CLI that manages reminders using launchd scheduling + osascript native notifications.
Single binary, no CGo, macOS only.

## Quick Reference

```bash
# Build
go build -ldflags "-X main.version=0.1.0" -o nudge .

# Test
go test ./...

# Test with coverage
go test -coverprofile=cover.out ./... && go tool cover -func=cover.out

# Run
./nudge add "test" --at 14:00
./nudge list
./nudge delete <id>
```

## Architecture

```
cmd/           — Cobra commands (add, list, show, delete, notify, version)
internal/
  model/       — Reminder, Schedule structs
  store/       — JSON file storage (~/.nudge/) with flock locking + atomic writes
  parser/      — CLI flag parsing (--at/--on/--every/--next → Schedule)
  scheduler/   — launchd plist generation + launchctl load/unload
  notifier/    — osascript notification via argv (no string interpolation)
  idgen/       — 8-char alphanumeric ID (crypto/rand)
  output/      — Human-readable + JSON formatting
```

## Key Design Decisions

- **launchd over cron**: Handles missed reminders on wake, macOS native
- **osascript argv**: Message passed as `on run argv` argument, NOT interpolated into script (prevents injection)
- **Atomic writes**: Store and plist both use temp file → `os.Rename`
- **File locking**: `syscall.Flock` on `~/.nudge/.lock` for concurrent safety
- **Year-less plist**: `StartCalendarInterval` has no Year field — `notify` command checks date match and runs `cleanupOnceReminder` for stale plists
- **Dependency injection**: Package-level vars in `cmd/root.go` for testability

## Runtime Paths

| Path | Purpose |
|------|---------|
| `~/.nudge/reminders.json` | Reminder database |
| `~/.nudge/logs/` | launchd stdout/stderr per reminder |
| `~/.nudge/.lock` | File lock for concurrent access |
| `~/Library/LaunchAgents/com.nudge.<id>.plist` | launchd job per reminder |

## Conventions

- **Module**: `github.com/3000-2/nudge`
- **External dependency**: cobra only
- **Test style**: Table-driven, `t.TempDir()` for isolation, package-level var override for DI
- **Error messages**: Lowercase, no period
- **Schedule types**: `once`, `daily`, `weekly`, `monthly`
- **Status values**: `active`, `completed`

## Coding Rules

- No CGo — pure Go only
- No shell execution for osascript — always `exec.Command` with argv
- Escape XML entities in plist templates via `escapeXML`
- All store operations must acquire flock before read-modify-write
- One-time reminders must auto-cleanup (unload plist + mark completed) after notify
