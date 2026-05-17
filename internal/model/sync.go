package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

type SyncStatus int

const (
	SyncStatusSynced   SyncStatus = 1
	SyncStatusModified SyncStatus = 2
	SyncStatusNew      SyncStatus = 3
	SyncStatusDeleted  SyncStatus = 4
	SyncStatusConflict SyncStatus = 5
)

var syncStatusNames = map[SyncStatus]string{
	SyncStatusSynced:   "synced",
	SyncStatusModified: "modified",
	SyncStatusNew:      "new",
	SyncStatusDeleted:  "deleted",
	SyncStatusConflict: "conflict",
}

func (s SyncStatus) String() string {
	if name, ok := syncStatusNames[s]; ok {
		return name
	}
	return "unknown"
}

type ChangeType int

const (
	ChangeTypeCreate ChangeType = 1
	ChangeTypeModify ChangeType = 2
	ChangeTypeDelete ChangeType = 3
)

var changeTypeNames = map[ChangeType]string{
	ChangeTypeCreate: "create",
	ChangeTypeModify: "modify",
	ChangeTypeDelete: "delete",
}

func (ct ChangeType) String() string {
	if name, ok := changeTypeNames[ct]; ok {
		return name
	}
	return "unknown"
}

func (ct ChangeType) Valid() bool {
	_, ok := changeTypeNames[ct]
	return ok
}

type FileInfo struct {
	Path        string
	Tool        string
	Category    string
	Content     []byte
	ContentHash string
	Size        int64
	Version     int64

	// Parsed from SKILL.md frontmatter
	MetaName        string
	MetaDescription string
	MetaSummary     string
	MetaTags        []string
}

func (f FileInfo) Hash() string {
	h := sha256.Sum256(f.Content)
	return hex.EncodeToString(h[:])
}

func (f FileInfo) ServerPath() string {
	return f.Path
}

type SyncState struct {
	Tool         string
	LastRevision int64
	LastSyncedAt time.Time
}

func (s SyncState) IsUpToDate(serverRevision int64) bool {
	return s.LastRevision == serverRevision
}

type ChangelogEntry struct {
	Revision  int64
	Path      string
	Action    ChangeType
	Hash      string
	ClientID  string
	CreatedAt time.Time
}

func (e ChangelogEntry) Validate() error {
	if e.Path == "" {
		return errors.New("path is required")
	}
	if !e.Action.Valid() {
		return fmt.Errorf("invalid action type: %d", e.Action)
	}
	return nil
}

type FileChange struct {
	Path    string     `json:"path"`
	Hash    string     `json:"hash"`
	Action  ChangeType `json:"action"`
	Content []byte     `json:"content,omitempty"`
}

type PushRequest struct {
	BaseRevision int64        `json:"base_revision"`
	ClientID     string       `json:"client_id"`
	Changes      []FileChange `json:"changes"`
}

type Conflict struct {
	Path          string `json:"path"`
	ServerVersion int64  `json:"server_version"`
	LocalHash     string `json:"local_hash"`
	ServerHash    string `json:"server_hash"`
}

func (c Conflict) Message() string {
	localShort := c.LocalHash
	if len(localShort) > 8 {
		localShort = localShort[:8]
	}
	serverShort := c.ServerHash
	if len(serverShort) > 8 {
		serverShort = serverShort[:8]
	}
	return fmt.Sprintf("conflict in %s (server version: %d, local hash: %s, server hash: %s)",
		c.Path, c.ServerVersion, localShort, serverShort)
}

type PushResponse struct {
	NewRevision int64      `json:"new_revision"`
	Applied     int        `json:"applied"`
	Conflicts   []Conflict `json:"conflicts"`
}

func (r PushResponse) HasConflicts() bool {
	return len(r.Conflicts) > 0
}

type PullRequest struct {
	ClientID     string           `json:"client_id"`
	LocalStates  []FileStateEntry `json:"local_states"`
}

type FileStateEntry struct {
	Path       string `json:"path"`
	LocalHash  string `json:"local_hash"`
	LocalMtime int64  `json:"local_mtime"`
}

type PullResponse struct {
	Revision int64        `json:"revision"`
	Changes  []FileChange `json:"changes"`
}

type FileState struct {
	Path          string
	LocalHash     string
	ServerVersion int64
	Status        SyncStatus
	LocalMtime    int64
	SyncedAt      time.Time
}

type RevisionInfo struct {
	Revision  int64     `json:"revision"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChangesResponse struct {
	Revision int64            `json:"revision"`
	Changes  []ChangelogEntry `json:"changes"`
}