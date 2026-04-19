/**
 * サマリーモデル
 * GET /api/summary のレスポンスに対応する
 */
import Foundation

/// リポジトリ別のサマリーデータ
struct Summary: Codable, Identifiable, Hashable {
    let repo: String
    let summary: String

    var id: String { repo }
}
