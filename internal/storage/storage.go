/**
 * SQLite ストレージ層
 * ラベリング済みアイテムの永続化とクエリを担当する
 */
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/senna-lang/sentei/internal/plugin"
)

// Storage は SQLite ベースのデータストア
type Storage struct {
	db *sql.DB
}

// New は SQLite ストレージを初期化する
func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("SQLite オープン失敗: %w", err)
	}

	// WAL モード有効化（並行読み書き改善）
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("WAL モード設定失敗: %w", err)
	}

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("スキーマ適用失敗: %w", err)
	}
	return s, nil
}

// migrate はスキーマを適用する
func (s *Storage) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source TEXT NOT NULL,
		source_id TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT,
		url TEXT,
		timestamp DATETIME NOT NULL,
		metadata JSON,

		-- Bonsai ラベリング結果
		urgency TEXT,
		category TEXT,
		summary TEXT,
		labeled_at DATETIME,

		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

		UNIQUE(source, source_id)
	);

	CREATE INDEX IF NOT EXISTS idx_items_urgency ON items(urgency, timestamp);
	CREATE INDEX IF NOT EXISTS idx_items_source ON items(source, timestamp);
	CREATE INDEX IF NOT EXISTS idx_items_category ON items(category, timestamp);
	`
	_, err := s.db.Exec(schema)
	return err
}

// SaveLabeledItem はラベリング済みアイテムを保存する（UPSERT）
func (s *Storage) SaveLabeledItem(li plugin.LabeledItem) error {
	metadataJSON, err := json.Marshal(li.Item.Metadata)
	if err != nil {
		return fmt.Errorf("metadata JSON 変換失敗: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO items (source, source_id, title, content, url, timestamp, metadata, urgency, category, summary, labeled_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(source, source_id) DO UPDATE SET
			title = excluded.title,
			content = excluded.content,
			url = excluded.url,
			urgency = excluded.urgency,
			category = excluded.category,
			summary = excluded.summary,
			labeled_at = excluded.labeled_at
	`,
		li.Item.Source,
		li.Item.SourceID,
		li.Item.Title,
		li.Item.Content,
		li.Item.URL,
		li.Item.Timestamp,
		string(metadataJSON),
		string(li.Label.Urgency),
		li.Label.Category,
		li.Label.Summary,
		li.LabeledAt,
	)
	return err
}

// Exists はアイテムが既に存在するか確認する（冪等性チェック）
func (s *Storage) Exists(source, sourceID string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM items WHERE source = ? AND source_id = ?", source, sourceID).Scan(&count)
	return count > 0, err
}

// ItemFilter はアイテム検索のフィルタ条件
type ItemFilter struct {
	Urgency  string
	Source   string
	Category string
	Limit    int
}

// ListItems はフィルタ条件に合うアイテムを取得する
func (s *Storage) ListItems(filter ItemFilter) ([]plugin.LabeledItem, error) {
	query := "SELECT source, source_id, title, content, url, timestamp, metadata, urgency, category, summary, labeled_at FROM items WHERE 1=1"
	args := []any{}

	if filter.Urgency != "" {
		query += " AND urgency = ?"
		args = append(args, filter.Urgency)
	}
	if filter.Source != "" {
		query += " AND source = ?"
		args = append(args, filter.Source)
	}
	if filter.Category != "" {
		query += " AND category = ?"
		args = append(args, filter.Category)
	}

	query += " ORDER BY CASE urgency WHEN 'urgent' THEN 0 WHEN 'should_check' THEN 1 WHEN 'can_wait' THEN 2 WHEN 'ignore' THEN 3 END, timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []plugin.LabeledItem
	for rows.Next() {
		var li plugin.LabeledItem
		var metadataJSON string
		var urgency, category, summary string
		var content, url sql.NullString
		var labeledAt sql.NullTime

		err := rows.Scan(
			&li.Item.Source,
			&li.Item.SourceID,
			&li.Item.Title,
			&content,
			&url,
			&li.Item.Timestamp,
			&metadataJSON,
			&urgency,
			&category,
			&summary,
			&labeledAt,
		)
		if err != nil {
			return nil, err
		}

		li.Item.Content = content.String
		li.Item.URL = url.String
		li.Label.Urgency = plugin.Urgency(urgency)
		li.Label.Category = category
		li.Label.Summary = summary
		if labeledAt.Valid {
			li.LabeledAt = labeledAt.Time
		}

		if metadataJSON != "" {
			json.Unmarshal([]byte(metadataJSON), &li.Item.Metadata)
		}

		items = append(items, li)
	}
	return items, rows.Err()
}

// ItemCount はアイテム数を返す
func (s *Storage) ItemCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count)
	return count, err
}

// LastLabeledAt は最後にラベリングされた時刻を返す
func (s *Storage) LastLabeledAt() (time.Time, error) {
	var t sql.NullTime
	err := s.db.QueryRow("SELECT MAX(labeled_at) FROM items").Scan(&t)
	if t.Valid {
		return t.Time, err
	}
	return time.Time{}, err
}

// DeleteItem はアイテムを物理削除する
func (s *Storage) DeleteItem(source, sourceID string) (bool, error) {
	result, err := s.db.Exec("DELETE FROM items WHERE source = ? AND source_id = ?", source, sourceID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// Close はデータベース接続を閉じる
func (s *Storage) Close() error {
	return s.db.Close()
}
