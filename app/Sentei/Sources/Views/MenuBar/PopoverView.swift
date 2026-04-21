/**
 * メニューバーのポップオーバー
 * 接続状態ヘッダー + 直近 10 件のアイテム + フッター（ダッシュボード / 終了）
 */
import SwiftUI

/// メニューバーポップオーバーの内容
struct PopoverView: View {
    @Bindable var viewModel: AppViewModel
    @Environment(\.openWindow) private var openWindow

    var body: some View {
        VStack(spacing: 0) {
            // ヘッダー: 接続状態 + urgent カウント
            header
            Divider()

            // アイテムリスト
            itemList
            Divider()

            // フッター: ダッシュボード + 終了
            footer
        }
        .frame(width: 360)
        .background(SenteiTheme.backgroundPrimary)
        .preferredColorScheme(.dark)
        .onAppear {
            viewModel.isUIVisible = true
        }
        .onDisappear {
            viewModel.isUIVisible = false
        }
    }

    // MARK: - Header

    private var header: some View {
        HStack {
            Circle()
                .fill(connectionColor)
                .frame(width: 8, height: 8)

            Text(connectionText)
                .font(.system(size: 11))
                .foregroundStyle(SenteiTheme.textSecondary)

            Spacer()

            if viewModel.urgentCount > 0 {
                Text("urgent: \(viewModel.urgentCount)")
                    .font(.system(size: 11, weight: .medium))
                    .foregroundStyle(SenteiTheme.urgencyColor(.urgent))
            }
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
    }

    private var connectionColor: Color {
        switch viewModel.connectionState {
        case .connected: return SenteiTheme.accentPrimary
        case .disconnected: return SenteiTheme.urgentKakishibu
        case .connecting: return SenteiTheme.urgencyKokeKi
        }
    }

    private var connectionText: String {
        switch viewModel.connectionState {
        case .connected: return "接続中"
        case .disconnected: return "未接続"
        case .connecting: return "接続中..."
        }
    }

    // MARK: - Item List

    private var itemList: some View {
        Group {
            if viewModel.connectionState == .disconnected {
                VStack(spacing: 8) {
                    Image(systemName: "wifi.slash")
                        .font(.system(size: 24))
                        .foregroundStyle(SenteiTheme.textTertiary)
                    Text("sentei serve を起動してください")
                        .font(.system(size: 12))
                        .foregroundStyle(SenteiTheme.textTertiary)
                }
                .frame(maxWidth: .infinity)
                .frame(height: 440)
            } else if viewModel.popoverItems.isEmpty {
                Text("アイテムがありません")
                    .font(.system(size: 12))
                    .foregroundStyle(SenteiTheme.textTertiary)
                    .frame(maxWidth: .infinity)
                    .frame(height: 440)
            } else {
                ScrollView {
                    VStack(spacing: 4) {
                        ForEach(viewModel.popoverItems) { item in
                            ItemRow(item: item) {
                                Task { await viewModel.deleteItem(item) }
                            }
                        }
                    }
                    .padding(8)
                }
                .frame(height: 440)
            }
        }
    }

    // MARK: - Footer

    private var footer: some View {
        HStack {
            Button("ダッシュボードを開く") {
                openWindow(id: "dashboard")
            }
            .buttonStyle(.plain)
            .font(.system(size: 11))
            .foregroundStyle(SenteiTheme.textSecondary)

            Spacer()

            Button("終了") {
                viewModel.shutdown()
                NSApplication.shared.terminate(nil)
            }
            .buttonStyle(.plain)
            .font(.system(size: 11))
            .foregroundStyle(SenteiTheme.textTertiary)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
    }
}
