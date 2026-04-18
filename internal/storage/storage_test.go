/**
 * SQLite ストレージ層のテスト
 */
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/senna-lang/sentei/internal/plugin"
)

func newTestStorage(t *testing.T) *Storage {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")
	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("ストレージ初期化失敗: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestNew_CreatesDatabase(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer s.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("データベースファイルが作成されていない")
	}
}

func TestSaveLabeledItem_AndListItems(t *testing.T) {
	s := newTestStorage(t)

	li := plugin.LabeledItem{
		Item: plugin.Item{
			Source:    "git",
			SourceID: "notif-001",
			Title:    "Review request from mentor",
			URL:      "https://github.com/test/pr/1",
			Timestamp: time.Now(),
			Metadata: map[string]string{
				"repo":              "arxiv-compass",
				"notification_type": "review_requested",
				"author":            "mentor",
			},
		},
		Label: plugin.Label{
			Urgency:  plugin.UrgencyUrgent,
			Category: "pr",
			Summary:  "Mentor requested review",
		},
		LabeledAt: time.Now(),
	}

	if err := s.SaveLabeledItem(li); err != nil {
		t.Fatalf("SaveLabeledItem() error = %v", err)
	}

	items, err := s.ListItems(ItemFilter{})
	if err != nil {
		t.Fatalf("ListItems() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListItems() len = %d, want 1", len(items))
	}

	got := items[0]
	if got.Item.Source != "git" {
		t.Errorf("Source = %q, want %q", got.Item.Source, "git")
	}
	if got.Label.Urgency != plugin.UrgencyUrgent {
		t.Errorf("Urgency = %q, want %q", got.Label.Urgency, plugin.UrgencyUrgent)
	}
	if got.Label.Category != "pr" {
		t.Errorf("Category = %q, want %q", got.Label.Category, "pr")
	}
	if got.Item.Metadata["repo"] != "arxiv-compass" {
		t.Errorf("Metadata[repo] = %q, want %q", got.Item.Metadata["repo"], "arxiv-compass")
	}
}

func TestSaveLabeledItem_Upsert(t *testing.T) {
	s := newTestStorage(t)

	li := plugin.LabeledItem{
		Item: plugin.Item{
			Source:    "git",
			SourceID: "notif-001",
			Title:    "Original title",
			Timestamp: time.Now(),
			Metadata: map[string]string{},
		},
		Label: plugin.Label{
			Urgency:  plugin.UrgencyCanWait,
			Category: "pr",
		},
		LabeledAt: time.Now(),
	}

	if err := s.SaveLabeledItem(li); err != nil {
		t.Fatalf("SaveLabeledItem() error = %v", err)
	}

	// 同じ source + source_id で更新
	li.Item.Title = "Updated title"
	li.Label.Urgency = plugin.UrgencyUrgent
	if err := s.SaveLabeledItem(li); err != nil {
		t.Fatalf("SaveLabeledItem() upsert error = %v", err)
	}

	items, _ := s.ListItems(ItemFilter{})
	if len(items) != 1 {
		t.Fatalf("upsert 後の len = %d, want 1", len(items))
	}
	if items[0].Item.Title != "Updated title" {
		t.Errorf("Title = %q, want %q", items[0].Item.Title, "Updated title")
	}
	if items[0].Label.Urgency != plugin.UrgencyUrgent {
		t.Errorf("Urgency = %q, want %q", items[0].Label.Urgency, plugin.UrgencyUrgent)
	}
}

func TestExists(t *testing.T) {
	s := newTestStorage(t)

	exists, err := s.Exists("git", "notif-999")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Error("存在しないアイテムが Exists=true")
	}

	li := plugin.LabeledItem{
		Item: plugin.Item{
			Source:    "git",
			SourceID: "notif-999",
			Title:    "test",
			Timestamp: time.Now(),
			Metadata: map[string]string{},
		},
		Label: plugin.Label{
			Urgency:  plugin.UrgencyShouldCheck,
			Category: "issue",
		},
		LabeledAt: time.Now(),
	}
	s.SaveLabeledItem(li)

	exists, err = s.Exists("git", "notif-999")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("保存済みアイテムが Exists=false")
	}
}

