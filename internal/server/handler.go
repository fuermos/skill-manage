package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"skill-manage/internal/auth"
	"skill-manage/internal/model"
	"skill-manage/internal/store"
)

type Handler struct {
	mux        *http.ServeMux
	syncStore  *store.SyncStore
	skillStore *store.SkillStore
	auth       *auth.TokenAuth
}

func NewHandler(syncStore *store.SyncStore, skillStore *store.SkillStore, tokenAuth *auth.TokenAuth) *Handler {
	h := &Handler{
		mux:        http.NewServeMux(),
		syncStore:  syncStore,
		skillStore: skillStore,
		auth:       tokenAuth,
	}

	mux := h.mux
	mux.HandleFunc("/api/v1/health", h.handleHealth)
	mux.HandleFunc("/api/v1/sync/status", h.auth.Wrap(h.handleSyncStatus))
	mux.HandleFunc("/api/v1/sync/push", h.auth.Wrap(h.handlePush))
	mux.HandleFunc("/api/v1/sync/pull", h.auth.Wrap(h.handlePull))
	mux.HandleFunc("/api/v1/sync/changes", h.auth.Wrap(h.handleChanges))
	mux.HandleFunc("/api/v1/sync/files", h.auth.Wrap(h.handleFiles))
	mux.HandleFunc("/api/v1/skills", h.auth.Wrap(h.handleSkillsList))
	mux.HandleFunc("/api/v1/skills/", h.auth.Wrap(h.handleSkillByID))
	mux.HandleFunc("/api/v1/usage/batch", h.auth.Wrap(h.handleUsageBatch))
	mux.HandleFunc("/api/v1/combinations", h.auth.Wrap(h.handleCombinations))
	mux.HandleFunc("/api/v1/chains", h.auth.Wrap(h.handleChains))

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	h.mux.ServeHTTP(w, r)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	info, err := h.syncStore.GetRevisionInfo()
	if err != nil {
		log.Printf("GetRevisionInfo error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get revision info")
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (h *Handler) handlePush(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	currentRev, err := h.syncStore.GetRevision()
	if err != nil {
		log.Printf("GetRevision error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if req.BaseRevision != currentRev {
		conflicts := h.detectConflicts(req.Changes, currentRev)
		resp := model.PushResponse{
			NewRevision: currentRev,
			Applied:     len(req.Changes) - len(conflicts),
			Conflicts:   conflicts,
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	newRev, err := h.syncStore.IncrementRevision()
	if err != nil {
		log.Printf("IncrementRevision error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to increment revision")
		return
	}

	applied := 0
	for _, change := range req.Changes {
		fi := model.FileInfo{
			Path:    change.Path,
			Tool:    extractTool(change.Path),
			Category: extractCategory(change.Path),
			Content: change.Content,
		}

		switch change.Action {
		case model.ChangeTypeCreate, model.ChangeTypeModify:
			if err := h.syncStore.UpsertFile(fi, newRev); err != nil {
				log.Printf("UpsertFile error: %v", err)
				continue
			}
		case model.ChangeTypeDelete:
			if err := h.syncStore.DeleteFile(change.Path, newRev); err != nil {
				log.Printf("DeleteFile error: %v", err)
				continue
			}
		}

		entry := model.ChangelogEntry{
			Revision: newRev,
			Path:     change.Path,
			Action:   change.Action,
			Hash:     change.Hash,
			ClientID: req.ClientID,
		}
		if err := h.syncStore.WriteChangelog(entry); err != nil {
			log.Printf("WriteChangelog error: %v", err)
			continue
		}
		applied++
	}

	writeJSON(w, http.StatusOK, model.PushResponse{
		NewRevision: newRev,
		Applied:     applied,
		Conflicts:   nil,
	})
}

func (h *Handler) detectConflicts(changes []model.FileChange, currentRev int64) []model.Conflict {
	var conflicts []model.Conflict
	for _, c := range changes {
		existing, err := h.syncStore.GetFile(c.Path)
		if err != nil || existing == nil {
			continue
		}
		if existing.ContentHash != c.Hash {
			conflicts = append(conflicts, model.Conflict{
				Path:          c.Path,
				ServerVersion: currentRev,
				LocalHash:     c.Hash,
				ServerHash:    existing.ContentHash,
			})
		}
	}
	return conflicts
}

func (h *Handler) handlePull(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.PullRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rev, err := h.syncStore.GetRevision()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	localHashes := make(map[string]string)
	for _, s := range req.LocalStates {
		localHashes[s.Path] = s.LocalHash
	}

	allFiles, err := h.syncStore.ListFiles("", "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list files")
		return
	}

	var changes []model.FileChange
	for _, f := range allFiles {
		localHash, exists := localHashes[f.Path]
		if !exists || localHash != f.ContentHash {
			changes = append(changes, model.FileChange{
				Path:   f.Path,
				Hash:   f.ContentHash,
				Action: model.ChangeTypeModify,
			})
		}
	}

	writeJSON(w, http.StatusOK, model.PullResponse{
		Revision: rev,
		Changes:  changes,
	})
}

func (h *Handler) handleChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sinceStr := r.URL.Query().Get("since")
	since, _ := strconv.ParseInt(sinceStr, 10, 64)

	changes, err := h.syncStore.GetChangesSince(since)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get changes")
		return
	}

	rev, _ := h.syncStore.GetRevision()
	writeJSON(w, http.StatusOK, model.ChangesResponse{
		Revision: rev,
		Changes:  changes,
	})
}

func (h *Handler) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Paths []string `json:"paths"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	files, err := h.syncStore.GetFilesByPaths(req.Paths)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get files")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"files": files,
	})
}

func (h *Handler) handleSkillsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tool := r.URL.Query().Get("tool")
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")

	skills, err := h.skillStore.ListSkills(tool, category, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list skills")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"skills": skills,
	})
}

func (h *Handler) handleSkillByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/skills/"):]
	if id == "" {
		writeError(w, http.StatusBadRequest, "skill id required")
		return
	}

	switch r.Method {
	case "GET":
		skill, err := h.skillStore.GetSkill(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get skill")
			return
		}
		if skill == nil {
			writeError(w, http.StatusNotFound, "skill not found")
			return
		}
		writeJSON(w, http.StatusOK, skill)

	case "PUT":
		var skill model.Skill
		if err := json.NewDecoder(r.Body).Decode(&skill); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		skill.ID = id
		if err := h.skillStore.UpsertSkill(&skill); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update skill")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleUsageBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		ClientID string            `json:"client_id"`
		Entries  []model.UsageLog  `json:"entries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.skillStore.BatchInsertUsage(req.Entries); err != nil {
		log.Printf("BatchInsertUsage error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to insert usage logs")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"inserted": len(req.Entries)})
}

