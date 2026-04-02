# nudge — macOS CLI Reminder Tool

## Overview
Go CLI + Swift notification helper. launchd scheduling + native macOS notifications with custom icon.
macOS only, no CGo.

## Quick Reference

```bash
# Build everything (Go binary + Swift Nudge.app)
make build

# Build only Swift helper
make swift

# Test
make test

# Install (requires sudo)
make install

# Clean
make clean
```

## Architecture

```
cmd/           — Cobra commands (add, list, show, delete, notify, version)
internal/
  model/       — Reminder, Schedule structs
  store/       — JSON file storage (~/.nudge/) with flock locking + atomic writes
  parser/      — CLI flag parsing (--at/--on/--every/--next → Schedule)
  scheduler/   — launchd plist generation + launchctl load/unload
  notifier/    — Nudge.app (Swift) with osascript fallback
  idgen/       — 8-char alphanumeric ID (crypto/rand)
  output/      — Human-readable + JSON formatting
swift/
  main.swift   — UNUserNotificationCenter notification sender
  Info.plist   — App bundle metadata (LSUIElement=true)
  Assets/      — nudge.icns app icon
  build.sh     — Builds Nudge.app bundle with swiftc
```

## Key Design Decisions

- **launchd over cron**: Handles missed reminders on wake, macOS native
- **Swift Nudge.app**: Custom icon in notifications via UNUserNotificationCenter, osascript fallback when .app not found
- **Atomic writes**: Store and plist both use temp file → `os.Rename`
- **File locking**: `syscall.Flock` on `~/.nudge/.lock` for concurrent safety
- **Year-less plist**: `StartCalendarInterval` has no Year field — `notify` command checks date match and runs `cleanupOnceReminder` for stale plists
- **Dependency injection**: Package-level vars in `cmd/root.go` for testability
- **Ad-hoc codesign**: Required for UNUserNotificationCenter, no developer account needed

## Runtime Paths

| Path | Purpose |
|------|---------|
| `~/.nudge/reminders.json` | Reminder database |
| `~/.nudge/logs/` | launchd stdout/stderr per reminder |
| `~/.nudge/.lock` | File lock for concurrent access |
| `~/Library/LaunchAgents/com.nudge.<id>.plist` | launchd job per reminder |
| `/usr/local/lib/nudge/Nudge.app` | Swift notification helper (installed) |

## Notifier Resolution Order

1. `$NUDGE_NOTIFY_APP` env var
2. `<binary_dir>/../lib/nudge/Nudge.app` (standard install)
3. `<binary_dir>/Nudge.app` (development)
4. `/usr/local/lib/nudge/Nudge.app` (hardcoded fallback)
5. osascript fallback (no custom icon)

## Conventions

- **Module**: `github.com/3000-2/nudge`
- **External dependency**: cobra only (Go), UserNotifications framework (Swift)
- **Test style**: Table-driven, `t.TempDir()` for isolation, package-level var override for DI
- **Error messages**: Lowercase, no period
- **Schedule types**: `once`, `daily`, `weekly`, `monthly`
- **Status values**: `active`, `completed`

## Coding Rules

- No CGo — pure Go only
- Notifier: prefer Nudge.app, fallback to osascript
- Escape XML entities in plist templates via `escapeXML`
- All store operations must acquire flock before read-modify-write
- One-time reminders must auto-cleanup (unload plist + mark completed) after notify
- Swift .app must be ad-hoc signed (`codesign --force --sign -`)
