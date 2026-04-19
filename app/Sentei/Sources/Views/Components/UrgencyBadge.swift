/**
 * urgency レベルを色付きバッジで表示するコンポーネント
 */
import SwiftUI

/// urgency を色丸で表示する
struct UrgencyBadge: View {
    let urgency: Urgency

    var body: some View {
        Circle()
            .fill(SenteiTheme.urgencyColor(urgency))
            .frame(width: 8, height: 8)
    }
}
