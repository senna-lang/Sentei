/**
 * ダッシュボードのサイドバー選択状態
 * Git / RSS の 2 プラグイン単位でタブのように切り替える。
 * Git は urgency / リポジトリ別サマリーで絞る。RSS は category で絞る。
 */
import Foundation

/// RSS plugin の category enum (Go 側 grammar.go と同期)
enum RssCategory: String, CaseIterable, Hashable {
    case llmResearch = "llm_research"
    case llmNews = "llm_news"
    case devTools = "dev_tools"
    case swe = "swe"
    case other = "other"
}

/// サイドバーでの選択項目
enum DashboardSelection: Hashable {
    /// Git アイテム全体 (フィルタなし)
    case gitAll
    /// Git アイテムを urgency でフィルタ
    case gitUrgency(Urgency)
    /// リポジトリ別サマリー (Git プラグイン由来)
    case gitSummary(repo: String)

    /// RSS アイテム全体 (フィルタなし)
    case rssAll
    /// RSS アイテムを category でフィルタ
    case rssCategory(RssCategory)

    /// サイドバーの表示ラベル
    var label: String {
        switch self {
        case .gitAll: return "git / すべて"
        case .gitUrgency(let u): return "git / \(u.rawValue)"
        case .gitSummary(let repo): return "git / \(repo)"
        case .rssAll: return "rss / すべて"
        case .rssCategory(let c): return "rss / \(c.rawValue)"
        }
    }
}

/// @SceneStorage に保存できるよう RawRepresentable に適合させる
extension DashboardSelection: RawRepresentable {
    init?(rawValue: String) {
        if rawValue == "git:all" {
            self = .gitAll
            return
        }
        if rawValue == "rss:all" {
            self = .rssAll
            return
        }
        if let u = rawValue.removingPrefix("git:urgency:").flatMap(Urgency.init(rawValue:)) {
            self = .gitUrgency(u)
            return
        }
        if let repo = rawValue.removingPrefix("git:summary:") {
            self = .gitSummary(repo: repo)
            return
        }
        if let c = rawValue.removingPrefix("rss:category:").flatMap(RssCategory.init(rawValue:)) {
            self = .rssCategory(c)
            return
        }
        return nil
    }

    var rawValue: String {
        switch self {
        case .gitAll: return "git:all"
        case .gitUrgency(let u): return "git:urgency:\(u.rawValue)"
        case .gitSummary(let repo): return "git:summary:\(repo)"
        case .rssAll: return "rss:all"
        case .rssCategory(let c): return "rss:category:\(c.rawValue)"
        }
    }
}

private extension String {
    /// 指定プレフィックスを取り除いた残りを返す（無ければ nil）
    func removingPrefix(_ prefix: String) -> String? {
        guard hasPrefix(prefix) else { return nil }
        return String(dropFirst(prefix.count))
    }
}
