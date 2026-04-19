/**
 * ダッシュボードのサイドバー
 * urgency 別カウント / サマリー（リポジトリ別ネスト）/ デーモンステータスを表示する
 */
import SwiftUI

/// サイドバーの List
struct SidebarView: View {
    @Bindable var viewModel: AppViewModel
    @Binding var selection: DashboardSelection

    var body: some View {
        List(selection: $selection) {
            Section("アイテム") {
                NavigationLink(value: DashboardSelection.allItems) {
                    itemRow(label: "すべて", count: viewModel.items.count, color: SenteiTheme.textSecondary)
                }
                ForEach(Urgency.allCases, id: \.self) { urgency in
                    NavigationLink(value: DashboardSelection.urgency(urgency)) {
                        itemRow(
                            label: urgency.rawValue,
                            count: viewModel.items.filter { $0.label.urgency == urgency }.count,
                            color: SenteiTheme.urgencyColor(urgency)
                        )
                    }
                }
            }

            if !viewModel.repos.isEmpty {
                Section("サマリー") {
                    ForEach(viewModel.repos, id: \.self) { repo in
                        NavigationLink(value: DashboardSelection.summary(repo: repo)) {
                            HStack(spacing: 6) {
                                Image(systemName: "doc.text")
                                    .font(.system(size: 11))
                                    .foregroundStyle(SenteiTheme.textTertiary)
                                Text(repo)
                                    .font(.system(size: 12))
                                    .lineLimit(1)
                                    .truncationMode(.middle)
                            }
                        }
                    }
                }
            }

            Section("ステータス") {
                statusBlock
            }
        }
        .listStyle(.sidebar)
        .scrollContentBackground(.hidden)
        .background(SenteiTheme.backgroundSecondary)
    }

    // MARK: - Rows

    private func itemRow(label: String, count: Int, color: Color) -> some View {
        HStack(spacing: 6) {
            Circle()
                .fill(color)
                .frame(width: 8, height: 8)
            Text(label)
                .font(.system(size: 12))
            Spacer()
            Text("\(count)")
                .font(.system(size: 11, weight: .medium))
                .foregroundStyle(SenteiTheme.textTertiary)
        }
    }

    private var statusBlock: some View {
        VStack(alignment: .leading, spacing: 4) {
            statusRow(label: "daemon", ok: viewModel.connectionState == .connected)
            statusRow(label: "bonsai", ok: viewModel.status?.isBonsaiOK ?? false)
            if let count = viewModel.status?.itemCount {
                HStack {
                    Text("items")
                        .font(.system(size: 11))
                        .foregroundStyle(SenteiTheme.textTertiary)
                    Spacer()
                    Text("\(count)")
                        .font(.system(size: 11))
                        .foregroundStyle(SenteiTheme.textSecondary)
                }
            }
            if let last = viewModel.status?.lastLabeled {
                HStack {
                    Text("last")
                        .font(.system(size: 11))
                        .foregroundStyle(SenteiTheme.textTertiary)
                    Spacer()
                    Text(last, style: .relative)
                        .font(.system(size: 11))
                        .foregroundStyle(SenteiTheme.textSecondary)
                }
            }
        }
        .padding(.vertical, 4)
    }

    private func statusRow(label: String, ok: Bool) -> some View {
        HStack(spacing: 6) {
            Circle()
                .fill(ok ? SenteiTheme.accentPrimary : SenteiTheme.urgentKakishibu)
                .frame(width: 6, height: 6)
            Text(label)
                .font(.system(size: 11))
                .foregroundStyle(SenteiTheme.textSecondary)
            Spacer()
            Text(ok ? "ok" : "error")
                .font(.system(size: 11))
                .foregroundStyle(SenteiTheme.textTertiary)
        }
    }
}
