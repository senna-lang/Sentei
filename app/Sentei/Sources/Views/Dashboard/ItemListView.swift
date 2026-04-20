/**
 * ダッシュボードのアイテム一覧
 * source (git/rss) で固定されたスコープ内で、必要なら urgency または category を更に絞る。
 * サイドバーからの選択で source + 固定値が決まり、Picker は「そのスコープで意味のあるもの」だけ出す。
 */
import SwiftUI

/// アイテム一覧ビュー
struct ItemListView: View {
    @Bindable var viewModel: AppViewModel
    /// source 固定値 ("git" または "rss")
    let source: String
    /// サイドバーで固定された urgency (git 用。nil なら Picker で選択可能)
    let fixedUrgency: Urgency?
    /// サイドバーで固定された category (rss 用。nil なら Picker で選択可能)
    let fixedCategory: String?

    @State private var urgencyFilter: Urgency? = nil
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
            categoryFilter = fixedCategory ?? "all"
        }
        .onChange(of: fixedUrgency) { _, newValue in
            urgencyFilter = newValue
        }
        .onChange(of: fixedCategory) { _, newValue in
            categoryFilter = newValue ?? "all"
        }
        .onChange(of: source) { _, _ in
            urgencyFilter = fixedUrgency
            categoryFilter = fixedCategory ?? "all"
        }
    }

    // MARK: - Filter Bar

    private var filterBar: some View {
        HStack(spacing: 12) {
            // source 固定表示 (Picker ではない、現在のスコープを明示)
            Text(source)
                .font(.system(size: 11, weight: .medium))
                .foregroundStyle(SenteiTheme.textSecondary)
                .padding(.horizontal, 6)
                .padding(.vertical, 2)
                .background(SenteiTheme.cardBackground)
                .cornerRadius(4)

            if source == "git", fixedUrgency == nil {
                filterPicker(
                    label: "urgency",
                    selection: Binding(
                        get: { urgencyFilter?.rawValue ?? "all" },
                        set: { urgencyFilter = $0 == "all" ? nil : Urgency(rawValue: $0) }
                    ),
                    options: ["all"] + Urgency.allCases.map(\.rawValue)
                )
            }

            if source == "rss", fixedCategory == nil {
                filterPicker(
                    label: "category",
                    selection: $categoryFilter,
                    options: ["all"] + availableCategories
                )
            }

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
            guard item.item.source == source else { return false }
            if let urgencyFilter, item.label.urgency != urgencyFilter { return false }
            if categoryFilter != "all", item.label.category != categoryFilter { return false }
            return true
        }
    }

    private var availableCategories: [String] {
        Array(Set(viewModel.items.filter { $0.item.source == source }.map(\.label.category))).sorted()
    }
}
