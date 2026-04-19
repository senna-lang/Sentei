/**
 * category に対応する SF Symbol アイコンを表示するコンポーネント
 */
import SwiftUI

/// category を SF Symbol で表示する
struct CategoryIcon: View {
    let category: String

    var body: some View {
        Image(systemName: SenteiTheme.categoryIcon(category))
            .font(.system(size: 11))
            .foregroundStyle(SenteiTheme.textSecondary)
            .frame(width: 16)
    }
}
