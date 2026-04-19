/**
 * ダッシュボード本体
 * NavigationSplitView でサイドバーとコンテンツを切り替える
 * ウィンドウを閉じても終了せず、メニューバーから再度開ける（LSUIElement=true 前提）
 */

import SwiftUI

/// ダッシュボードウィンドウのルートビュー
struct DashboardView: View {
    @Bindable var viewModel: AppViewModel
    @SceneStorage("dashboard.selection") private var selection: DashboardSelection = .allItems

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
        case .allItems:
            ItemListView(viewModel: viewModel, fixedUrgency: nil)
        case .urgency(let u):
            ItemListView(viewModel: viewModel, fixedUrgency: u)
        case .summary(let repo):
            SummaryView(viewModel: viewModel, repo: repo)
        }
    }
}
