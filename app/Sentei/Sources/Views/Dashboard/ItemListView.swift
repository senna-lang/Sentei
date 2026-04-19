/**
 * ダッシュボードのアイテム一覧
 * urgency / source / category の Picker フィルタ + ItemRow の縦リスト
 * サイドバーで urgency 選択時はその urgency に固定する
 */
import SwiftUI

/// アイテム一覧ビュー
struct ItemListView: View {
    @Bindable var viewModel: AppViewModel
    /// サイドバーで固定された urgency（nil なら Picker で選択可能）
    let fixedUrgency: Urgency?

    @State private var urgencyFilter: Urgency? = nil
    @State private var sourceFilter: String = "all"
    @State private var categoryFilter: String = "all"

    var body: some View {
        VStack(spacing: 0) {
            filterBar
            Divider()
            itemList
        }
        .background(SenteiTheme.backgroundPrimary)
        .onAppear {
            urgencyFilter = fixedUrgency
        }
        .onChange(of: fixedUrgency) { _, newValue in
            urgencyFilter = newValue
        }
    }

    // MARK: - Filter Bar

    private var filterBar: some View {
        HStack(spacing: 12) {
            if fixedUrgency == nil {
                filterPicker(
                    label: "urgency",
                    selection: Binding(
                        get: { urgencyFilter?.rawValue ?? "all" },
                        set: { urgencyFilter = $0 == "all" ? nil : Urgency(rawValue: $0) }
                    ),
                    options: ["all"] + Urgency.allCases.map(\.rawValue)
                )
            }

            filterPicker(
                label: "source",
                selection: $sourceFilter,
                options: ["all"] + availableSources
            )

            filterPicker(
                label: "category",
                selection: $categoryFilter,
                options: ["all"] + availableCategories
            )

            Spacer()

            Text("\(filteredItems.count) 件")
                .font(.system(size: 11))
                .foregroundStyle(SenteiTheme.textTertiary)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 8)
        .background(SenteiTheme.backgroundSecondary)
    }

    private func filterPicker(label: String, selection: Binding<String>, options: [String]) -> some View {
        HStack(spacing: 4) {
            Text(label)
                .font(.system(size: 11))
                .foregroundStyle(SenteiTheme.textTertiary)
            Picker("", selection: selection) {
                ForEach(options, id: \.self) { opt in
                    Text(opt).tag(opt)
                }
            }
            .pickerStyle(.menu)
            .labelsHidden()
            .font(.system(size: 11))
        }
    }

    // MARK: - Item List

    private var itemList: some View {
        Group {
            if filteredItems.isEmpty {
                emptyState
            } else {
                ScrollView {
                    LazyVStack(spacing: 4) {
                        ForEach(filteredItems) { item in
                            ItemRow(item: item) {
                                Task { await viewModel.deleteItem(item) }
                            }
                        }
                    }
                    .padding(12)
                }
            }
        }
    }

    private var emptyState: some View {
        VStack(spacing: 8) {
            Image(systemName: "tray")
                .font(.system(size: 32))
                .foregroundStyle(SenteiTheme.textTertiary)
            Text("該当するアイテムがありません")
                .font(.system(size: 12))
                .foregroundStyle(SenteiTheme.textTertiary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    // MARK: - Computed

    private var filteredItems: [SenteiItem] {
        viewModel.items.filter { item in
            if let urgencyFilter, item.label.urgency != urgencyFilter { return false }
            if sourceFilter != "all", item.item.source != sourceFilter { return false }
            if categoryFilter != "all", item.label.category != categoryFilter { return false }
            return true
        }
    }

    private var availableSources: [String] {
        Array(Set(viewModel.items.map(\.item.source))).sorted()
    }

    private var availableCategories: [String] {
        Array(Set(viewModel.items.map(\.label.category))).sorted()
    }
}
