/**
 * sentei デーモンのライフサイクル管理
 * アプリ起動時に sentei serve を spawn、アプリ終了時に停止する（Ollama パターン）
 */
import Foundation

/// sentei デーモンプロセスの管理
final class DaemonManager {
    private var process: Process?
    private let binaryPath: String

    init(binaryPath: String? = nil) {
        // sentei バイナリのパスを検索
        self.binaryPath = binaryPath ?? DaemonManager.findBinary()
    }

    /// デーモンが稼働中でなければ起動する
    func ensureRunning(apiClient: APIClient) async {
        let alive = await apiClient.isAlive()
        if alive { return }

        start()

        // デーモンの起動を待つ（最大 5 秒）
        for _ in 0..<10 {
            try? await Task.sleep(for: .milliseconds(500))
            let alive = await apiClient.isAlive()
            if alive { return }
        }
    }

    /// デーモンを起動する
    func start() {
        guard process == nil else { return }

        let proc = Process()
        proc.executableURL = URL(fileURLWithPath: binaryPath)
        proc.arguments = ["serve"]
        proc.standardOutput = FileHandle.nullDevice
        proc.standardError = FileHandle.nullDevice

        do {
            try proc.run()
            process = proc
        } catch {
            print("sentei serve 起動失敗: \(error)")
        }
    }

    /// デーモンを停止する
    func stop() {
        guard let proc = process, proc.isRunning else {
            process = nil
            return
        }
        proc.terminate()
        proc.waitUntilExit()
        process = nil
    }

    /// sentei バイナリのパスを検索する
    private static func findBinary() -> String {
        // 1. PATH 上の sentei を検索
        let whichProcess = Process()
        let pipe = Pipe()
        whichProcess.executableURL = URL(fileURLWithPath: "/usr/bin/which")
        whichProcess.arguments = ["sentei"]
        whichProcess.standardOutput = pipe
        whichProcess.standardError = FileHandle.nullDevice

        do {
            try whichProcess.run()
            whichProcess.waitUntilExit()
            let data = pipe.fileHandleForReading.readDataToEndOfFile()
            let path = String(data: data, encoding: .utf8)?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
            if !path.isEmpty { return path }
        } catch {}

        // 2. go install のデフォルトパス
        let home = FileManager.default.homeDirectoryForCurrentUser.path
        let goPath = "\(home)/go/bin/sentei"
        if FileManager.default.fileExists(atPath: goPath) { return goPath }

        // 3. フォールバック
        return "/usr/local/bin/sentei"
    }
}
