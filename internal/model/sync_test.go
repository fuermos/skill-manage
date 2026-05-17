package model

import (
	"testing"
	"time"
)

func TestFileInfoHash(t *testing.T) {
	fi := FileInfo{
		Path:    "claude/skills/java-coding/SKILL.md",
		Tool:    "claude",
		Category: "skills",
		Content: []byte("# Java Coding Standards\n\n## Naming conventions"),
		Size:    49,
	}

	hash := fi.Hash()
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if len(hash) != 64 {
		t.Errorf("expected sha256 hash length 64, got %d", len(hash))
	}

	sameContent := FileInfo{
		Path:    "different/path.md",
		Content: []byte("# Java Coding Standards\n\n## Naming conventions"),
	}
	if hash != sameContent.Hash() {
		t.Error("same content should produce same hash")
	}

	differentContent := FileInfo{
		Content: []byte("different content"),
	}
	if hash == differentContent.Hash() {
		t.Error("different content should produce different hash")
	}
}

func TestFileInfoServerPath(t *testing.T) {
	tests := []struct {
		name string
		fi   FileInfo
		want string
	}{
		{
			name: "skill file",
			fi: FileInfo{
				Path:    "claude/skills/java-coding/SKILL.md",
				Tool:    "claude",
				Category: "skills",
			},
			want: "claude/skills/java-coding/SKILL.md",
		},
		{
			name: "rule file",
			fi: FileInfo{
				Path:    "claude/rules/common/coding-style.md",
				Tool:    "claude",
				Category: "rules",
			},
			want: "claude/rules/common/coding-style.md",
		},
		{
			name: "opencode config",
			fi: FileInfo{
				Path:    "opencode/opencode.jsonc",
				Tool:    "opencode",
				Category: "config",
			},
			want: "opencode/opencode.jsonc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fi.ServerPath()
			if got != tt.want {
				t.Errorf("ServerPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSyncStatusString(t *testing.T) {
	tests := []struct {
		status SyncStatus
		want   string
	}{
		{SyncStatusSynced, "synced"},
		{SyncStatusModified, "modified"},
		{SyncStatusNew, "new"},
		{SyncStatusDeleted, "deleted"},
		{SyncStatusConflict, "conflict"},
		{SyncStatus(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("SyncStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestSyncStateIsUpToDate(t *testing.T) {
	state := SyncState{
		Tool:            "claude",
		LastRevision:    42,
		LastSyncedAt:    time.Now(),
	}

	if !state.IsUpToDate(42) {
		t.Error("IsUpToDate(42) should be true when last_revision matches")
	}
	if state.IsUpToDate(43) {
		t.Error("IsUpToDate(43) should be false when server revision is newer")
	}
	if state.IsUpToDate(41) {
		t.Error("IsUpToDate(41) should be false when server revision is older")
	}
}

func TestChangeTypeString(t *testing.T) {
	tests := []struct {
		ct   ChangeType
		want string
	}{
		{ChangeTypeCreate, "create"},
		{ChangeTypeModify, "modify"},
		{ChangeTypeDelete, "delete"},
		{ChangeType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.ct.String()
			if got != tt.want {
				t.Errorf("ChangeType(%d).String() = %q, want %q", tt.ct, got, tt.want)
			}
		})
	}
}

func TestChangelogEntryValidate(t *testing.T) {
	valid := ChangelogEntry{
		Revision:   42,
		Path:       "claude/skills/test/SKILL.md",
		Action:     ChangeTypeCreate,
		Hash:       "abc123",
		ClientID:   "win-pc",
		CreatedAt:  time.Now(),
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid entry should not error: %v", err)
	}

	invalid := ChangelogEntry{
		Path: "",
	}
	if err := invalid.Validate(); err == nil {
		t.Error("entry with empty path should error")
	}

	invalidAction := ChangelogEntry{
		Path:   "test.md",
		Action: ChangeType(99),
	}
	if err := invalidAction.Validate(); err == nil {
		t.Error("entry with invalid action should error")
	}
}

func TestPushRequestBaseRevision(t *testing.T) {
	req := PushRequest{
		BaseRevision: 40,
		ClientID:     "win-pc",
		Changes: []FileChange{
			{
				Path:    "claude/skills/test/SKILL.md",
				Hash:    "abc123",
				Action:  ChangeTypeCreate,
				Content: []byte("test content"),
			},
		},
	}

	if req.BaseRevision != 40 {
		t.Errorf("BaseRevision = %d, want 40", req.BaseRevision)
	}
	if len(req.Changes) != 1 {
		t.Errorf("len(Changes) = %d, want 1", len(req.Changes))
	}
}

func TestPushResponseHasConflicts(t *testing.T) {
	resp := PushResponse{
		NewRevision: 43,
		Applied:     2,
		Conflicts:   []Conflict{},
	}
	if resp.HasConflicts() {
		t.Error("empty conflicts should return false")
	}

	resp.Conflicts = []Conflict{
		{Path: "test.md", ServerVersion: 42},
	}
	if !resp.HasConflicts() {
		t.Error("non-empty conflicts should return true")
	}
}

func TestConflictMessage(t *testing.T) {
	c := Conflict{
		Path:          "claude/skills/java-coding/SKILL.md",
		ServerVersion: 42,
		LocalHash:     "abc",
		ServerHash:    "def",
	}

	msg := c.Message()
	if msg == "" {
		t.Error("conflict message should not be empty")
	}
	if !contains(msg, "claude/skills/java-coding/SKILL.md") {
		t.Error("conflict message should contain the file path")
	}
	if !contains(msg, "42") {
		t.Error("conflict message should contain the server version")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}