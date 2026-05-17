package store

import (
	"database/sql"
	"testing"
	"time"

	"skill-manage/internal/model"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestMigrateCreatesTables(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	expectedTables := []string{"meta", "changelog", "files", "skills", "skill_examples",
		"error_cases", "error_clusters", "skill_combinations", "skill_chains", "skill_usage_log"}

	for _, table := range expectedTables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("failed to check table %s: %v", table, err)
		}
		if count == 0 {
			t.Errorf("table %s should exist", table)
		}
	}
}

func TestMigrateIdempotent(t *testing.T) {
	db := setupTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("second migration should not error: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("third migration should not error: %v", err)
	}
	db.Close()
}

func TestRevisionGetSet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSyncStore(db)

	rev, err := store.GetRevision()
	if err != nil {
		t.Fatalf("GetRevision: %v", err)
	}
	if rev != 0 {
		t.Errorf("initial revision should be 0, got %d", rev)
	}

	newRev, err := store.IncrementRevision()
	if err != nil {
		t.Fatalf("IncrementRevision: %v", err)
	}
	if newRev != 1 {
		t.Errorf("expected revision 1, got %d", newRev)
	}

	rev, _ = store.GetRevision()
	if rev != 1 {
		t.Errorf("expected revision 1 after increment, got %d", rev)
	}
}

func TestFileInsertAndGet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSyncStore(db)

	f := model.FileInfo{
		Path:    "claude/skills/test/SKILL.md",
		Tool:    "claude",
		Category: "skills",
		Content: []byte("# Test Skill"),
	}
	if err := store.UpsertFile(f, 1); err != nil {
		t.Fatalf("UpsertFile: %v", err)
	}

	got, err := store.GetFile(f.Path)
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if got == nil {
		t.Fatal("expected file to exist")
	}
	if got.ContentHash != f.Hash() {
		t.Errorf("hash mismatch: expected %s, got %s", f.Hash(), got.ContentHash)
	}
	if got.Tool != f.Tool {
		t.Errorf("tool mismatch: expected %s, got %s", f.Tool, got.Tool)
	}
	if got.Version != 1 {
		t.Errorf("version should be 1, got %d", got.Version)
	}
}

func TestFileUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSyncStore(db)

	f := model.FileInfo{
		Path:    "claude/rules/test.md",
		Tool:    "claude",
		Category: "rules",
		Content: []byte("original"),
	}
	if err := store.UpsertFile(f, 1); err != nil {
		t.Fatalf("first UpsertFile: %v", err)
	}

	f.Content = []byte("updated")
	if err := store.UpsertFile(f, 2); err != nil {
		t.Fatalf("second UpsertFile: %v", err)
	}

	got, err := store.GetFile(f.Path)
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if got.ContentHash != f.Hash() {
		t.Error("hash should reflect updated content")
	}
	if got.Version != 2 {
		t.Errorf("version should be 2, got %d", got.Version)
	}
}

func TestFileDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSyncStore(db)

	f := model.FileInfo{
		Path:    "to-delete.md",
		Tool:    "claude",
		Category: "rules",
		Content: []byte("delete me"),
	}
	if err := store.UpsertFile(f, 1); err != nil {
		t.Fatalf("UpsertFile: %v", err)
	}
	if err := store.DeleteFile(f.Path, 2); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}

	got, _ := store.GetFile(f.Path)
	if got != nil {
		t.Error("deleted file should not exist")
	}
}

func TestListFiles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSyncStore(db)

	files := []model.FileInfo{
		{Path: "claude/skills/a/SKILL.md", Tool: "claude", Category: "skills", Content: []byte("a")},
		{Path: "claude/skills/b/SKILL.md", Tool: "claude", Category: "skills", Content: []byte("b")},
		{Path: "claude/rules/common/x.md", Tool: "claude", Category: "rules", Content: []byte("x")},
		{Path: "opencode/config.jsonc", Tool: "opencode", Category: "config", Content: []byte("{}")},
	}
	for _, f := range files {
		if err := store.UpsertFile(f, 1); err != nil {
			t.Fatalf("UpsertFile %s: %v", f.Path, err)
		}
	}

	all, err := store.ListFiles("", "")
	if err != nil {
		t.Fatalf("ListFiles(all): %v", err)
	}
	if len(all) != 4 {
		t.Errorf("expected 4 files, got %d", len(all))
	}

	claudeOnly, _ := store.ListFiles("claude", "")
	if len(claudeOnly) != 3 {
		t.Errorf("expected 3 claude files, got %d", len(claudeOnly))
	}

	opencodeOnly, _ := store.ListFiles("opencode", "")
	if len(opencodeOnly) != 1 {
		t.Errorf("expected 1 opencode file, got %d", len(opencodeOnly))
	}

	skillsOnly, _ := store.ListFiles("claude", "skills")
	if len(skillsOnly) != 2 {
		t.Errorf("expected 2 claude skills, got %d", len(skillsOnly))
	}
}

