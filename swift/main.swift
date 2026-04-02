import Foundation
import UserNotifications

let center = UNUserNotificationCenter.current()

func sendNotification(title: String, body: String) {
    var done = false

    center.requestAuthorization(options: [.alert, .sound, .badge]) { granted, error in
        if let error = error {
            fputs("error: \(error.localizedDescription)\n", stderr)
            exit(1)
        }

        if !granted {
            fputs("⚠ Notification permission denied. Enable in System Settings > Notifications > Nudge\n", stderr)
        }

        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body
        content.sound = UNNotificationSound.default

        let request = UNNotificationRequest(
            identifier: UUID().uuidString,
            content: content,
            trigger: nil
        )

        center.add(request) { error in
            if let error = error {
                fputs("error: \(error.localizedDescription)\n", stderr)
                exit(1)
            }
            done = true
            CFRunLoopStop(CFRunLoopGetMain())
        }
    }

    // Run the main loop with a timeout
    let deadline = Date(timeIntervalSinceNow: 5)
    while !done && RunLoop.main.run(mode: .default, before: deadline) {
        if Date() >= deadline { break }
    }

    if !done {
        fputs("error: notification delivery timed out\n", stderr)
        exit(1)
    }
}

// Parse arguments: NudgeNotify "message" or NudgeNotify "title" "message"
let args = CommandLine.arguments.dropFirst()

switch args.count {
case 1:
    sendNotification(title: "⏰ Nudge", body: args.first!)
case 2:
    let argsArray = Array(args)
    sendNotification(title: argsArray[0], body: argsArray[1])
default:
    fputs("usage: NudgeNotify \"message\" or NudgeNotify \"title\" \"message\"\n", stderr)
    exit(1)
}
