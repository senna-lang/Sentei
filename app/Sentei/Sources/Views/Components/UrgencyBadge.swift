/**
 * urgency レベルを色付きバッジで表示するコンポーネント
 */
import SwiftUI

/// urgency を色丸で表示する。urgency が nil (RSS 等) の場合は淡いグレー丸。
struct UrgencyBadge: View {
    let urgency: Urgency?

    var body: some View {
        Circle()
            .fill(urgency.map(SenteiTheme.urgencyColor) ?? SenteiTheme.divider)
            .frame(width: 8, height: 8)
    }
}
