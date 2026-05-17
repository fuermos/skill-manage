package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"skill-manage/internal/adapter"
	"skill-manage/internal/client"
	"skill-manage/internal/model"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx      context.Context
	engine   *client.SyncEngine
	configPath string
}

func NewApp() *App {
	configPath := getConfigPath()
	cfg := loadConfig(configPath)

	engine := client.NewSyncEngine(cfg.Server, cfg.Token, getHostname())
	return &App{engine: engine, configPath: configPath}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.LogInfo(ctx, "Skill Sync app started successfully")
	runtime.LogInfo(ctx, "Config path: "+a.configPath)
	runtime.LogInfo(ctx, "Server: "+a.engine.ServerURL())
}

func (a *App) shutdown(ctx context.Context) {}

type AppStatus struct {
	Revision int64  `json:"revision"`
	Updated  string `json:"updated"`
}

func (a *App) GetStatus() AppStatus {
	info, err := a.engine.GetStatus()
	if err != nil {
		return AppStatus{}
	}
	return AppStatus{
		Revision: info.Revision,
		Updated:  info.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

type ToolInfo struct {
	Name      string `json:"name"`
	Display   string `json:"display"`
	Enabled   bool   `json:"enabled"`
	Installed bool   `json:"installed"`
	Files     int    `json:"files"`
}

func (a *App) GetTools() []ToolInfo {
	tools, _ := loadTools()
	var result []ToolInfo
	for _, t := range tools {
		files, _ := t.DiscoverFiles()
		result = append(result, ToolInfo{
			Name:      t.Name,
			Display:   t.DisplayName,
			Enabled:   t.Enabled,
			Installed: t.IsInstalled(),
			Files:     len(files),
		})
	}
	return result
}

type SkillInfo struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Tool        string   `json:"tool"`
	Category    string   `json:"category"`
	Size        int      `json:"size"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func (a *App) GetSkills(tool, category, search string) []SkillInfo {
	tools, _ := loadTools()
	var result []SkillInfo
	for _, t := range tools {
		if tool != "" && t.Name != tool {
			continue
		}
		if !t.IsInstalled() {
			continue
		}
		files, _ := t.DiscoverFiles()
		for _, f := range files {
			if category != "" && f.Category != category {
				continue
			}
			baseName := fileBaseOnly(f.Path)
			displayName := f.MetaName
			if displayName == "" {
				displayName = baseName
			}
			summary := f.MetaSummary
			if summary == "" {
				summary = f.MetaDescription
			}
			if summary == "" && len(f.Content) > 0 {
				end := len(f.Content)
				if end > 100 {
					end = 100
				}
				summary = string(f.Content[end:end])
			}
			if search != "" && !clientContains(displayName, search) && !clientContains(f.MetaDescription, search) {
				continue
			}
			result = append(result, SkillInfo{
				Name:        filepath.Base(f.Path),
				DisplayName: displayName,
				Tool:        f.Tool,
				Category:    f.Category,
				Size:        len(f.Content),
				Summary:     summary,
				Description: f.MetaDescription,
				Tags:        f.MetaTags,
			})
		}
	}
	return result
}

type PushResult struct {
	Applied     int   `json:"applied"`
	NewRevision int64 `json:"new_revision"`
}

func (a *App) Push() PushResult {
	tools, _ := loadTools()
	var changes []model.FileChange
	for _, t := range tools {
		if !t.IsInstalled() {
			continue
		}
		files, _ := t.DiscoverFiles()
		for _, f := range files {
			changes = append(changes, model.FileChange{
				Path:    t.LocalToServer(f.Path),
				Hash:    f.Hash(),
				Action:  model.ChangeTypeCreate,
				Content: f.Content,
			})
		}
	}
	resp, err := a.engine.Push(0, changes)
	if err != nil {
		log.Printf("Push error: %v", err)
		return PushResult{}
	}
	return PushResult{Applied: resp.Applied, NewRevision: resp.NewRevision}
}

type PullResult struct {
	Changes []PullChange `json:"changes"`
}

type PullChange struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}

func (a *App) Pull() PullResult {
	resp, err := a.engine.Pull(nil)
	if err != nil {
		return PullResult{}
	}
	var changes []PullChange
	for _, c := range resp.Changes {
		changes = append(changes, PullChange{Path: c.Path, Action: c.Action.String()})
	}
	return PullResult{Changes: changes}
}

type DiffItem struct {
	Path   string `json:"path"`
	Tool   string `json:"tool"`
	Status string `json:"status"`
}

func (a *App) GetDiff() []DiffItem {
	tools, _ := loadTools()
	var items []DiffItem
	for _, t := range tools {
		if !t.IsInstalled() {
			continue
		}
		files, _ := t.DiscoverFiles()
		for _, f := range files {
			items = append(items, DiffItem{
				Path:   f.Path,
				Tool:   f.Tool,
				Status: "local",
			})
		}
	}
	return items
}

type RecItem struct {
	ToSkillID string  `json:"to_skill_id"`
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`
}

func (a *App) GetRecommendations() []RecItem {
	return []RecItem{
		{ToSkillID: "jpa-patterns", Score: 0.85, Reason: "frequently used together with java-coding"},
		{ToSkillID: "springboot-security", Score: 0.72, Reason: "tag similarity (java, spring)"},
	}
}

type AppConfig struct {
	Server string `json:"server"`
	Token  string `json:"token"`
}

func (a *App) GetConfig() AppConfig {
	cfg := loadConfig(a.configPath)
	return cfg
}

func (a *App) SaveConfig(server, token string) {
	cfg := AppConfig{Server: server, Token: token}
	saveConfig(a.configPath, cfg)
	a.engine = client.NewSyncEngine(server, token, getHostname())
}

func (a *App) ShowDialog(title, msg string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: msg,
	})
}

func loadTools() ([]*adapter.ToolConfig, error) {
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

func fileBaseOnly(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[i+1:]
		}
	}
	return p
}

func clientContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			a := s[i+j]
			b := substr[j]
			if a >= 'A' && a <= 'Z' {
				a += 32
			}
			if b >= 'A' && b <= 'Z' {
				b += 32
			}
			if a != b {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func getHostname() string {
	return "desktop-client"
}

func getConfigPath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, _ := os.UserHomeDir()
		appData = filepath.Join(home, ".config")
	}
	dir := filepath.Join(appData, "skill-sync")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "config.json")
}

func loadConfig(path string) AppConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{
			Server: "http://localhost:8080",
			Token:  "skill-sync-2026",
		}
	}
	var cfg AppConfig
	json.Unmarshal(data, &cfg)
	if cfg.Server == "" {
		cfg.Server = "http://localhost:8080"
	}
	if cfg.Token == "" {
		cfg.Token = "skill-sync-2026"
	}
	return cfg
}

func saveConfig(path string, cfg AppConfig) {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(path, data, 0644)
}

func init() {
	_ = log.Println
}