# 実装タスク

## Phase 1: テーマ基盤の刷新

1. `SenteiTheme.swift` の background パレットを更新（primary `#141815` / secondary `#1A1F1C` / card `#222823` / cardBorder を緑寄り 6% 不透明度に）
2. `SenteiTheme.swift` のテキスト階層を washi 系に置換（primary `#E8EDE5` / secondary `#A8B0A4` / tertiary `#6B756A`）
3. `SenteiTheme.swift` のアクセント緑を再定義（accentPrimary=koke `#7BA05B` / accentEvergreen `#4A6B3D` / accentLeaf `#A4C97A`）
4. `SenteiTheme.swift` の `urgencyColor(_:)` を自然色マッピングへ変更（urgent `#C97E4A` / shouldCheck `#C9B560` / canWait `#7A8A78` / ignore `#4A524A`）
5. `SenteiTheme.swift` に divider / borderSubtle / focusRing を新設（`#2A302C` / `#2F3631` / koke + opacity）
6. 既存定数（`accentGreen` / `accentRed` 等）の互換性を確認し、参照箇所を新パレットへ置換

## Phase 2: ビュー適用

7. `DashboardView.swift` の background 参照を新 backgroundPrimary に当て、視覚確認
8. `SidebarView.swift` の status 行（ok/error）を accentPrimary / urgent 新色に差し替え、background を新 backgroundSecondary へ
9. `ItemListView.swift` の filterBar 背景・divider・empty state を新パレットへ
10. `BoardView.swift` の text / empty state を新パレットへ
11. `PopoverView.swift` の connection state（accentGreen → accentPrimary、accentRed → urgent 新色、orange は維持か再検討）と footer ボタンを新パレットへ
12. `ItemRow.swift` のカード背景 / border / チェックアイコン色を新パレットへ
13. `UrgencyBadge.swift` / `CategoryIcon.swift` の参照確認（パレット経由なので原則無修正）

## Phase 3: 仕上げ

14. `swift build` でビルドを通す
15. `swift run` でアプリを起動し、ポップオーバー / ダッシュボード / 各 urgency 表示を目視確認
16. コントラスト比を実機で確認（特に textTertiary on cardBackground）
17. 必要に応じて focus ring / hover の挙動を微調整

## Phase 4: 仕様 / ドキュメント

18. `spec/specs/macos-app/spec.md` の Requirement「デザインテーマ」を spec-delta の MODIFIED 内容で更新
19. README にスクリーンショットがあれば差し替え（任意）

---

**メモ**:
- パレットは `SenteiTheme.swift` に集約されているため、ビュー側のハードコード色を見つけたら都度パレット参照へ寄せる
- 視覚確認は `swift run` ではなく `.app` 経由が望ましい（フォントレンダリング差異）が、最低 `swift run` で確認できれば可
