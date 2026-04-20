/**
 * sentei REST API クライアント
 * localhost:7890 の sentei デーモンと通信する
 */
import Foundation

/// ISO8601 風文字列の fractional seconds を ISO8601DateFormatter (3 桁固定) に合わせて
/// 3 桁に切り揃える。3 桁未満なら 0 埋め、3 桁超なら先頭 3 桁のみ残す。
/// fractional が無い場合は入力をそのまま返す。
func normalizeFractionalSeconds(_ s: String) -> String {
    guard let dotIndex = s.firstIndex(of: ".") else { return s }
    // ISO8601 の timezone suffix 境界を探す (Z または +/- で始まる hh:mm)
    let afterDot = s.index(after: dotIndex)
    var endOfFractional = afterDot
    while endOfFractional < s.endIndex, s[endOfFractional].isNumber {
        endOfFractional = s.index(after: endOfFractional)
    }
    let fractional = String(s[afterDot..<endOfFractional])
    if fractional.isEmpty { return s }

    var normalized: String
    if fractional.count >= 3 {
        normalized = String(fractional.prefix(3))
    } else {
        normalized = fractional + String(repeating: "0", count: 3 - fractional.count)
    }

    return String(s[..<afterDot]) + normalized + String(s[endOfFractional...])
}

/// API 通信エラー
enum APIError: Error, LocalizedError {
    case connectionFailed
    case invalidResponse(Int)
    case decodingFailed(Error)

    var errorDescription: String? {
        switch self {
        case .connectionFailed:
            return "デーモンに接続できません"
        case .invalidResponse(let code):
            return "API エラー: HTTP \(code)"
        case .decodingFailed(let error):
            return "レスポンスのパース失敗: \(error.localizedDescription)"
        }
    }
}

/// sentei REST API クライアント
actor APIClient {
    private let baseURL: URL
    private let session: URLSession
    private let decoder: JSONDecoder

    init(baseURL: URL = URL(string: "http://127.0.0.1:7890")!) {
        self.baseURL = baseURL
        self.session = URLSession(configuration: .default)

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let dateString = try container.decode(String.self)

            // Go の time.Time (RFC3339Nano) は fractional seconds が 0〜9 桁の可変精度で
            // 出る (例: "2026-04-19T23:02:56.70401Z" や "2026-04-19T23:02:56Z")。
            // ISO8601DateFormatter の withFractionalSeconds は 3 桁固定のため、
            // 可変精度に対応するには先に fractional を 3 桁に丸めるか除去する。
            let normalized = normalizeFractionalSeconds(dateString)

            let formatters: [ISO8601DateFormatter] = {
                let f1 = ISO8601DateFormatter()
                f1.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
                let f2 = ISO8601DateFormatter()
                f2.formatOptions = [.withInternetDateTime]
                return [f1, f2]
            }()

            for formatter in formatters {
                if let date = formatter.date(from: normalized) {
                    return date
                }
                if let date = formatter.date(from: dateString) {
                    return date
                }
            }
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "日付パース失敗: \(dateString)")
        }
        self.decoder = decoder
    }

    /// アイテム一覧を取得する
    func fetchItems(urgency: String? = nil, source: String? = nil, category: String? = nil) async throws -> [SenteiItem] {
        var components = URLComponents(url: baseURL.appendingPathComponent("/api/items"), resolvingAgainstBaseURL: false)!
        var queryItems: [URLQueryItem] = []
        if let urgency { queryItems.append(URLQueryItem(name: "urgency", value: urgency)) }
        if let source { queryItems.append(URLQueryItem(name: "source", value: source)) }
        if let category { queryItems.append(URLQueryItem(name: "category", value: category)) }
        if !queryItems.isEmpty { components.queryItems = queryItems }

        return try await get(url: components.url!, as: [SenteiItem].self)
    }

    /// サマリー一覧を取得する
    func fetchSummaries() async throws -> [Summary] {
        let url = baseURL.appendingPathComponent("/api/summary")
        return try await get(url: url, as: [Summary].self)
    }

    /// ステータスを取得する
    func fetchStatus() async throws -> DaemonStatus {
        let url = baseURL.appendingPathComponent("/api/status")
        return try await get(url: url, as: DaemonStatus.self)
    }

    /// アイテムを削除する
    func deleteItem(source: String, sourceID: String) async throws {
        let url = baseURL.appendingPathComponent("/api/items/\(source)/\(sourceID)")
        var request = URLRequest(url: url)
        request.httpMethod = "DELETE"

        let (_, response) = try await session.data(for: request)
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.connectionFailed
        }
        if httpResponse.statusCode != 200 {
            throw APIError.invalidResponse(httpResponse.statusCode)
        }
    }

    /// デーモンが稼働中か確認する
    func isAlive() async -> Bool {
        do {
            _ = try await fetchStatus()
            return true
        } catch {
            return false
        }
    }

    // MARK: - Private

    private func get<T: Decodable>(url: URL, as type: T.Type) async throws -> T {
        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(from: url)
        } catch {
            throw APIError.connectionFailed
        }

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.connectionFailed
        }
        if httpResponse.statusCode != 200 {
            throw APIError.invalidResponse(httpResponse.statusCode)
        }

        do {
            return try decoder.decode(type, from: data)
        } catch {
            throw APIError.decodingFailed(error)
        }
    }
}