func TestChangelogWriteAndRead(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSyncStore(db)

	entries := []model.ChangelogEntry{
		{Revision: 1, Path: "a.md", Action: model.ChangeTypeCreate, Hash: "aaa", ClientID: "pc-1"},
		{Revision: 2, Path: "b.md", Action: model.ChangeTypeModify, Hash: "bbb", ClientID: "pc-1"},
		{Revision: 3, Path: "c.md", Action: model.ChangeTypeDelete, Hash: "ccc", ClientID: "pc-2"},
	}
	for _, e := range entries {
		if err := store.WriteChangelog(e); err != nil {
			t.Fatalf("WriteChangelog: %v", err)
		}
	}

	changes, err := store.GetChangesSince(0)
	if err != nil {
		t.Fatalf("GetChangesSince(0): %v", err)
	}
	if len(changes) != 3 {
		t.Errorf("expected 3 changes, got %d", len(changes))
	}

	changes, _ = store.GetChangesSince(2)
	if len(changes) != 1 {
		t.Errorf("expected 1 change since rev 2, got %d", len(changes))
	}

	changes, _ = store.GetChangesSince(3)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes since rev 3, got %d", len(changes))
	}
}

func TestBatchUpsertFilesAndGetChanges(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSyncStore(db)

	files := []model.FileInfo{
		{Path: "batch/a.md", Tool: "claude", Category: "skills", Content: []byte("a")},
		{Path: "batch/b.md", Tool: "claude", Category: "skills", Content: []byte("b")},
		{Path: "batch/c.md", Tool: "claude", Category: "rules", Content: []byte("c")},
	}
	if err := store.BatchUpsertFiles(files, 1, "test-pc"); err != nil {
		t.Fatalf("BatchUpsertFiles: %v", err)
	}

	changes, _ := store.GetChangesSince(0)
	if len(changes) != 3 {
		t.Errorf("expected 3 changelog entries, got %d", len(changes))
	}

	all, _ := store.ListFiles("", "")
	if len(all) != 3 {
		t.Errorf("expected 3 files, got %d", len(all))
	}
}

func TestGetFilesByPath(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSyncStore(db)

	files := []model.FileInfo{
		{Path: "p1.md", Tool: "claude", Category: "skills", Content: []byte("p1")},
		{Path: "p2.md", Tool: "claude", Category: "skills", Content: []byte("p2")},
	}
	if err := store.BatchUpsertFiles(files, 1, "test"); err != nil {
		t.Fatalf("BatchUpsertFiles: %v", err)
	}

	result, err := store.GetFilesByPaths([]string{"p1.md", "p2.md", "nonexistent.md"})
	if err != nil {
		t.Fatalf("GetFilesByPaths: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 files, got %d", len(result))
	}

	found := make(map[string]bool)
	for _, f := range result {
		found[f.Path] = true
	}
	if !found["p1.md"] || !found["p2.md"] {
		t.Error("missing expected files in result")
	}
}

func TestRevisionUpdateTime(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewSyncStore(db)

	info, err := store.GetRevisionInfo()
	if err != nil {
		t.Fatalf("GetRevisionInfo: %v", err)
	}
	if info.Revision != 0 {
		t.Errorf("initial revision should be 0, got %d", info.Revision)
	}
	if info.UpdatedAt.IsZero() {
		t.Error("updated_at should not be zero")
	}

	time.Sleep(1500 * time.Millisecond)
	store.IncrementRevision()

	info2, _ := store.GetRevisionInfo()
	if !info2.UpdatedAt.After(info.UpdatedAt) {
		t.Error("updated_at should be newer after increment")
	}
}