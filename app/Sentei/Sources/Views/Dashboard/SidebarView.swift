/**
 * ダッシュボードのサイドバー
 * プラグイン (Git / RSS) をタブで切り替え、選択中プラグインの絞り込み
 * (urgency / category 等) だけを表示する。プラグイン追加時はタブを増やすだけで
 * スクロール量が肥大化しない。最下段にデーモンステータス。
 */
import SwiftUI

/// サイドバーの List
struct SidebarView: View {
    @Bindable var viewModel: AppViewModel
    @Binding var selection: DashboardSelection

    var body: some View {
        VStack(spacing: 0) {
            Picker("plugin", selection: pluginScopeBinding) {
                ForEach(PluginScope.allCases, id: \.self) { scope in
                    Text(scope.label).tag(scope)
                }
            }
            .pickerStyle(.segmented)
            .labelsHidden()
            .padding(.horizontal, 12)
            .padding(.top, 12)
            .padding(.bottom, 8)

            List(selection: $selection) {
                switch selection.pluginScope {
                case .git: gitSection
                case .rss: rssSection
                }

                Section("ステータス") {
                    statusBlock
                }
            }
            .listStyle(.sidebar)
            .scrollContentBackground(.hidden)
        }
        .background(SenteiTheme.backgroundSecondary)
    }

    /// Picker ⇄ selection の橋渡し。タブ切替時はそのプラグインの「すべて」に戻す。
    private var pluginScopeBinding: Binding<PluginScope> {
        Binding(
            get: { selection.pluginScope },
            set: { newScope in
                guard newScope != selection.pluginScope else { return }
                selection = DashboardSelection.defaultSelection(for: newScope)
            }
        )
    }

    // MARK: - Sections

    private var gitSection: some View {
        Section("Git") {
            NavigationLink(value: DashboardSelection.gitAll) {
                itemRow(
                    label: "すべて",
                    count: gitItems.count,
                    color: SenteiTheme.textSecondary
                )
            }
            ForEach(Urgency.allCases, id: \.self) { urgency in
                NavigationLink(value: DashboardSelection.gitUrgency(urgency)) {
                    itemRow(
                        label: urgency.rawValue,
                        count: gitItems.filter { $0.label.urgency == urgency }.count,
                        color: SenteiTheme.urgencyColor(urgency)
                    )
                }
            }

            if !viewModel.repos.isEmpty {
                ForEach(viewModel.repos, id: \.self) { repo in
                    NavigationLink(value: DashboardSelection.gitSummary(repo: repo)) {
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
    }

    private var rssSection: some View {
        Section("RSS") {
            NavigationLink(value: DashboardSelection.rssAll) {
                itemRow(
                    label: "すべて",
                    count: rssItems.count,
                    color: SenteiTheme.textSecondary
                )
            }
            ForEach(RssCategory.allCases, id: \.self) { category in
                NavigationLink(value: DashboardSelection.rssCategory(category)) {
                    itemRow(
                        label: category.rawValue,
                        count: rssItems.filter { $0.label.category == category.rawValue }.count,
                        color: SenteiTheme.textTertiary
                    )
                }
            }
        }
    }

    // MARK: - Computed

    private var gitItems: [SenteiItem] {
        viewModel.items.filter { $0.item.source == "git" }
    }

    private var rssItems: [SenteiItem] {
        viewModel.items.filter { $0.item.source == "rss" }
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
