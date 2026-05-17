package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadToolConfig(t *testing.T) {
	cfg, err := LoadToolConfig("../../configs/tools/claude.yaml")
	if err != nil {
		t.Fatalf("LoadToolConfig: %v", err)
	}
	if cfg.Name != "claude" {
		t.Errorf("Name = %q, want claude", cfg.Name)
	}
}

func TestResolveWindowsEnvVar(t *testing.T) {
	os.Setenv("TEST_VAR", "C:\\myhome")
	defer os.Unsetenv("TEST_VAR")

	cfg := &ToolConfig{
		Name: "test",
		LocalPath: map[string]string{
			"windows": "%TEST_VAR%\\.config\\test",
		},
	}

	path := cfg.ResolveLocalPath()
	expected := "C:\\myhome\\.config\\test"
	if !contains(path, "myhome") || !contains(path, "test") {
		t.Errorf("ResolveLocalPath() = %q, should contain myhome and test", path)
	}
	_ = expected
}

func TestResolveClaudePath(t *testing.T) {
	cfg, err := LoadToolConfig("../../configs/tools/claude.yaml")
	if err != nil { t.Fatal(err) }

	path := cfg.ResolveLocalPath()
	t.Logf("Claude resolved path: %s", path)
	if path == "" {
		t.Fatal("path should not be empty")
	}

	_, err = os.Stat(path)
	if err != nil {
		t.Errorf("path does not exist: %s (%v)", path, err)
	}
}

func TestDiscoverRealFiles(t *testing.T) {
	cfg, err := LoadToolConfig("../../configs/tools/claude.yaml")
	if err != nil { t.Fatal(err) }

	if !cfg.IsInstalled() {
		t.Skip("Claude Code not installed")
	}

	files, err := cfg.DiscoverFiles()
	if err != nil {
		t.Fatalf("DiscoverFiles: %v", err)
	}
	t.Logf("Found %d files", len(files))
	for _, f := range files {
		t.Logf("  [%s] %s", f.Category, f.Path)
	}
}

func TestDiscoverFiles(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "skills", "java-coding")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Java Coding"), 0644)

	cfg := &ToolConfig{
		Name:       "claude",
		ServerPath: "claude",
		Categories: map[string]CategoryConfig{
			"skills": {LocalDir: "skills", Pattern: "SKILL.md", Priority: "high"},
			"rules":  {LocalDir: "rules", Pattern: "*.md", Priority: "high"},
		},
		Ignore: []string{".git", "learned"},
	}
	cfg.SetLocalRoot(dir)

	files, err := cfg.DiscoverFiles()
	if err != nil {
		t.Fatalf("DiscoverFiles: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("should have discovered files")
	}
}

func TestOpenCodeConfig(t *testing.T) {
	cfg, err := LoadToolConfig("../../configs/tools/opencode.yaml")
	if err != nil {
		t.Fatalf("LoadToolConfig: %v", err)
	}
	if cfg.Name != "opencode" {
		t.Errorf("Name = %q, want opencode", cfg.Name)
	}
}

func TestGlobMatch(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"**/SKILL.md", "SKILL.md", true},
		{"**/SKILL.md", "skills/java-coding/SKILL.md", true},
		{"**/*.md", "SKILL.md", true},
		{"**/*.md", "rules/common/coding-style.md", true},
		{"SKILL.md", "SKILL.md", true},
		{"SKILL.md", "OTHER.md", false},
		{"**/learned/**", "skills/learned/SKILL.md", true},
		{"**/learned/**", "skills/java-coding/SKILL.md", false},
		{"**/.git/**", "skills/.git/SKILL.md", true},
		{"**/.git/**", "skills/java-coding/SKILL.md", false},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+"/"+tt.name, func(t *testing.T) {
			got := matchPattern(tt.pattern, tt.name)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
			}
			got = globMatch(tt.pattern, tt.name)
			if got != tt.want {
				t.Errorf("globMatch(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
			}
		})
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