func TestListItems_FilterByUrgency(t *testing.T) {
	s := newTestStorage(t)

	// urgent 1件 + should_check 2件
	for i, u := range []plugin.Urgency{plugin.UrgencyUrgent, plugin.UrgencyShouldCheck, plugin.UrgencyShouldCheck} {
		s.SaveLabeledItem(plugin.LabeledItem{
			Item: plugin.Item{
				Source:    "git",
				SourceID: fmt.Sprintf("n-%d", i),
				Title:    fmt.Sprintf("item %d", i),
				Timestamp: time.Now(),
				Metadata: map[string]string{},
			},
			Label:     plugin.Label{Urgency: u, Category: "pr"},
			LabeledAt: time.Now(),
		})
	}

	items, _ := s.ListItems(ItemFilter{Urgency: "urgent"})
	if len(items) != 1 {
		t.Errorf("urgent フィルタ len = %d, want 1", len(items))
	}

	items, _ = s.ListItems(ItemFilter{Urgency: "should_check"})
	if len(items) != 2 {
		t.Errorf("should_check フィルタ len = %d, want 2", len(items))
	}
}

func TestListItems_OrderByUrgency(t *testing.T) {
	s := newTestStorage(t)

	// can_wait → urgent → should_check の順で保存
	urgencies := []plugin.Urgency{plugin.UrgencyCanWait, plugin.UrgencyUrgent, plugin.UrgencyShouldCheck}
	for i, u := range urgencies {
		s.SaveLabeledItem(plugin.LabeledItem{
			Item: plugin.Item{
				Source:    "git",
				SourceID: fmt.Sprintf("n-%d", i),
				Title:    fmt.Sprintf("item %d", i),
				Timestamp: time.Now(),
				Metadata: map[string]string{},
			},
			Label:     plugin.Label{Urgency: u, Category: "pr"},
			LabeledAt: time.Now(),
		})
	}

	items, _ := s.ListItems(ItemFilter{})
	if len(items) != 3 {
		t.Fatalf("len = %d, want 3", len(items))
	}

	// urgent → should_check → can_wait の順で返るべき
	expectedOrder := []plugin.Urgency{plugin.UrgencyUrgent, plugin.UrgencyShouldCheck, plugin.UrgencyCanWait}
	for i, expected := range expectedOrder {
		if items[i].Label.Urgency != expected {
			t.Errorf("items[%d].Urgency = %q, want %q", i, items[i].Label.Urgency, expected)
		}
	}
}

func TestItemCount(t *testing.T) {
	s := newTestStorage(t)

	count, _ := s.ItemCount()
	if count != 0 {
		t.Errorf("初期 ItemCount = %d, want 0", count)
	}

	s.SaveLabeledItem(plugin.LabeledItem{
		Item: plugin.Item{
			Source: "git", SourceID: "1", Title: "test", Timestamp: time.Now(),
			Metadata: map[string]string{},
		},
		Label:     plugin.Label{Urgency: plugin.UrgencyCanWait, Category: "pr"},
		LabeledAt: time.Now(),
	})

	count, _ = s.ItemCount()
	if count != 1 {
		t.Errorf("保存後 ItemCount = %d, want 1", count)
	}
}

func TestDeleteItem_ExistingItem(t *testing.T) {
	s := newTestStorage(t)

	s.SaveLabeledItem(plugin.LabeledItem{
		Item: plugin.Item{
			Source: "git", SourceID: "del-1", Title: "to be deleted", Timestamp: time.Now(),
			Metadata: map[string]string{},
		},
		Label:     plugin.Label{Urgency: plugin.UrgencyUrgent, Category: "pr"},
		LabeledAt: time.Now(),
	})

	deleted, err := s.DeleteItem("git", "del-1")
	if err != nil {
		t.Fatalf("DeleteItem() error = %v", err)
	}
	if !deleted {
		t.Error("DeleteItem() should return true for existing item")
	}

	exists, _ := s.Exists("git", "del-1")
	if exists {
		t.Error("削除後のアイテムが Exists=true")
	}

	count, _ := s.ItemCount()
	if count != 0 {
		t.Errorf("削除後 ItemCount = %d, want 0", count)
	}
}

func TestDeleteItem_NonExistentItem(t *testing.T) {
	s := newTestStorage(t)

	deleted, err := s.DeleteItem("git", "nonexistent")
	if err != nil {
		t.Fatalf("DeleteItem() error = %v", err)
	}
	if deleted {
		t.Error("DeleteItem() should return false for non-existent item")
	}
}
