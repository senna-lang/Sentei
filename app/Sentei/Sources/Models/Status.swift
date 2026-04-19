/**
 * デーモンステータスモデル
 * GET /api/status のレスポンスに対応する
 */
import Foundation

/// デーモンの稼働状態
struct DaemonStatus: Codable {
    let daemon: String
    let bonsai: String
    let plugins: [String]
    let itemCount: Int
    let lastLabeled: Date?

    enum CodingKeys: String, CodingKey {
        case daemon
        case bonsai
        case plugins
        case itemCount = "item_count"
        case lastLabeled = "last_labeled"
    }

    var isRunning: Bool { daemon == "running" }
    var isBonsaiOK: Bool { bonsai == "ok" }
}
