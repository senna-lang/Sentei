/**
 * アプリ全体の状態管理
 * 動的ポーリング: バックグラウンド 60 秒 / フォアグラウンド 15 秒 / 表示時即時 fetch
 */
import Foundation
import SwiftUI

/// アプリの接続状態
enum ConnectionState {
    case connected
    case disconnected
    case connecting
}

/// アプリ全体の状態を管理する ViewModel
@Observable
final class AppViewModel {
    // MARK: - Published State

    var items: [SenteiItem] = []
    var summaries: [Summary] = []
    var status: DaemonStatus?
    var connectionState: ConnectionState = .connecting

    /// UI が表示中か（ポップオーバーまたはウィンドウ）
    var isUIVisible: Bool = false {
        didSet {
            if isUIVisible {
                // 表示時に即時 fetch + ポーリング間隔を短縮
                Task { await refresh() }
            }
            restartPolling()
        }
    }

    // MARK: - Computed

    var urgentCount: Int {
        items.filter { $0.label.urgency == .urgent }.count
    }

    var recentItems: [SenteiItem] {
        Array(items.prefix(10))
    }

    var repos: [String] {
        summaries.map(\.repo).sorted()
    }

    // MARK: - Private

    private let apiClient = APIClient()
    private let daemonManager = DaemonManager()
    private let notificationService = NotificationService()
    private var pollingTask: Task<Void, Never>?

    // MARK: - Lifecycle

    /// アプリ起動時に呼ばれる初期化処理
    func start() async {
        notificationService.requestPermission()
        connectionState = .connecting

        await daemonManager.ensureRunning(apiClient: apiClient)
        await refresh()
        restartPolling()
    }

    /// アプリ終了時にデーモンを停止する
    func shutdown() {
        pollingTask?.cancel()
        daemonManager.stop()
    }

    // MARK: - Data Operations

    /// 最新データを取得する
    /// items は「自分宛の通知」だけを保持する（survey 由来のレポジトリ全体活動はサマリー側で扱う）
    func refresh() async {
        do {
            async let fetchedItems = apiClient.fetchItems()
            async let fetchedStatus = apiClient.fetchStatus()

            let raw = try await fetchedItems
            items = raw.filter { $0.surveyType == nil }
            status = try await fetchedStatus
            connectionState = .connected

            // urgent 通知の処理
            notificationService.processItems(items)
        } catch {
            connectionState = .disconnected
        }
    }

    /// サマリーデータを取得する
    func refreshSummaries() async {
        do {
            summaries = try await apiClient.fetchSummaries()
        } catch {
            // サマリーの取得失敗はサイレントに処理
        }
    }

    /// アイテムを削除する（チェックボタン）
    func deleteItem(_ item: SenteiItem) async {
        do {
            try await apiClient.deleteItem(source: item.item.source, sourceID: item.item.sourceID)
            items.removeAll { $0.id == item.id }
        } catch {
            // 削除失敗時は次のポーリングで同期される
        }
    }

    /// フィルタ付きでアイテムを取得する
    /// items は「自分宛の通知」だけを保持する
    func fetchFiltered(urgency: String? = nil, source: String? = nil, category: String? = nil) async {
        do {
            let raw = try await apiClient.fetchItems(urgency: urgency, source: source, category: category)
            items = raw.filter { $0.surveyType == nil }
            connectionState = .connected
        } catch {
            connectionState = .disconnected
        }
    }

    // MARK: - Polling

    private func restartPolling() {
        pollingTask?.cancel()
        pollingTask = Task {
            let interval: Duration = isUIVisible ? .seconds(15) : .seconds(60)
            while !Task.isCancelled {
                try? await Task.sleep(for: interval)
                guard !Task.isCancelled else { break }
                await refresh()
            }
        }
    }
}