func (h *Handler) handleCombinations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tag := r.URL.Query().Get("tag")
		search := r.URL.Query().Get("search")
		combos, err := h.skillStore.ListCombinations(tag, search)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list combinations")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"combinations": combos})

	case "POST":
		var combo model.SkillCombination
		if err := json.NewDecoder(r.Body).Decode(&combo); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := h.skillStore.UpsertCombination(&combo); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create combination")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleChains(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		chains, err := h.skillStore.ListChains()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list chains")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"chains": chains})

	case "POST":
		var chain model.SkillChain
		if err := json.NewDecoder(r.Body).Decode(&chain); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := h.skillStore.UpsertChain(&chain); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create chain")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func extractTool(path string) string {
	switch {
	case len(path) >= 6 && path[:6] == "claude":
		return "claude"
	case len(path) >= 8 && path[:8] == "opencode":
		return "opencode"
	case len(path) >= 4 && path[:4] == "trae":
		return "trae"
	default:
		return "unknown"
	}
}

func extractCategory(path string) string {
	parts := splitPath(path)
	if len(parts) >= 3 {
		return parts[1]
	}
	if len(parts) >= 2 {
		return parts[0]
	}
	return "unknown"
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, c := range path {
		if c == '/' || c == '\\' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func JSONMiddleware(handler *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler.ServeHTTP(w, r)
	})
}

func NewServer(syncStore *store.SyncStore, skillStore *store.SkillStore, tokenAuth *auth.TokenAuth, addr string) *http.Server {
	handler := NewHandler(syncStore, skillStore, tokenAuth)
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func init() {
	_ = fmt.Println
}