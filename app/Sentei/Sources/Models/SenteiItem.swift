/**
 * sentei REST API のアイテムモデル
 * Go 側の plugin.LabeledItem に対応する Codable 構造体
 */
import Foundation

/// Bonsai がラベリングした urgency レベル
enum Urgency: String, Codable, CaseIterable {
    case urgent = "urgent"
    case shouldCheck = "should_check"
    case canWait = "can_wait"
    case ignore = "ignore"

    /// urgency の表示順（低い値ほど優先）
    var sortOrder: Int {
        switch self {
        case .urgent: return 0
        case .shouldCheck: return 1
        case .canWait: return 2
        case .ignore: return 3
        }
    }
}

/// プラグインからの正規化されたアイテム
struct SenteiItemData: Codable, Hashable {
    let source: String
    let sourceID: String
    let title: String
    let content: String?
    let url: String?
    let timestamp: Date
    let metadata: [String: String]?

    enum CodingKeys: String, CodingKey {
        case source = "Source"
        case sourceID = "SourceID"
        case title = "Title"
        case content = "Content"
        case url = "URL"
        case timestamp = "Timestamp"
        case metadata = "Metadata"
    }
}

/// Bonsai によるラベリング結果
///
/// urgency は optional。RSS プラグインのように urgency 分類をしないプラグイン
/// (Go 側が空文字列を返す) を受けるため、decode 時に空文字列 / 既知外の値は nil に落とす。
struct Label: Codable, Hashable {
    let urgency: Urgency?
    let category: String
    let summary: String?

    enum CodingKeys: String, CodingKey {
        case urgency = "Urgency"
        case category = "Category"
        case summary = "Summary"
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        let raw = try container.decodeIfPresent(String.self, forKey: .urgency) ?? ""
        self.urgency = Urgency(rawValue: raw)
        self.category = try container.decode(String.self, forKey: .category)
        self.summary = try container.decodeIfPresent(String.self, forKey: .summary)
    }

    init(urgency: Urgency?, category: String, summary: String?) {
        self.urgency = urgency
        self.category = category
        self.summary = summary
    }
}

/// ラベリング済みアイテム（API レスポンスの 1 要素）
struct SenteiItem: Codable, Identifiable, Hashable {
    let item: SenteiItemData
    let label: Label
    let labeledAt: Date?

    enum CodingKeys: String, CodingKey {
        case item = "Item"
        case label = "Label"
        case labeledAt = "LabeledAt"
    }

    /// Identifiable 用 ID（source + sourceID で一意）
    var id: String { "\(item.source)/\(item.sourceID)" }

    /// 表示用の author（metadata から取得）
    var author: String? { item.metadata?["author"] }

    /// 表示用の repo（metadata から取得）
    var repo: String? { item.metadata?["repo"] }

    /// survey プラグイン由来のアイテム種別（merged_pr / new_issue など）
    var surveyType: String? { item.metadata?["survey_type"] }

    /// マージ済み PR アイテムか（ダッシュボードでのデフォルト非表示対象）
    var isMergedPR: Bool { surveyType == "merged_pr" }

    /// GitHub 通知（メンション・レビュー依頼など）由来か
    /// survey_type がなく notification_type があるものを通知扱いとする
    var isNotification: Bool {
        surveyType == nil && item.metadata?["notification_type"] != nil
    }
}
