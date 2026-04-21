/**
 * macOS UserNotifications による GitHub 通知転送
 * GitHub Notifications API 由来のアイテム (mention / review_requested / comment 等) が
 * 新着で検出されたときにデスクトップ通知を送る。Gitify に近い挙動。
 * urgency は分類のためのメタ情報であり通知発火の条件ではない。
 */
import Foundation
import UserNotifications

/// GitHub 通知アイテムの macOS 通知を管理する
final class NotificationService {
    private var knownNotifiedIDs: Set<String> = []
    /// アプリ起動直後の seed 完了フラグ。起動時点で既に溜まっていた古い通知で
    /// macOS 通知がバースト発火するのを防ぐため、初回 processItems は通知せず seed だけ行う。
    private var hasSeeded = false

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

    /// 新しい GitHub 通知アイテムがあれば macOS 通知を送る
    func processItems(_ items: [SenteiItem]) {
        let notifications = items.filter(\.isNotification)

        if !hasSeeded {
            // 起動時に daemon に溜まっていた未読分は「起動前から知っていた」扱い。
            // これらで通知が一斉発火するとノイズなので seed のみ行う。
            knownNotifiedIDs = Set(notifications.map(\.id))
            hasSeeded = true
            return
        }

        for item in notifications {
            if knownNotifiedIDs.contains(item.id) { continue }
            knownNotifiedIDs.insert(item.id)
            sendNotification(for: item)
        }

        // 削除されたアイテムの ID をクリーンアップ
        let currentIDs = Set(items.map(\.id))
        knownNotifiedIDs = knownNotifiedIDs.intersection(currentIDs)
    }

    private func sendNotification(for item: SenteiItem) {
        // notification_type (review_requested / mention / comment 等) があれば優先。
        // 無ければ Bonsai カテゴリにフォールバック。
        let reason = item.item.metadata?["notification_type"] ?? item.label.category
        let body = "[\(reason)] \(item.item.title)"

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
