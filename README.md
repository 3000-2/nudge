# nudge

A CLI reminder tool for macOS. Schedule one-time or recurring reminders with native notifications, powered by `launchd`.

Designed to be used by humans and AI agents (Claude Code, Codex CLI) alike.

## Features

- **Flexible scheduling** — daily, weekly, monthly, specific dates, relative dates
- **Native macOS notifications** with sound via `osascript`
- **Missed reminder handling** — `launchd` fires missed reminders when your Mac wakes from sleep
- **One-time auto-cleanup** — completed reminders are automatically unloaded
- **Machine-readable output** — `--json` flag for agent/script consumption
- **Zero dependencies** — single static Go binary, no CGo

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/3000-2/nudge/main/install.sh | sh
```

Or with Go:

```bash
go install github.com/3000-2/nudge@latest
```

Or build from source:

```bash
git clone https://github.com/3000-2/nudge.git
cd nudge
go build -ldflags "-X main.version=0.1.0" -o nudge .
sudo cp nudge /usr/local/bin/
```

## Usage

```bash
# One-time reminders
nudge add "코드 리뷰" --at 14:00                      # today (or tomorrow if past)
nudge add "미팅" --on 2026-05-12 --at 15:00            # specific date
nudge add "회의" --next wednesday --at 14:00            # next occurrence of a weekday

# Recurring reminders
nudge add "데일리 스탠드업" --every day --at 09:30      # daily
nudge add "주간 보고" --every monday --at 09:00         # weekly
nudge add "월간 정산" --every 1st --at 18:00            # monthly

# Management
nudge list                    # show active reminders
nudge list --all              # include completed
nudge list --json             # JSON output
nudge show <id>               # show detail
nudge delete <id>             # remove reminder + unload from launchd

# JSON output (for agents)
nudge add "test" --at 14:00 --json
```

## How It Works

1. `nudge add` saves the reminder to `~/.nudge/reminders.json` and creates a `launchd` plist in `~/Library/LaunchAgents/`
2. At the scheduled time, `launchd` runs `nudge notify <id>`, which shows a native macOS notification
3. **One-time** reminders are automatically marked as completed and their plist is removed
4. **Recurring** reminders stay active until you `nudge delete` them
5. If your Mac was asleep, `launchd` fires missed reminders on wake

## Scheduling Flags

| Flag | Example | Type |
|------|---------|------|
| `--at HH:MM` | `--at 09:30` | Required. 24-hour format. |
| `--on YYYY-MM-DD` | `--on 2026-05-12` | One-time on specific date. |
| `--every day` | `--every day` | Daily recurring. |
| `--every <weekday>` | `--every monday` | Weekly recurring. Accepts full names and abbreviations (mon, tue, ...). |
| `--every <ordinal>` | `--every 1st` | Monthly recurring on the Nth day. |
| `--next <weekday>` | `--next wednesday` | One-time on the next occurrence of that weekday. |

> `--on`, `--every`, and `--next` are mutually exclusive.

## Agent Integration

nudge is designed for AI agents to register reminders on behalf of users:

```bash
# Agent discovers usage
nudge add --help

# Agent creates a reminder with JSON output
nudge add "Deploy to production" --on 2026-04-15 --at 10:00 --json

# Agent checks existing reminders
nudge list --json

# Agent removes a reminder
nudge delete abc12345 --json
```

## Data Storage

| Path | Purpose |
|------|---------|
| `~/.nudge/reminders.json` | Reminder database |
| `~/.nudge/logs/` | launchd stdout/stderr logs |
| `~/Library/LaunchAgents/com.nudge.*.plist` | launchd job definitions |

## License

MIT
