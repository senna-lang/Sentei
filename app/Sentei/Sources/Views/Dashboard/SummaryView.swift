/**
 * リポジトリ別のサマリー表示
 * Go 側で Render されたプレーンテキストを monospace で表示する
 */
import SwiftUI

/// サマリービュー
struct SummaryView: View {
    @Bindable var viewModel: AppViewModel
    let repo: String

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 0) {
                if let text = summaryText {
                    Text(text)
                        .font(.system(size: 12, design: .monospaced))
                        .foregroundStyle(SenteiTheme.textPrimary)
                        .textSelection(.enabled)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(16)
                } else {
                    VStack(spacing: 8) {
                        Image(systemName: "doc.text")
                            .font(.system(size: 32))
                            .foregroundStyle(SenteiTheme.textTertiary)
                        Text("\(repo) のサマリーはまだ生成されていません")
                            .font(.system(size: 12))
                            .foregroundStyle(SenteiTheme.textTertiary)
                    }
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .padding(32)
                }
            }
        }
        .background(SenteiTheme.backgroundPrimary)
        .task(id: repo) {
            await viewModel.refreshSummaries()
        }
    }

    private var summaryText: String? {
        viewModel.summaries.first(where: { $0.repo == repo })?.summary
    }
}
