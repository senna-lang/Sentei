/**
 * Git プラグイン - バッチサーベイ型
 * 監視リポジトリの活動を定期的に GitHub API で収集し、
 * 各アイテムに Bonsai ラベルを付けてコアに送信する
 */
package git

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/senna-lang/sentei/internal/plugin"
)

// gitSurveyCmdTimeout は 1 回の gh コマンドに許すタイムアウト。
// gh 本体にも TCP タイムアウトはあるが、hang したとき hourly ticker が埋まって
// 以降のサーベイが走らなくなる事故を防ぐため、明示的な context 締切を入れる。
const gitSurveyCmdTimeout = 30 * time.Second

// runGH はタイムアウト付きで gh コマンドを実行して stdout を返す。
// context.DeadlineExceeded の場合は識別しやすいエラーに包む。
func runGH(ctx context.Context, args ...string) ([]byte, error) {
	cctx, cancel := context.WithTimeout(ctx, gitSurveyCmdTimeout)
	defer cancel()

	out, err := exec.CommandContext(cctx, "gh", args...).Output()
	if err != nil {
		if cctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("gh コマンドタイムアウト (%s): %w", gitSurveyCmdTimeout, err)
		}
		return nil, err
	}
	return out, nil
}

// runSurvey は監視リポジトリを定期的にサーベイする
func (p *Plugin) runSurvey(ctx context.Context) {
	defer p.wg.Done()

	slog.Info("Git サーベイ開始", "interval", p.config.Survey.Interval, "repos", len(p.config.Survey.Repos))

	// 初回は即実行
	p.surveyAllRepos(ctx)

	ticker := time.NewTicker(p.config.Survey.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Git サーベイ停止")
			return
		case <-ticker.C:
			p.surveyAllRepos(ctx)
		}
	}
}

// surveyAllRepos は全監視リポジトリをサーベイする
// ティック毎に 1 行ログを残すので、ticker が止まっていないか運用中に確認できる。
func (p *Plugin) surveyAllRepos(ctx context.Context) {
	slog.Info("Git サーベイ実行", "repos", len(p.config.Survey.Repos))

	for _, repo := range p.config.Survey.Repos {
		if ctx.Err() != nil {
			return
		}
		items, err := p.surveyRepo(ctx, repo.GitHub)
		if err != nil {
			slog.Error("リポジトリサーベイ失敗", "repo", repo.GitHub, "error", err)
			continue
		}

		submitted := 0
		for _, item := range items {
			if err := p.core.Submit(item); err != nil {
				slog.Error("サーベイ Item Submit 失敗", "repo", repo.GitHub, "error", err)
				continue
			}
			submitted++
		}

		if submitted > 0 {
			slog.Info("サーベイ完了", "repo", repo.GitHub, "items", submitted)
		}
	}
}

// surveyRepo は 1 つのリポジトリの最新活動を収集する
func (p *Plugin) surveyRepo(ctx context.Context, repo string) ([]plugin.Item, error) {
	var items []plugin.Item

	// マージ済み PR
	if prs, err := p.fetchMergedPRs(ctx, repo); err != nil {
		slog.Warn("マージ済み PR 取得失敗", "repo", repo, "error", err)
	} else {
		items = append(items, prs...)
	}

	// オープン中の PR
	if prs, err := p.fetchOpenPRs(ctx, repo); err != nil {
		slog.Warn("オープン PR 取得失敗", "repo", repo, "error", err)
	} else {
		items = append(items, prs...)
	}

	// 新規 Issue
	if issues, err := p.fetchRecentIssues(ctx, repo); err != nil {
		slog.Warn("新規 Issue 取得失敗", "repo", repo, "error", err)
	} else {
		items = append(items, issues...)
	}

	// 直近のリリース
	if releases, err := p.fetchRecentReleases(ctx, repo); err != nil {
		slog.Warn("リリース取得失敗", "repo", repo, "error", err)
	} else {
		items = append(items, releases...)
	}

	return items, nil
}

