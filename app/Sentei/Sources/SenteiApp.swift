/**
 * sentei macOS アプリのエントリポイント
 * MenuBarExtra でメニューバー常駐 + WindowGroup でダッシュボードウィンドウ
 * アプリがデーモンのライフサイクルを管理する（Ollama パターン）
 */
import SwiftUI

@main
struct SenteiApp: App {
    @State private var viewModel = AppViewModel()

    var body: some Scene {
        // メニューバー常駐（ポップオーバー）
        MenuBarExtra {
            PopoverView(viewModel: viewModel)
        } label: {
            MenuBarLabel(urgentCount: viewModel.urgentCount, connectionState: viewModel.connectionState)
        }
        .menuBarExtraStyle(.window)

        // ダッシュボードウィンドウ（単一インスタンス）
        // Window を使うと常に同じインスタンスになるので、openWindow 再呼出しでも複製されない
        Window("sentei Dashboard", id: "dashboard") {
            DashboardView(viewModel: viewModel)
                .task {
                    // 初回起動時にデーモン起動 + データ取得
                    await viewModel.start()
                }
        }
        .windowResizability(.contentMinSize)
        .defaultSize(width: 1100, height: 780)
    }
}

/// メニューバーアイコンの 3 状態表現
struct MenuBarLabel: View {
    let urgentCount: Int
    let connectionState: ConnectionState

    var body: some View {
        ZStack(alignment: .topTrailing) {
            Image(systemName: iconName)
                .symbolRenderingMode(.hierarchical)
                .foregroundStyle(iconColor)

            if urgentCount > 0 {
                Text("\(urgentCount)")
                    .font(.system(size: 7, weight: .bold))
                    .foregroundStyle(SenteiTheme.textPrimary)
                    .padding(.horizontal, 3)
                    .padding(.vertical, 1)
                    .background(SenteiTheme.urgentKakishibu)
                    .clipShape(Capsule())
                    .offset(x: 4, y: -4)
            }
        }
    }

    private var iconName: String {
        "leaf.fill"
    }

    private var iconColor: Color {
        switch connectionState {
        case .connected: return .primary
        case .disconnected: return SenteiTheme.textTertiary
        case .connecting: return SenteiTheme.urgencyKokeKi
        }
    }
}
