/**
 * serve コマンド — デーモンとして API サーバー + プラグインを起動する
 * config.toml から設定を読み込み、コアエンジンを初期化する
 */
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/senna-lang/sentei/internal/config"
	"github.com/senna-lang/sentei/internal/core"
	"github.com/senna-lang/sentei/internal/server"
	gitplugin "github.com/senna-lang/sentei/plugins/git"
	rssplugin "github.com/senna-lang/sentei/plugins/rss"
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "デーモンとして API サーバーを起動する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe()
		},
	}
}

func runServe() error {
	// 設定読み込み
	cfg := loadConfig()

	slog.Info("sentei 起動中", "addr", cfg.Daemon.Addr)

	// config ディレクトリが存在しなければ作成
	if err := os.MkdirAll(config.ConfigDir(), 0755); err != nil {
		return fmt.Errorf("config ディレクトリ作成失敗: %w", err)
	}

	dbPath := config.ExpandPath(cfg.Daemon.DBPath)

	engine, err := core.New(core.Config{
		DBPath:    dbPath,
		BonsaiURL: cfg.Bonsai.URL,
	})
	if err != nil {
		return fmt.Errorf("コアエンジン初期化失敗: %w", err)
	}

	// Git プラグイン登録
	if cfg.Plugins.Git.Enabled {
		repos := make([]gitplugin.RepoConfig, len(cfg.Plugins.Git.Repos))
		for i, r := range cfg.Plugins.Git.Repos {
			repos[i] = gitplugin.RepoConfig{GitHub: r}
		}

		gp := gitplugin.NewPlugin(gitplugin.Config{
			Enabled: true,
			Notification: gitplugin.NotificationConfig{
				PollInterval: time.Duration(cfg.Plugins.Git.PollIntervalSec) * time.Second,
			},
			Survey: gitplugin.SurveyConfig{
				Interval: time.Duration(cfg.Plugins.Git.SurveyIntervalSec) * time.Second,
				Repos:    repos,
			},
		})
		engine.RegisterPlugin(gp, gitplugin.GitGrammar, gitplugin.GitPromptTemplate)
	}

	// RSS プラグイン登録 (opt-in)
	if cfg.Plugins.Rss.Enabled {
		userAgent := fmt.Sprintf("sentei/%s (+https://github.com/senna-lang/Sentei)", Version)
		rp := rssplugin.NewPlugin(cfg.Plugins.Rss, engine.Storage(), userAgent)
		engine.RegisterPlugin(rp, rssplugin.Grammar, rssplugin.PromptTemplate)
	}

	// コアエンジン起動
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		return fmt.Errorf("コアエンジン起動失敗: %w", err)
	}

	// API サーバー起動
	srv := server.New(engine, cfg.Daemon.Addr)
	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("API サーバーエラー", "error", err)
		}
	}()

	// シグナルハンドリング + API shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		slog.Info("シャットダウン開始", "signal", sig)
	case <-srv.ShutdownCh():
		slog.Info("API 経由でシャットダウン要求を受信")
	}

	srv.Shutdown(ctx)
	engine.Stop()

	slog.Info("sentei 停止完了")
	return nil
}
