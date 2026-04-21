/**
 * HTTP サーバー（REST API）
 * CLI や将来の Web UI からのリクエストを処理する
 */
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/senna-lang/sentei/internal/core"
	"github.com/senna-lang/sentei/internal/plugin"
	"github.com/senna-lang/sentei/internal/storage"
	"github.com/senna-lang/sentei/internal/summary"
)

// Server は REST API サーバー
type Server struct {
	engine     *core.Engine
	addr       string
	srv        *http.Server
	shutdownCh chan struct{}
	ghUser     githubUserCache
}

// New は Server を生成する
func New(engine *core.Engine, addr string) *Server {
	return &Server{
		engine:     engine,
		addr:       addr,
		shutdownCh: make(chan struct{}),
	}
}

// ShutdownCh はシャットダウン要求を受け取るチャネルを返す
func (s *Server) ShutdownCh() <-chan struct{} {
	return s.shutdownCh
}

// Start はサーバーを起動する
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/items", s.handleListItems)
	mux.HandleFunc("GET /api/summary", s.handleSummary)
	mux.HandleFunc("GET /api/summary/{repo...}", s.handleSummaryRepo)
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("POST /api/shutdown", s.handleShutdown)
	mux.HandleFunc("DELETE /api/items/{source}/{source_id}", s.handleDeleteItem)

	s.srv = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	slog.Info("API サーバー起動", "addr", s.addr)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("サーバー起動失敗: %w", err)
	}
	return nil
}

// Shutdown はサーバーをグレースフルに停止する
func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

// handleListItems は GET /api/items を処理する
func (s *Server) handleListItems(w http.ResponseWriter, r *http.Request) {
	filter := storage.ItemFilter{
		Urgency:  r.URL.Query().Get("urgency"),
		Source:   r.URL.Query().Get("source"),
		Category: r.URL.Query().Get("category"),
		Limit:    100,
	}

	items, err := s.engine.Storage().ListItems(filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// handleSummary は GET /api/summary を処理する（全リポジトリ）
// サマリーはレポジトリ全体の「今日の」活動オーバービュー
// 過去 30 日の survey 結果はストレージに残っているが、表示は今日分のみに絞る
// repo リストは過去全 survey から取り、今日活動ゼロの repo も空サマリーで返す
func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	items, err := s.engine.Storage().ListItems(storage.ItemFilter{Source: "git"})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	allSurveyed := groupByRepo(filterSurveyed(items))
	today := groupByRepo(filterTodaySurvey(items))
	date := time.Now().Format("2006-01-02")
	myUser := s.ghUser.get()

	var summaries []map[string]string
	for repo := range allSurveyed {
		repoItems := today[repo] // 今日は空もあり得る
		bonsaiSummary := s.engine.BonsaiClient().GenerateSummary(repo, repoItems)

		summaryText := summary.Render(summary.SummaryData{
			Repo:    repo,
			Date:    date,
			Items:   repoItems,
			Summary: bonsaiSummary,
			MyUser:  myUser,
		})
		summaries = append(summaries, map[string]string{
			"repo":    repo,
			"summary": summaryText,
		})
	}

	writeJSON(w, http.StatusOK, summaries)
}

// handleSummaryRepo は GET /api/summary/{repo} を処理する（特定リポジトリ）
// 今日のアクティビティのみを返す（過去 survey があれば repo 自体は有効）
func (s *Server) handleSummaryRepo(w http.ResponseWriter, r *http.Request) {
	repo := r.PathValue("repo")
	items, err := s.engine.Storage().ListItems(storage.ItemFilter{Source: "git"})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 過去に survey されたことがあれば repo 存在ありとみなす
	allSurveyed := groupByRepo(filterSurveyed(items))
	if _, ok := allSurveyed[repo]; !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("リポジトリ %s のサマリーがありません", repo))
		return
	}

	today := groupByRepo(filterTodaySurvey(items))
	summaryText := summary.Render(summary.SummaryData{
		Repo:   repo,
		Date:   time.Now().Format("2006-01-02"),
		Items:  today[repo],
		MyUser: s.ghUser.get(),
	})
	writeJSON(w, http.StatusOK, map[string]string{
		"repo":    repo,
		"summary": summaryText,
	})
}

// filterSurveyed は survey 由来のアイテム（過去全期間）を返す
// 「この repo は survey 対象として登録されている」判定に使う
func filterSurveyed(items []plugin.LabeledItem) []plugin.LabeledItem {
	var out []plugin.LabeledItem
	for _, li := range items {
		if li.Item.Metadata["survey_type"] != "" {
			out = append(out, li)
		}
	}
	return out
}

// filterTodaySurvey は今日の survey 由来アイテムだけを返す
// サマリー = 今日のレポジトリ活動サマリー、という意図に合わせる
// ストレージ自体は 30 日分を保持（履歴・トレンド用途）するが、サマリー表示は今日分に絞る
func filterTodaySurvey(items []plugin.LabeledItem) []plugin.LabeledItem {
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var out []plugin.LabeledItem
	for _, li := range items {
		if li.Item.Metadata["survey_type"] == "" {
			continue
		}
		if li.Item.Timestamp.Before(startOfToday) {
			continue
		}
		out = append(out, li)
	}
	return out
}

// handleStatus は GET /api/status を処理する
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	itemCount, _ := s.engine.Storage().ItemCount()
	lastLabeled, _ := s.engine.Storage().LastLabeledAt()

	bonsaiOK := "error"
	if err := s.engine.BonsaiClient().Ping(); err == nil {
		bonsaiOK = "ok"
	}

	status := map[string]any{
		"daemon":     "running",
		"bonsai":     bonsaiOK,
		"plugins":    s.engine.PluginNames(),
		"item_count": itemCount,
	}
	// zero time (ラベリング履歴なし / DB 値が parse 不能) の場合はフィールド自体を省く
	if !lastLabeled.IsZero() {
		status["last_labeled"] = lastLabeled
	}

	writeJSON(w, http.StatusOK, status)
}

// handleDeleteItem は DELETE /api/items/{source}/{source_id} を処理する
func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	source := r.PathValue("source")
	sourceID := r.PathValue("source_id")

	deleted, err := s.engine.Storage().DeleteItem(source, sourceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, fmt.Sprintf("アイテムが見つかりません: %s/%s", source, sourceID))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// handleShutdown は POST /api/shutdown を処理する
func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "shutting_down"})

	// レスポンス送信後にシャットダウンシグナルを送る
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(s.shutdownCh)
	}()
}

func groupByRepo(items []plugin.LabeledItem) map[string][]plugin.LabeledItem {
	groups := make(map[string][]plugin.LabeledItem)
	for _, item := range items {
		repo := item.Item.Metadata["repo"]
		if repo == "" {
			repo = item.Item.Source
		}
		groups[repo] = append(groups[repo], item)
	}
	return groups
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
