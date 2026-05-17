package adapter

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"skill-manage/internal/model"

	"gopkg.in/yaml.v3"
)

type ToolConfig struct {
	Name        string                  `yaml:"name"`
	DisplayName string                  `yaml:"display_name"`
	Enabled     bool                    `yaml:"enabled"`
	LocalPath   map[string]string       `yaml:"local_path"`
	ServerPath  string                  `yaml:"server_path"`
	Categories  map[string]CategoryConfig `yaml:"categories"`
	Ignore      []string                `yaml:"ignore"`
	OnMissing   string                  `yaml:"on_missing"`

	localRoot string
}

type CategoryConfig struct {
	LocalDir string `yaml:"local_dir"`
	Pattern  string `yaml:"pattern"`
	Priority string `yaml:"priority"`
}

func LoadToolConfig(path string) (*ToolConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &ToolConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *ToolConfig) SetLocalRoot(dir string) {
	c.localRoot = dir
}

func (c *ToolConfig) ResolveLocalPath() string {
	if c.localRoot != "" {
		return c.localRoot
	}

	path, ok := c.LocalPath[runtime.GOOS]
	if !ok {
		return ""
	}

	// Expand Windows-style %VAR% environment variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := "%" + parts[0] + "%"
			path = strings.ReplaceAll(path, key, parts[1])
		}
	}
	// Expand Unix-style $VAR/$HOME environment variables
	path = os.ExpandEnv(path)

	return path
}

func (c *ToolConfig) IsInstalled() bool {
	path := c.ResolveLocalPath()
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func (c *ToolConfig) DiscoverFiles() ([]model.FileInfo, error) {
	root := c.ResolveLocalPath()
	if root == "" {
		return nil, nil
	}

	var files []model.FileInfo

	for _, cat := range c.Categories {
		catDir := filepath.Join(root, cat.LocalDir)
		if _, err := os.Stat(catDir); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(catDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}

			relPath, _ := filepath.Rel(root, path)
			relPath = filepath.ToSlash(relPath)

			if c.shouldIgnore(relPath) {
				return nil
			}

			filename := info.Name()
			if !matchPattern(cat.Pattern, filename) {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			fi := model.FileInfo{
				Path:     relPath,
				Tool:     c.Name,
				Category: c.findCategory(relPath),
				Content:  content,
				Size:     info.Size(),
			}
			files = append(files, fi)
			return nil
		})
	}

return files, nil
}

func (c *ToolConfig) shouldIgnore(relPath string) bool {
	normalized := filepath.ToSlash(relPath)
	for _, pattern := range c.Ignore {
		normalizedPattern := filepath.ToSlash(pattern)
		if globMatch(normalizedPattern, normalized) {
			return true
		}
	}
	return false
}

func (c *ToolConfig) findCategory(relPath string) string {
	for name, cat := range c.Categories {
		prefix := filepath.ToSlash(cat.LocalDir + "/")
		path := filepath.ToSlash(relPath)
		if strings.HasPrefix(path, prefix) || strings.HasPrefix(relPath, cat.LocalDir+string(filepath.Separator)) {
			return name
		}
	}
	return "unknown"
}

func matchPattern(pattern, filename string) bool {
	return globMatch(filepath.ToSlash(pattern), filepath.ToSlash(filename))
}

func globMatch(pattern, name string) bool {
	px := 0
	nx := 0
	nextPx := 0
	nextNx := 0
	for px < len(pattern) || nx < len(name) {
		if px < len(pattern) {
			c := pattern[px]
			switch c {
			case '*':
				if px+1 < len(pattern) && pattern[px+1] == '*' {
					px += 2
					rest := pattern[px:]
					if rest == "" {
						return true
					}
					if len(rest) > 0 && rest[0] == '/' {
						rest = rest[1:]
					}
					for i := nx; i <= len(name); i++ {
						if globMatch(rest, name[i:]) {
							return true
						}
						if i < len(name) && name[i] == '/' {
							continue
						}
						if i == len(name) {
							return false
						}
					}
					return false
				}
				nextPx = px
				nextNx = nx + 1
				px++
				continue
			default:
				if nx < len(name) && name[nx] == c {
					px++
					nx++
					continue
				}
			}
		}
		if 0 < nextNx && nextNx <= len(name) {
			px = nextPx
			nx = nextNx
			continue
		}
		return false
	}
	return true
}

func (c *ToolConfig) LocalToServer(localPath string) string {
	return filepath.ToSlash(filepath.Join(c.ServerPath, localPath))
}

func (c *ToolConfig) ServerToLocal(serverPath string) string {
	prefix := c.ServerPath + "/"
	if strings.HasPrefix(serverPath, prefix) {
		return strings.TrimPrefix(serverPath, prefix)
	}
	return serverPath
}

func (c *ToolConfig) WriteFile(relPath string, content []byte) error {
	root := c.ResolveLocalPath()
	fullPath := filepath.Join(root, relPath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, content, 0644)
}