/**
 * macOS UserNotifications による urgent 通知
 * 新しい urgent アイテムが検出されたときにデスクトップ通知を送る
 */
import Foundation
import UserNotifications

/// urgent アイテムの通知を管理する
final class NotificationService {
    private var knownUrgentIDs: Set<String> = []

    /// UNUserNotificationCenter は .app bundle で起動した時のみ安全に呼べる
    /// swift run / バイナリ直叩きだと UN Center 初期化で NSException → 強制終了する
    /// bundlePath が .app で終わるかを基準にする（bundleIdentifier は swift run でも設定される）
    private static let isBundled: Bool = {
        Bundle.main.bundlePath.hasSuffix(".app")
    }()

    /// 通知権限をリクエストする
    func requestPermission() {
        guard Self.isBundled else {
            print("通知: bundle 外で実行中のためスキップ")
            return
        }
        UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound]) { _, error in
            if let error {
                print("通知権限リクエスト失敗: \(error)")
            }
        }
    }

    /// 新しい urgent アイテムがあれば通知を送る
    func processItems(_ items: [SenteiItem]) {
        let urgentItems = items.filter { $0.label.urgency == .urgent }

        for item in urgentItems {
            if knownUrgentIDs.contains(item.id) { continue }
            knownUrgentIDs.insert(item.id)
            sendNotification(for: item)
        }

        // 削除されたアイテムの ID をクリーンアップ
        let currentIDs = Set(items.map(\.id))
        knownUrgentIDs = knownUrgentIDs.intersection(currentIDs)
    }

    private func sendNotification(for item: SenteiItem) {
        let body = "[\(item.label.category)] \(item.item.title)"

        // bundle 外 (swift run 等) では UN 呼び出しは落ちるので stdout にログだけ出す。
        // 「本番なら通知が飛んでいた」内容を開発中でも観察できるようにする。
        guard Self.isBundled else {
            print("[DEV NOTIFY] \(body)")
            return
        }

        let content = UNMutableNotificationContent()
        content.title = "sentei"
        content.body = body
        content.sound = .default

        let request = UNNotificationRequest(
            identifier: item.id,
            content: content,
            trigger: nil
        )

        UNUserNotificationCenter.current().add(request)
    }
}
