/**
 * Bonsai ラベリングクライアント
 * llama.cpp の /completion エンドポイントに GBNF grammar 付きリクエストを送信し、
 * urgency + category + summary の構造化 JSON を取得する
 */
package bonsai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/senna-lang/sentei/internal/plugin"
)

// Client は Bonsai (llama.cpp server) へのHTTPクライアント
type Client struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

// NewClient は Bonsai クライアントを生成する
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
	}
}

// completionRequest は /completion エンドポイントへのリクエスト
type completionRequest struct {
	Prompt   string `json:"prompt"`
	Grammar  string `json:"grammar"`
	Temperature float64 `json:"temperature"`
	NPredict int    `json:"n_predict"`
	Stop     []string `json:"stop"`
}

// completionResponse は /completion エンドポイントからのレスポンス
type completionResponse struct {
	Content string `json:"content"`
}

// labelResponse は Bonsai が返す JSON の構造
type labelResponse struct {
	Urgency  string `json:"urgency"`
	Category string `json:"category"`
	Summary  string `json:"summary"`
}

// Label は Item に対して Bonsai ラベリングを実行する
func (c *Client) Label(item plugin.Item, grammar string, promptTemplate string) (plugin.Label, error) {
	prompt := buildPrompt(item, promptTemplate)

	var lastErr error
	for attempt := range c.maxRetries {
		label, err := c.callCompletion(prompt, grammar)
		if err == nil {
			return label, nil
		}
		lastErr = err
		slog.Warn("Bonsai ラベリング失敗、リトライ中",
			"attempt", attempt+1,
			"max", c.maxRetries,
			"error", err,
		)
		time.Sleep(time.Duration(1<<uint(attempt)) * time.Second) // 指数バックオフ
	}
	return fallbackLabel(), fmt.Errorf("Bonsai ラベリング失敗 (%d 回リトライ後): %w", c.maxRetries, lastErr)
}

// callCompletion は /completion エンドポイントを呼び出す
func (c *Client) callCompletion(prompt, grammar string) (plugin.Label, error) {
	reqBody := completionRequest{
		Prompt:      prompt,
		Grammar:     grammar,
		Temperature: 0.2,
		NPredict:    150,
		Stop:        []string{"}"},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return plugin.Label{}, fmt.Errorf("リクエスト JSON 生成失敗: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/completion", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return plugin.Label{}, fmt.Errorf("Bonsai 接続失敗: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return plugin.Label{}, fmt.Errorf("レスポンス読み取り失敗: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return plugin.Label{}, fmt.Errorf("Bonsai エラー (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var compResp completionResponse
	if err := json.Unmarshal(body, &compResp); err != nil {
		return plugin.Label{}, fmt.Errorf("レスポンス JSON パース失敗: %w", err)
	}

	return parseLabel(compResp.Content)
}

// parseLabel は Bonsai の出力 JSON をパースする
func parseLabel(content string) (plugin.Label, error) {
	content = strings.TrimSpace(content)
	// stop で } が消える場合の補完
	if !strings.HasSuffix(content, "}") {
		content += "}"
	}

	var lr labelResponse
	if err := json.Unmarshal([]byte(content), &lr); err != nil {
		return fallbackLabel(), fmt.Errorf("ラベル JSON パース失敗: %w (raw: %s)", err, content)
	}

	return plugin.Label{
		Urgency:  plugin.Urgency(lr.Urgency),
		Category: lr.Category,
		Summary:  lr.Summary,
	}, nil
}

// fallbackLabel はパース失敗時のデフォルトラベルを返す
func fallbackLabel() plugin.Label {
	return plugin.Label{
		Urgency:  plugin.UrgencyShouldCheck,
		Category: "other",
		Summary:  "",
	}
}

// buildPrompt は Item からプロンプトを組み立てる
func buildPrompt(item plugin.Item, template string) string {
	itemJSON, _ := json.Marshal(map[string]any{
		"source":   item.Source,
		"title":    item.Title,
		"content":  item.Content,
		"metadata": item.Metadata,
	})
	return strings.Replace(template, "{notification_json}", string(itemJSON), 1)
}

// Ping は Bonsai サーバーの生存確認
func (c *Client) Ping() error {
	resp, err := c.httpClient.Get(c.baseURL + "/v1/models")
	if err != nil {
		return fmt.Errorf("Bonsai 接続失敗: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bonsai 応答異常 (HTTP %d)", resp.StatusCode)
	}
	return nil
}
