/**
 * Codable モデルの JSON デコードテスト
 * Go 側の JSON レスポンスを正確にパースできることを確認する
 */
import Testing
import Foundation
@testable import Sentei

@Suite("SenteiItem JSON Decoding")
struct SenteiItemTests {

    /// Go 側の LabeledItem JSON をデコードできる
    @Test func decodeLabeledItem() throws {
        let json = """
        {
            "Item": {
                "Source": "git",
                "SourceID": "notif-123",
                "Title": "Review request from mentor",
                "Content": "review_requested: Review request",
                "URL": "https://github.com/test/pull/1",
                "Timestamp": "2026-04-17T10:00:00Z",
                "Metadata": {
                    "repo": "arxiv-compass",
                    "notification_type": "review_requested",
                    "author": "mentor"
                }
            },
            "Label": {
                "Urgency": "urgent",
                "Category": "pr",
                "Summary": "Mentor requested review"
            },
            "LabeledAt": "2026-04-17T10:00:03Z"
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        let item = try decoder.decode(SenteiItem.self, from: json)

        #expect(item.item.source == "git")
        #expect(item.item.sourceID == "notif-123")
        #expect(item.item.title == "Review request from mentor")
        #expect(item.item.url == "https://github.com/test/pull/1")
        #expect(item.label.urgency == .urgent)
        #expect(item.label.category == "pr")
        #expect(item.label.summary == "Mentor requested review")
        #expect(item.author == "mentor")
        #expect(item.repo == "arxiv-compass")
        #expect(item.id == "git/notif-123")
    }

    /// アイテム配列（GET /api/items レスポンス）をデコードできる
    @Test func decodeItemArray() throws {
        let json = """
        [
            {
                "Item": {
                    "Source": "git",
                    "SourceID": "n-1",
                    "Title": "Fix bug",
                    "Content": "",
                    "URL": "",
                    "Timestamp": "2026-04-17T10:00:00Z",
                    "Metadata": {}
                },
                "Label": {
                    "Urgency": "should_check",
                    "Category": "pr",
                    "Summary": ""
                },
                "LabeledAt": "2026-04-17T10:00:01Z"
            },
            {
                "Item": {
                    "Source": "git",
                    "SourceID": "n-2",
                    "Title": "New issue",
                    "Content": "",
                    "URL": "",
                    "Timestamp": "2026-04-17T11:00:00Z",
                    "Metadata": {}
                },
                "Label": {
                    "Urgency": "can_wait",
                    "Category": "issue",
                    "Summary": ""
                },
                "LabeledAt": "2026-04-17T11:00:01Z"
            }
        ]
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        let items = try decoder.decode([SenteiItem].self, from: json)

        #expect(items.count == 2)
        #expect(items[0].label.urgency == .shouldCheck)
        #expect(items[1].label.urgency == .canWait)
    }

    /// urgency の sortOrder が正しい
    @Test func urgencySortOrder() {
        #expect(Urgency.urgent.sortOrder < Urgency.shouldCheck.sortOrder)
        #expect(Urgency.shouldCheck.sortOrder < Urgency.canWait.sortOrder)
        #expect(Urgency.canWait.sortOrder < Urgency.ignore.sortOrder)
    }
}

@Suite("Summary JSON Decoding")
struct SummaryTests {

    /// サマリーレスポンスをデコードできる
    @Test func decodeSummaryArray() throws {
        let json = """
        [
            {"repo": "arxiv-compass", "summary": "📋 arxiv-compass\\nマージ 3件"},
            {"repo": "logosyncx", "summary": "📋 logosyncx\\n変化なし"}
        ]
        """.data(using: .utf8)!

        let summaries = try JSONDecoder().decode([Summary].self, from: json)

        #expect(summaries.count == 2)
        #expect(summaries[0].repo == "arxiv-compass")
        #expect(summaries[0].id == "arxiv-compass")
    }
}

@Suite("DaemonStatus JSON Decoding")
struct StatusTests {

    /// ステータスレスポンスをデコードできる
    @Test func decodeStatus() throws {
        let json = """
        {
            "daemon": "running",
            "bonsai": "ok",
            "plugins": ["git"],
            "item_count": 15,
            "last_labeled": "2026-04-17T12:00:00Z"
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        let status = try decoder.decode(DaemonStatus.self, from: json)

        #expect(status.isRunning == true)
        #expect(status.isBonsaiOK == true)
        #expect(status.plugins == ["git"])
        #expect(status.itemCount == 15)
    }

    /// bonsai が error の場合
    @Test func decodeStatusBonsaiError() throws {
        let json = """
        {
            "daemon": "running",
            "bonsai": "error",
            "plugins": [],
            "item_count": 0,
            "last_labeled": null
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        let status = try decoder.decode(DaemonStatus.self, from: json)

        #expect(status.isBonsaiOK == false)
        #expect(status.lastLabeled == nil)
    }
}
