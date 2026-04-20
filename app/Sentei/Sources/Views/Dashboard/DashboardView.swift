/**
 * ダッシュボード本体
 * NavigationSplitView でサイドバーとコンテンツを切り替える
 * ウィンドウを閉じても終了せず、メニューバーから再度開ける（LSUIElement=true 前提）
 */

import SwiftUI

/// ダッシュボードウィンドウのルートビュー
struct DashboardView: View {
    @Bindable var viewModel: AppViewModel
    @SceneStorage("dashboard.selection") private var selection: DashboardSelection = .gitAll

    var body: some View {
        NavigationSplitView {
            SidebarView(viewModel: viewModel, selection: $selection)
                .navigationSplitViewColumnWidth(min: 180, ideal: 220, max: 300)
        } detail: {
            detailContent
        }
        .frame(minWidth: 900, idealHeight: 760, maxHeight: .infinity)
        .frame(minHeight: 640)
        .background(SenteiTheme.backgroundPrimary)
        .preferredColorScheme(.dark)
        .task {
            viewModel.isUIVisible = true
            await viewModel.refresh()
            await viewModel.refreshSummaries()
        }
        .onDisappear {
            viewModel.isUIVisible = false
        }
    }

    @ViewBuilder
    private var detailContent: some View {
        switch selection {
        case .gitAll:
            ItemListView(viewModel: viewModel, source: "git", fixedUrgency: nil, fixedCategory: nil)
        case .gitUrgency(let u):
            ItemListView(viewModel: viewModel, source: "git", fixedUrgency: u, fixedCategory: nil)
        case .gitSummary(let repo):
            SummaryView(viewModel: viewModel, repo: repo)
        case .rssAll:
            ItemListView(viewModel: viewModel, source: "rss", fixedUrgency: nil, fixedCategory: nil)
        case .rssCategory(let c):
            ItemListView(viewModel: viewModel, source: "rss", fixedUrgency: nil, fixedCategory: c.rawValue)
        }
    }
}
