/**
 * sentei のカスタムテーマ定義
 * 盆栽 / 剪定モチーフの緑基調ダークテーマ（苔庭 × 墨 × 古銅）
 * - 背景: 純黒ではなく緑寄りの墨色
 * - アクセント: 苔色の低彩度グリーン
 * - urgent のみ柿渋（古銅）— 唯一の暖色で侘び寂びの一点破調
 */
import SwiftUI

/// sentei のカラーパレット
enum SenteiTheme {
    // MARK: - Urgency Colors (自然色マッピング)

    static func urgencyColor(_ urgency: Urgency) -> Color {
        switch urgency {
        case .urgent: return urgentKakishibu
        case .shouldCheck: return urgencyKokeKi
        case .canWait: return urgencyTakeNezu
        case .ignore: return urgencyKare
        }
    }

    // MARK: - Category Icons

    static func categoryIcon(_ category: String) -> String {
        switch category {
        case "pr": return "arrow.triangle.branch"
        case "issue": return "exclamationmark.circle"
        case "ci": return "gearshape.2"
        case "release": return "shippingbox"
        case "discussion": return "bubble.left.and.bubble.right"
        default: return "circle"
        }
    }

    // MARK: - Background (墨 sumi)

    static let backgroundPrimary = Color(hex: 0x141815)   // 墨
    static let backgroundSecondary = Color(hex: 0x1A1F1C) // 影
    static let cardBackground = Color(hex: 0x222823)      // 苔土
    static let cardBorder = Color(hex: 0xB4C8B4, opacity: 0.06)

    // MARK: - Text (和紙 washi)

    static let textPrimary = Color(hex: 0xE8EDE5)   // 和紙
    static let textSecondary = Color(hex: 0xA8B0A4) // 古葉
    static let textTertiary = Color(hex: 0x6B756A)  // 涸苔

    // MARK: - Accent (苔 koke)

    /// 主要アクセント（接続中・focus・primary action）
    static let accentPrimary = Color(hex: 0x7BA05B)   // 苔
    /// 二次アクセント（hover・section header）
    static let accentEvergreen = Color(hex: 0x4A6B3D) // 常緑
    /// 高彩度アクセント（success・celebration）
    static let accentLeaf = Color(hex: 0xA4C97A)      // 若葉

    // MARK: - Urgency tones (個別アクセス用)

    /// urgent — 柿渋 / 古銅。緑の調和を一点だけ破る唯一の暖色
    static let urgentKakishibu = Color(hex: 0xC97E4A)
    /// should_check — 苔黄
    static let urgencyKokeKi = Color(hex: 0xC9B560)
    /// can_wait — 竹鼠
    static let urgencyTakeNezu = Color(hex: 0x7A8A78)
    /// ignore — 涸
    static let urgencyKare = Color(hex: 0x4A524A)

    // MARK: - Divider / Border / Focus

    static let divider = Color(hex: 0x2A302C)
    static let borderSubtle = Color(hex: 0x2F3631)
    static let focusRing = Color(hex: 0x7BA05B, opacity: 0.6)
}

// MARK: - Color Extension

extension Color {
    init(hex: UInt, opacity: Double = 1.0) {
        self.init(
            .sRGB,
            red: Double((hex >> 16) & 0xFF) / 255,
            green: Double((hex >> 8) & 0xFF) / 255,
            blue: Double(hex & 0xFF) / 255,
            opacity: opacity
        )
    }
}
