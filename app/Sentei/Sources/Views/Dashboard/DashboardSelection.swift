/**
 * ダッシュボードのサイドバー選択状態
 * アイテム一覧（urgency フィルタ付き）とサマリー（リポジトリ別）を一つの enum で表す
 */
import Foundation

/// サイドバーでの選択項目
enum DashboardSelection: Hashable {
    /// 全アイテム（フィルタなし）
    case allItems
    /// urgency でフィルタしたアイテム一覧
    case urgency(Urgency)
    /// リポジトリ別のサマリー
    case summary(repo: String)

    /// サイドバーの表示ラベル
    var label: String {
        switch self {
        case .allItems: return "すべて"
        case .urgency(let u):
            switch u {
            case .urgent: return "urgent"
            case .shouldCheck: return "should_check"
            case .canWait: return "can_wait"
            case .ignore: return "ignore"
            }
        case .summary(let repo): return repo
        }
    }
}

/// @SceneStorage に保存できるよう RawRepresentable に適合させる
extension DashboardSelection: RawRepresentable {
    init?(rawValue: String) {
        if rawValue == "all" {
            self = .allItems
            return
        }
        if let u = rawValue.removingPrefix("urgency:").flatMap(Urgency.init(rawValue:)) {
            self = .urgency(u)
            return
        }
        if let repo = rawValue.removingPrefix("summary:") {
            self = .summary(repo: repo)
            return
        }
        return nil
    }

    var rawValue: String {
        switch self {
        case .allItems: return "all"
        case .urgency(let u): return "urgency:\(u.rawValue)"
        case .summary(let repo): return "summary:\(repo)"
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
