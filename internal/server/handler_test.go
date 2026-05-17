package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"skill-manage/internal/model"
	"skill-manage/internal/store"
	"skill-manage/internal/auth"

	_ "modernc.org/sqlite"
)

func setupTestServer(t *testing.T) (*Handler, func()) {
	t.Helper()

	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}

	syncStore := store.NewSyncStore(db)
	skillStore := store.NewSkillStore(db)

	tokenAuth := auth.NewTokenAuth("test-token")
	handler := NewHandler(syncStore, skillStore, tokenAuth)

	return handler, func() { db.Close() }
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "skill-sync-test-*")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func authHeader() string {
	return "Bearer test-token"
}

func TestHealthEndpoint(t *testing.T) {
	handler, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

func TestGetRevisionFreshServer(t *testing.T) {
	handler, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/sync/status", nil)
	req.Header.Set("Authorization", authHeader())
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var info model.RevisionInfo
	json.NewDecoder(w.Body).Decode(&info)
	if info.Revision != 0 {
		t.Errorf("initial revision should be 0, got %d", info.Revision)
	}
}

func TestPushNewFiles(t *testing.T) {
	handler, cleanup := setupTestServer(t)
	defer cleanup()

	fi := model.FileInfo{
		Path:     "claude/skills/test/SKILL.md",
		Tool:     "claude",
		Category: "skills",
		Content:  []byte("# Test Skill"),
	}

	pushReq := model.PushRequest{
		BaseRevision: 0,
		ClientID:     "test-pc",
		Changes: []model.FileChange{
			{
				Path:    fi.Path,
				Hash:    fi.Hash(),
				Action:  model.ChangeTypeCreate,
				Content: fi.Content,
			},
		},
	}

	body, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/api/v1/sync/push", bytes.NewReader(body))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.PushResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.NewRevision != 1 {
		t.Errorf("expected revision 1, got %d", resp.NewRevision)
	}
	if resp.Applied != 1 {
		t.Errorf("expected 1 applied, got %d", resp.Applied)
	}
	if resp.HasConflicts() {
		t.Error("unexpected conflicts")
	}

	req = httptest.NewRequest("GET", "/api/v1/sync/status", nil)
	req.Header.Set("Authorization", authHeader())
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var info model.RevisionInfo
	json.NewDecoder(w.Body).Decode(&info)
	if info.Revision != 1 {
		t.Errorf("revision after push should be 1, got %d", info.Revision)
	}
}

func TestPushConflictDetection(t *testing.T) {
	handler, cleanup := setupTestServer(t)
	defer cleanup()

	push1 := model.PushRequest{
		BaseRevision: 0,
		ClientID:     "pc-1",
		Changes: []model.FileChange{
			{Path: "shared.md", Hash: "hash1", Action: model.ChangeTypeCreate, Content: []byte("pc-1 content")},
		},
	}
	body, _ := json.Marshal(push1)
	req := httptest.NewRequest("POST", "/api/v1/sync/push", bytes.NewReader(body))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("first push failed: %d", w.Code)
	}

	push2 := model.PushRequest{
		BaseRevision: 0,
		ClientID:     "pc-2",
		Changes: []model.FileChange{
			{Path: "shared.md", Hash: "hash2", Action: model.ChangeTypeModify, Content: []byte("pc-2 content")},
		},
	}
	body, _ = json.Marshal(push2)
	req = httptest.NewRequest("POST", "/api/v1/sync/push", bytes.NewReader(body))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("second push failed: %d", w.Code)
	}

	var resp model.PushResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.HasConflicts() {
		t.Error("expected conflicts when base_revision is stale")
	}
	if len(resp.Conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(resp.Conflicts))
	}
}

func TestPullChanges(t *testing.T) {
	handler, cleanup := setupTestServer(t)
	defer cleanup()

	pushReq := model.PushRequest{
		BaseRevision: 0,
		ClientID:     "pc-1",
		Changes: []model.FileChange{
			{Path: "claude/skills/a/SKILL.md", Hash: "hash-a", Action: model.ChangeTypeCreate, Content: []byte("skill a")},
			{Path: "claude/skills/b/SKILL.md", Hash: "hash-b", Action: model.ChangeTypeCreate, Content: []byte("skill b")},
		},
	}
	body, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/api/v1/sync/push", bytes.NewReader(body))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	req = httptest.NewRequest("GET", "/api/v1/sync/changes?since=0", nil)
	req.Header.Set("Authorization", authHeader())
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var changesResp model.ChangesResponse
	json.NewDecoder(w.Body).Decode(&changesResp)
	if len(changesResp.Changes) != 2 {
		t.Errorf("expected 2 changes, got %d", len(changesResp.Changes))
	}

	pullReq := model.PullRequest{
		ClientID: "pc-2",
		LocalStates: []model.FileStateEntry{
			{Path: "claude/skills/a/SKILL.md", LocalHash: "", LocalMtime: 0},
		},
	}
	body, _ = json.Marshal(pullReq)
	req = httptest.NewRequest("POST", "/api/v1/sync/pull", bytes.NewReader(body))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var pullResp model.PullResponse
	json.NewDecoder(w.Body).Decode(&pullResp)
	if len(pullResp.Changes) < 1 {
		t.Errorf("expected at least 1 change in pull, got %d", len(pullResp.Changes))
	}
}

func TestUnauthorizedRequest(t *testing.T) {
	handler, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/sync/status"},
		{"POST", "/api/v1/sync/push"},
		{"POST", "/api/v1/sync/pull"},
		{"GET", "/api/v1/skills"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", w.Code)
			}
		})
	}
}

func TestInvalidToken(t *testing.T) {
	handler, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/sync/status", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestPostFileDownload(t *testing.T) {
	handler, cleanup := setupTestServer(t)
	defer cleanup()

	fi := model.FileInfo{
		Path:     "claude/skills/test/SKILL.md",
		Tool:     "claude",
		Category: "skills",
		Content:  []byte("# Test\n## Section\ncontent here"),
	}
	pushReq := model.PushRequest{
		BaseRevision: 0,
		ClientID:     "pc-1",
		Changes: []model.FileChange{
			{Path: fi.Path, Hash: fi.Hash(), Action: model.ChangeTypeCreate, Content: fi.Content},
		},
	}
	body, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/api/v1/sync/push", bytes.NewReader(body))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	fileReq := map[string][]string{
		"paths": {fi.Path},
	}
	body, _ = json.Marshal(fileReq)
	req = httptest.NewRequest("POST", "/api/v1/sync/files", bytes.NewReader(body))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCORSHeaders(t *testing.T) {
	handler, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("OPTIONS", "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("missing CORS header")
	}
}

func init() {
	_ = time.Now
}