package main

import (
	"encoding/json"
	"fmt"
	"os"

	"skill-manage/internal/adapter"
	"skill-manage/internal/client"
	"skill-manage/internal/model"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("SKILL_SYNC_SERVER")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}
	authToken := os.Getenv("SKILL_SYNC_TOKEN")
	if authToken == "" {
		authToken = "change-me-in-production"
	}
	clientID := getClientID()

	engine := client.NewSyncEngine(serverURL, authToken, clientID)

	cmd := os.Args[1]
	cmdArgs := os.Args[2:]

	switch cmd {
	case "status":
		handleStatus(engine)
	case "push":
		handlePush(engine, cmdArgs)
	case "pull":
		handlePull(engine, cmdArgs)
	case "diff":
		handleDiff(engine, cmdArgs)
	case "config":
		handleConfig(cmdArgs)
	case "discover":
		handleDiscover(cmdArgs)
	case "ui":
		handleUI(engine)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Skill Sync - Multi-machine skill sharing tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  skill-sync ui                   Open web UI dashboard")
	fmt.Println("  skill-sync status              Show sync status")
	fmt.Println("  skill-sync push [-y]            Push local changes to server")
	fmt.Println("  skill-sync pull                 Pull changes from server")
	fmt.Println("  skill-sync diff                 Show diff between local and server")
	fmt.Println("  skill-sync discover [tool]      List local files for tool")
	fmt.Println("  skill-sync config               Show current config")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  SKILL_SYNC_SERVER   Server URL (default: http://localhost:8080)")
	fmt.Println("  SKILL_SYNC_TOKEN    API auth token")
}

func handleStatus(engine *client.SyncEngine) {
	info, err := engine.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	jsonData, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(jsonData))
	fmt.Printf("\nServer revision: %d\n", info.Revision)
}

func handlePush(engine *client.SyncEngine, args []string) {
	skipConfirm := false
	for _, a := range args {
		if a == "-y" || a == "--yes" {
			skipConfirm = true
		}
	}

	tools, _ := client.LoadTools()
	if len(tools) == 0 {
		fmt.Println("No tools configured or installed")
		return
	}

	var allChanges []model.FileChange

	for _, cfg := range tools {
		if !cfg.IsInstalled() {
			fmt.Printf("Skipping %s (not installed)\n", cfg.DisplayName)
			continue
		}

		files, err := cfg.DiscoverFiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error discovering %s: %v\n", cfg.DisplayName, err)
			continue
		}

		fmt.Printf("%s: %d files found\n", cfg.DisplayName, len(files))

		for _, f := range files {
			serverPath := cfg.LocalToServer(f.Path)
			change := model.FileChange{
				Path:    serverPath,
				Hash:    f.Hash(),
				Action:  model.ChangeTypeCreate,
				Content: f.Content,
			}
			allChanges = append(allChanges, change)
		}
	}

	if len(allChanges) == 0 {
		fmt.Println("No changes to push")
		return
	}

	fmt.Printf("\nPending changes (%d files):\n", len(allChanges))
	for _, c := range allChanges {
		fmt.Printf("  [%s] %s (%d bytes)\n", c.Action, c.Path, len(c.Content))
	}

	if !skipConfirm {
		fmt.Print("\nPush changes? [y/N]: ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Cancelled")
			return
		}
	}

	resp, err := engine.Push(0, allChanges)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Push error: %v\n", err)
		os.Exit(1)
	}

	if resp.HasConflicts() {
		fmt.Println("\n⚠ Conflicts detected:")
		for _, c := range resp.Conflicts {
			fmt.Printf("  - %s\n", c.Message())
		}
	}
	fmt.Printf("\nPushed %d changes. Server revision: %d\n", resp.Applied, resp.NewRevision)
}