// ghPullRequest は GitHub PR のレスポンス
type ghPullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	HTMLURL   string    `json:"html_url"`
	MergedAt  *string   `json:"merged_at"`
	User      ghUser    `json:"user"`
	Labels    []ghLabel `json:"labels"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ghIssue は GitHub Issue のレスポンス
type ghIssue struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	HTMLURL     string    `json:"html_url"`
	User        ghUser    `json:"user"`
	Labels      []ghLabel `json:"labels"`
	CreatedAt   time.Time `json:"created_at"`
	PullRequest *struct{} `json:"pull_request"` // nil なら Issue、非 nil なら PR
}

type ghUser struct {
	Login string `json:"login"`
}

type ghLabel struct {
	Name string `json:"name"`
}

// fetchMergedPRs は直近にマージされた PR を取得する
func (p *Plugin) fetchMergedPRs(ctx context.Context, repo string) ([]plugin.Item, error) {
	since := time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339)

	output, err := runGH(ctx, "api",
		fmt.Sprintf("repos/%s/pulls?state=closed&sort=updated&direction=desc&per_page=20", repo),
	)
	if err != nil {
		return nil, fmt.Errorf("gh api 失敗: %w", err)
	}

	var prs []ghPullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("PR JSON パース失敗: %w", err)
	}

	var items []plugin.Item
	for _, pr := range prs {
		if pr.MergedAt == nil || *pr.MergedAt < since {
			continue
		}

		labels := make([]string, len(pr.Labels))
		for i, l := range pr.Labels {
			labels[i] = l.Name
		}

		items = append(items, plugin.Item{
			Source:    "git",
			SourceID:  fmt.Sprintf("survey-pr-%s-%d", repo, pr.Number),
			Title:     fmt.Sprintf("#%d %s", pr.Number, pr.Title),
			Content:   fmt.Sprintf("Merged PR: %s by %s", pr.Title, pr.User.Login),
			URL:       pr.HTMLURL,
			Timestamp: pr.UpdatedAt,
			Metadata: map[string]string{
				"repo":        repo,
				"author":      pr.User.Login,
				"subject_type": "PullRequest",
				"survey_type": "merged_pr",
				"labels":      strings.Join(labels, ","),
			},
		})
	}

	return items, nil
}

// fetchRecentIssues は直近に作成された Issue を取得する
func (p *Plugin) fetchRecentIssues(ctx context.Context, repo string) ([]plugin.Item, error) {
	since := time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339)

	output, err := runGH(ctx, "api",
		fmt.Sprintf("repos/%s/issues?state=all&since=%s&sort=created&direction=desc&per_page=20", repo, since),
	)
	if err != nil {
		return nil, fmt.Errorf("gh api 失敗: %w", err)
	}

	var issues []ghIssue
	if err := json.Unmarshal(output, &issues); err != nil {
		return nil, fmt.Errorf("Issue JSON パース失敗: %w", err)
	}

	var items []plugin.Item
	for _, iss := range issues {
		// PR を除外（GitHub API は Issue エンドポイントに PR も含む）
		if iss.PullRequest != nil {
			continue
		}

		labels := make([]string, len(iss.Labels))
		for i, l := range iss.Labels {
			labels[i] = l.Name
		}

		items = append(items, plugin.Item{
			Source:    "git",
			SourceID:  fmt.Sprintf("survey-issue-%s-%d", repo, iss.Number),
			Title:     fmt.Sprintf("#%d %s", iss.Number, iss.Title),
			Content:   fmt.Sprintf("Issue: %s by %s", iss.Title, iss.User.Login),
			URL:       iss.HTMLURL,
			Timestamp: iss.CreatedAt,
			Metadata: map[string]string{
				"repo":        repo,
				"author":      iss.User.Login,
				"subject_type": "Issue",
				"survey_type": "new_issue",
				"labels":      strings.Join(labels, ","),
			},
		})
	}

	return items, nil
}

// fetchOpenPRs はオープン中の PR を取得する
// 直近 30 日以内に更新されたものに限定する
func (p *Plugin) fetchOpenPRs(ctx context.Context, repo string) ([]plugin.Item, error) {
	since := time.Now().Add(-30 * 24 * time.Hour)

	output, err := runGH(ctx, "api",
		fmt.Sprintf("repos/%s/pulls?state=open&sort=updated&direction=desc&per_page=20", repo),
	)
	if err != nil {
		return nil, fmt.Errorf("gh api 失敗: %w", err)
	}

	var prs []ghPullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("PR JSON パース失敗: %w", err)
	}

	var items []plugin.Item
	for _, pr := range prs {
		if pr.UpdatedAt.Before(since) {
			continue
		}

		labels := make([]string, len(pr.Labels))
		for i, l := range pr.Labels {
			labels[i] = l.Name
		}

		items = append(items, plugin.Item{
			Source:    "git",
			SourceID:  fmt.Sprintf("survey-pr-open-%s-%d", repo, pr.Number),
			Title:     fmt.Sprintf("#%d %s", pr.Number, pr.Title),
			Content:   fmt.Sprintf("Open PR: %s by %s", pr.Title, pr.User.Login),
			URL:       pr.HTMLURL,
			Timestamp: pr.UpdatedAt,
			Metadata: map[string]string{
				"repo":         repo,
				"author":       pr.User.Login,
				"subject_type": "PullRequest",
				"survey_type":  "open_pr",
				"labels":       strings.Join(labels, ","),
			},
		})
	}

	return items, nil
}

// ghRelease は GitHub Release のレスポンス
type ghRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	Author      ghUser    `json:"author"`
	PublishedAt time.Time `json:"published_at"`
}

// fetchRecentReleases は直近のリリースを取得する
func (p *Plugin) fetchRecentReleases(ctx context.Context, repo string) ([]plugin.Item, error) {
	since := time.Now().Add(-30 * 24 * time.Hour)

	output, err := runGH(ctx, "api",
		fmt.Sprintf("repos/%s/releases?per_page=10", repo),
	)
	if err != nil {
		return nil, fmt.Errorf("gh api 失敗: %w", err)
	}

	var releases []ghRelease
	if err := json.Unmarshal(output, &releases); err != nil {
		return nil, fmt.Errorf("Release JSON パース失敗: %w", err)
	}

	var items []plugin.Item
	for _, r := range releases {
		if r.Draft {
			continue
		}
		if r.PublishedAt.Before(since) {
			continue
		}

		displayName := r.Name
		if displayName == "" {
			displayName = r.TagName
		}

		items = append(items, plugin.Item{
			Source:    "git",
			SourceID:  fmt.Sprintf("survey-release-%s-%s", repo, r.TagName),
			Title:     fmt.Sprintf("Release %s", displayName),
			Content:   fmt.Sprintf("Release %s by %s", r.TagName, r.Author.Login),
			URL:       r.HTMLURL,
			Timestamp: r.PublishedAt,
			Metadata: map[string]string{
				"repo":         repo,
				"author":       r.Author.Login,
				"subject_type": "Release",
				"survey_type":  "release",
				"tag":          r.TagName,
				"prerelease":   boolStr(r.Prerelease),
			},
		})
	}

	return items, nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
