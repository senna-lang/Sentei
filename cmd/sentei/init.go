/**
 * init コマンド — 初期設定ディレクトリと config.toml を作成する
 */
package main

import (
	"fmt"
	"os"

	"github.com/senna-lang/sentei/internal/cli"
	"github.com/senna-lang/sentei/internal/config"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "初期設定を作成する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "既存の設定ファイルを上書きする")

	return cmd
}

func runInit(force bool) error {
	configDir := config.ConfigDir()

	// ディレクトリ作成
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("ディレクトリ作成失敗: %w", err)
	}

	// config.toml 作成
	configPath := config.DefaultPath()
	if _, err := os.Stat(configPath); err == nil && !force {
		fmt.Println(cli.Success("既に初期化されています") + " " + configPath)
		fmt.Println("上書きするには --force を指定してください")
		return nil
	}

	if err := os.WriteFile(configPath, []byte(config.DefaultTOML()), 0644); err != nil {
		return fmt.Errorf("config.toml 作成失敗: %w", err)
	}
	fmt.Printf("%s %s\n", cli.Success("config.toml 作成完了:"), configPath)

	// LaunchAgent plist 作成
	execPath, err := os.Executable()
	if err != nil {
		execPath = "sentei"
	}

	plistContent := config.GeneratePlist(execPath)
	plistPath := config.PlistPath()

	// LaunchAgents ディレクトリ確認
	if err := os.MkdirAll(os.ExpandEnv("$HOME/Library/LaunchAgents"), 0755); err != nil {
		fmt.Printf("%s LaunchAgents ディレクトリ作成失敗: %v\n", cli.Error("警告:"), err)
		return nil
	}

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		fmt.Printf("%s plist 作成失敗: %v\n", cli.Error("警告:"), err)
		return nil
	}
	fmt.Printf("%s %s\n", cli.Success("LaunchAgent plist 作成完了:"), plistPath)

	fmt.Println()
	fmt.Println("自動起動を有効にするには:")
	fmt.Printf("  launchctl load %s\n", plistPath)
	fmt.Println()
	fmt.Println("手動で起動するには:")
	fmt.Println("  sentei serve")

	return nil
}
