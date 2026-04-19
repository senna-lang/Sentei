/**
 * アイテム行の共有コンポーネント
 * ポップオーバーとダッシュボードの両方で使用する
 * チェックボタンで削除、行クリックで URL をブラウザで開く
 */
import SwiftUI

/// アイテムの 1 行表示（カード風）
struct ItemRow: View {
    let item: SenteiItem
    let onCheck: () -> Void

    @Environment(\.openURL) private var openURL

    var body: some View {
        HStack(spacing: 8) {
            // チェックボタン（削除）
            Button(action: onCheck) {
                Image(systemName: "checkmark.circle")
                    .font(.system(size: 14))
                    .foregroundStyle(SenteiTheme.textTertiary)
            }
            .buttonStyle(.plain)
            .help("対応済みとしてマーク")

            // Urgency バッジ
            UrgencyBadge(urgency: item.label.urgency)

            // Category アイコン
            CategoryIcon(category: item.label.category)

            // Title + Author
            VStack(alignment: .leading, spacing: 2) {
                Text(item.item.title)
                    .font(.system(size: 12))
                    .foregroundStyle(SenteiTheme.textPrimary)
                    .lineLimit(1)

                if let author = item.author {
                    Text("@\(author)")
                        .font(.system(size: 10))
                        .foregroundStyle(SenteiTheme.textTertiary)
                }
            }

            Spacer()
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 6)
        .background(SenteiTheme.cardBackground)
        .overlay(
            RoundedRectangle(cornerRadius: 6)
                .stroke(SenteiTheme.cardBorder, lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 6))
        .contentShape(Rectangle())
        .onTapGesture {
            if let urlString = item.item.url, let url = URL(string: urlString) {
                openURL(url)
            }
        }
    }
}