func handlePull(engine *client.SyncEngine, args []string) {
	info, err := engine.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server revision: %d\n", info.Revision)

	resp, err := engine.Pull(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Pull error: %v\n", err)
		os.Exit(1)
	}

	if len(resp.Changes) == 0 {
		fmt.Println("Already up to date")
		return
	}

	fmt.Printf("Changes available: %d\n", len(resp.Changes))
	for _, c := range resp.Changes {
		fmt.Printf("  [%s] %s\n", c.Action, c.Path)
	}

	tools, _ := client.LoadTools()
	if len(tools) == 0 {
		fmt.Println("No tools configured to apply changes")
		return
	}

	toolMap := make(map[string]*adapter.ToolConfig)
	for _, t := range tools {
		toolMap[t.Name] = t
	}

	var filePaths []string
	for _, c := range resp.Changes {
		filePaths = append(filePaths, c.Path)
	}

	files, err := engine.DownloadFiles(filePaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Download error: %v\n", err)
		os.Exit(1)
	}

	applied := 0
	for _, f := range files {
		parts := splitFirstN(f.Path, "/", 2)
		if len(parts) != 2 {
			continue
		}
		toolName := parts[0]
		relPath := parts[1]

		cfg, ok := toolMap[toolName]
		if !ok {
			continue
		}

		if err := cfg.WriteFile(relPath, f.Content); err != nil {
			fmt.Fprintf(os.Stderr, "Write error %s: %v\n", f.Path, err)
			continue
		}
		applied++
	}

	fmt.Printf("\nApplied %d files locally.\n", applied)
}

func handleDiff(engine *client.SyncEngine, args []string) {
	tools, _ := client.LoadTools()
	if len(tools) == 0 {
		fmt.Println("No tools configured")
		return
	}

	info, err := engine.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server revision: %d\n\n", info.Revision)

	for _, cfg := range tools {
		if !cfg.IsInstalled() {
			fmt.Printf("%s: not installed\n", cfg.DisplayName)
			continue
		}

		localFiles, err := cfg.DiscoverFiles()
		if err != nil {
			continue
		}

		fmt.Printf("%s: %d local files\n", cfg.DisplayName, len(localFiles))
		for _, f := range localFiles {
			serverPath := cfg.LocalToServer(f.Path)
			size := len(f.Content)
			if size < 1024 {
				fmt.Printf("  %s (%dB)\n", serverPath, size)
			} else {
				fmt.Printf("  %s (%dKB)\n", serverPath, size/1024)
			}
		}
		fmt.Println()
	}
}

func handleDiscover(args []string) {
	toolFilter := ""
	if len(args) > 0 {
		toolFilter = args[0]
	}

	tools, _ := client.LoadTools()
	for _, cfg := range tools {
		if toolFilter != "" && cfg.Name != toolFilter {
			continue
		}
		if !cfg.IsInstalled() {
			fmt.Printf("%s: not installed\n", cfg.DisplayName)
			continue
		}

		files, err := cfg.DiscoverFiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Printf("%s: %d files\n", cfg.DisplayName, len(files))
		for _, f := range files {
			fmt.Printf("  [%s] %s (%dB)\n", f.Category, f.Path, len(f.Content))
		}
	}
}

func handleConfig(args []string) {
	tools, err := client.LoadTools()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server: %s\n", os.Getenv("SKILL_SYNC_SERVER"))
	fmt.Printf("Client ID: %s\n\n", getClientID())

	for _, t := range tools {
		status := " enabled"
		if !t.Enabled {
			status = " disabled"
		}
		installed := ""
		if t.IsInstalled() {
			installed = " [INSTALLED]"
		}
		fmt.Printf("%s (%s):%s%s\n", t.DisplayName, t.Name, status, installed)
	}
}

func getClientID() string {
	name, _ := os.Hostname()
	if name == "" {
		name = "unknown"
	}
	return name
}

func handleUI(engine *client.SyncEngine) {
	addr := ":3000"
	fmt.Printf("Starting Skill Sync UI at http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop")

	ui := client.NewUIServer(engine, addr)
	if err := ui.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "UI error: %v\n", err)
		os.Exit(1)
	}
}

func splitFirstN(s, sep string, n int) []string {
	for i := 0; i < len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return []string{s[:i], s[i+len(sep):]}
		}
	}
	return []string{s}
}