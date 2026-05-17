package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"skill-manage/internal/model"
)

type SyncEngine struct {
	serverURL string
	authToken string
	clientID  string
	client    *http.Client
}

func (e *SyncEngine) ServerURL() string { return e.serverURL }
func (e *SyncEngine) AuthToken() string { return e.authToken }

func NewSyncEngine(serverURL, authToken, clientID string) *SyncEngine {
	return &SyncEngine{
		serverURL: strings.TrimRight(serverURL, "/"),
		authToken: authToken,
		clientID:  clientID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (e *SyncEngine) get(path string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", e.serverURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+e.authToken)
	return e.client.Do(req)
}

func (e *SyncEngine) post(path string, body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", e.serverURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+e.authToken)
	req.Header.Set("Content-Type", "application/json")
	return e.client.Do(req)
}

func (e *SyncEngine) GetStatus() (*model.RevisionInfo, error) {
	resp, err := e.get("/api/v1/sync/status")
	if err != nil {
		return nil, fmt.Errorf("get status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status returned %d", resp.StatusCode)
	}

	var info model.RevisionInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (e *SyncEngine) Push(baseRevision int64, changes []model.FileChange) (*model.PushResponse, error) {
	req := model.PushRequest{
		BaseRevision: baseRevision,
		ClientID:     e.clientID,
		Changes:      changes,
	}

	resp, err := e.post("/api/v1/sync/push", req)
	if err != nil {
		return nil, fmt.Errorf("push request: %w", err)
	}
	defer resp.Body.Close()

	var pushResp model.PushResponse
	if err := json.NewDecoder(resp.Body).Decode(&pushResp); err != nil {
		return nil, err
	}
	return &pushResp, nil
}

func (e *SyncEngine) Pull(localStates []model.FileStateEntry) (*model.PullResponse, error) {
	req := model.PullRequest{
		ClientID:    e.clientID,
		LocalStates: localStates,
	}

	resp, err := e.post("/api/v1/sync/pull", req)
	if err != nil {
		return nil, fmt.Errorf("pull request: %w", err)
	}
	defer resp.Body.Close()

	var pullResp model.PullResponse
	if err := json.NewDecoder(resp.Body).Decode(&pullResp); err != nil {
		return nil, err
	}
	return &pullResp, nil
}

func (e *SyncEngine) DownloadFiles(paths []string) ([]model.FileInfo, error) {
	body := map[string][]string{"paths": paths}
	resp, err := e.post("/api/v1/sync/files", body)
	if err != nil {
		return nil, fmt.Errorf("download files: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Files []model.FileInfo `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Files, nil
}

func (e *SyncEngine) GetChanges(since int64) (*model.ChangesResponse, error) {
	resp, err := e.get(fmt.Sprintf("/api/v1/sync/changes?since=%d", since))
	if err != nil {
		return nil, fmt.Errorf("get changes: %w", err)
	}
	defer resp.Body.Close()

	var changes model.ChangesResponse
	if err := json.NewDecoder(resp.Body).Decode(&changes); err != nil {
		return nil, err
	}
	return &changes, nil
}

func ServerToLocalPath(serverPath, toolName string) string {
	prefix := toolName + "/"
	if strings.HasPrefix(serverPath, prefix) {
		return serverPath[len(prefix):]
	}
	return serverPath
}

func LocalToServerPath(localPath, toolName string) string {
	return filepath.ToSlash(filepath.Join(toolName, localPath))
}

func HandleSyncResult(resp *model.PushResponse) {
	if resp.HasConflicts() {
		fmt.Printf("⚠ Conflicts detected:\n")
		for _, c := range resp.Conflicts {
			fmt.Printf("  - %s\n", c.Message())
		}
	}
	fmt.Printf("Applied: %d changes (revision: %d)\n", resp.Applied, resp.NewRevision)
}

func init() {
	_ = io.Discard
}