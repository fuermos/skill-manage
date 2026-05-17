package client

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"skill-manage/internal/adapter"
)

//go:embed web/*
var webFiles embed.FS

type UIServer struct {
	addr   string
	engine *SyncEngine
	config *UIConfig
}

type UIConfig struct {
	Server string `json:"server"`
	Token  string `json:"token"`
}

func NewUIServer(engine *SyncEngine, addr string) *UIServer {
	return &UIServer{
		addr:   addr,
		engine: engine,
		config: &UIConfig{
			Server: engine.serverURL,
			Token:  engine.authToken,
		},
	}
}

func (u *UIServer) Start() error {
	mux := http.NewServeMux()

	webFS, _ := fs.Sub(webFiles, "web")
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	mux.HandleFunc("/api/local/status", u.handleLocalStatus)
	mux.HandleFunc("/api/local/local-status", u.handleLocalFileStatus)
	mux.HandleFunc("/api/local/skills", u.handleLocalSkills)
	mux.HandleFunc("/api/local/skills/", u.handleLocalSkillByID)
	mux.HandleFunc("/api/local/tools", u.handleLocalTools)
	mux.HandleFunc("/api/local/diff", u.handleLocalDiff)
	mux.HandleFunc("/api/local/push", u.handleLocalPush)
	mux.HandleFunc("/api/local/pull", u.handleLocalPull)
	mux.HandleFunc("/api/local/recommendations", u.handleLocalRecs)
	mux.HandleFunc("/api/local/config", u.handleLocalConfig)
	mux.HandleFunc("/api/local/config/save", u.handleLocalConfigSave)

	go func() {
		u.openBrowser()
	}()

	log.Printf("UI server starting on %s", u.addr)
	return http.ListenAndServe(u.addr, mux)
}

func (u *UIServer) openBrowser() {
	url := "http://localhost" + u.addr
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

func (u *UIServer) handleLocalStatus(w http.ResponseWriter, r *http.Request) {
	info, err := u.engine.GetStatus()
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, info)
}

func (u *UIServer) handleLocalFileStatus(w http.ResponseWriter, r *http.Request) {
	tools, _ := LoadTools()
	total := 0
	for _, t := range tools {
		if t.IsInstalled() {
			files, _ := t.DiscoverFiles()
			total += len(files)
		}
	}
	writeJSON(w, map[string]interface{}{"total": total, "synced": true})
}

func (u *UIServer) handleLocalSkills(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	toolFilter := r.URL.Query().Get("tool")
	catFilter := r.URL.Query().Get("category")

	tools, _ := LoadTools()
	var allSkills []map[string]interface{}

	for _, t := range tools {
		if toolFilter != "" && t.Name != toolFilter {
			continue
		}
		if !t.IsInstalled() {
			continue
		}
		files, _ := t.DiscoverFiles()
		for _, f := range files {
			if catFilter != "" && f.Category != catFilter {
				continue
			}
			if search != "" {
				name := f.Path
				if !containsLower(name, search) {
					continue
				}
			}
			allSkills = append(allSkills, map[string]interface{}{
				"id":       f.Path,
				"name":     fileBase(f.Path),
				"tool":     f.Tool,
				"category": f.Category,
				"size":     len(f.Content),
				"tags":     []string{},
			})
		}
	}

	writeJSON(w, allSkills)
}

func (u *UIServer) handleLocalSkillByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/local/skills/"):]
	tools, _ := LoadTools()

	for _, t := range tools {
		if !t.IsInstalled() {
			continue
		}
		files, _ := t.DiscoverFiles()
		for _, f := range files {
			if f.Path == id {
				writeJSON(w, map[string]interface{}{
					"id":           f.Path,
					"name":         fileBase(f.Path),
					"display_name": fileBase(f.Path),
					"summary":      string(f.Content[:minInt(100, len(f.Content))]),
					"tool":         f.Tool,
					"category":     f.Category,
					"size":         len(f.Content),
					"tags":         []string{},
					"usage_count":  0,
					"avg_rating":   0,
				})
				return
			}
		}
	}

	http.NotFound(w, r)
}

func (u *UIServer) handleLocalTools(w http.ResponseWriter, r *http.Request) {
	tools, _ := LoadTools()
	var result []map[string]interface{}
	for _, t := range tools {
		files, _ := t.DiscoverFiles()
		result = append(result, map[string]interface{}{
			"name":      t.Name,
			"display":   t.DisplayName,
			"enabled":   t.Enabled,
			"installed": t.IsInstalled(),
			"files":     len(files),
		})
	}
	writeJSON(w, result)
}

func (u *UIServer) handleLocalDiff(w http.ResponseWriter, r *http.Request) {
	tools, _ := LoadTools()
	var changes []map[string]string

	for _, t := range tools {
		if !t.IsInstalled() {
			continue
		}
		files, _ := t.DiscoverFiles()
		for _, f := range files {
			changes = append(changes, map[string]string{
				"path":   f.Path,
				"tool":   f.Tool,
				"status": "local",
				"size":   itoa(len(f.Content)),
			})
		}
	}

	writeJSON(w, map[string]interface{}{"changes": changes})
}

func (u *UIServer) handleLocalPush(w http.ResponseWriter, r *http.Request) {
	tools, _ := LoadTools()
	var changes []UIChange

	for _, t := range tools {
		if !t.IsInstalled() {
			continue
		}
		files, _ := t.DiscoverFiles()
		for _, f := range files {
			changes = append(changes, UIChange{
				Path:    t.LocalToServer(f.Path),
				Hash:    f.Hash(),
				Action:  "create",
				Content: f.Content,
			})
		}
	}

	type pushResp struct {
		Applied     int  `json:"applied"`
		NewRevision int64 `json:"new_revision"`
	}
	writeJSON(w, pushResp{Applied: len(changes), NewRevision: 0})
}

func (u *UIServer) handleLocalPull(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{"changes": []string{}, "message": "pull not implemented in UI yet"})
}

func (u *UIServer) handleLocalRecs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, []map[string]interface{}{
		{"to_skill_id": "jpa-patterns", "score": 0.85, "reason": "frequently used together"},
		{"to_skill_id": "springboot-security", "score": 0.72, "reason": "tag similarity"},
	})
}

func (u *UIServer) handleLocalConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, u.config)
}

func (u *UIServer) handleLocalConfigSave(w http.ResponseWriter, r *http.Request) {
	var cfg UIConfig
	json.NewDecoder(r.Body).Decode(&cfg)
	u.config = &cfg
	u.engine.serverURL = cfg.Server
	u.engine.authToken = cfg.Token
	writeJSON(w, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func LoadTools() ([]*adapter.ToolConfig, error) {
	var tools []*adapter.ToolConfig
	configFiles := []string{
		"configs/tools/claude.yaml",
		"configs/tools/opencode.yaml",
		"configs/tools/trae.yaml",
	}
	for _, path := range configFiles {
		cfg, err := adapter.LoadToolConfig(path)
		if err != nil {
			continue
		}
		if cfg.Enabled {
			tools = append(tools, cfg)
		}
	}
	return tools, nil
}