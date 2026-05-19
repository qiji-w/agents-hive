package config

import (
	"os"
	"path/filepath"
	"strings"
)

// LoadDotEnv 从 .env 加载 KEY=VALUE 到进程环境变量。
// 已存在的环境变量不会被覆盖（shell export 优先）。
func LoadDotEnv(paths ...string) {
	seen := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			abs = p
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		loadDotEnvFile(abs)
	}
}

func loadDotEnvForConfig(configPath string) {
	LoadDotEnv(dotEnvCandidatePaths(configPath)...)
}

// dotEnvCandidatePaths 返回可能存在的 .env 路径（配置文件目录、cwd、向上最多 3 层）。
func dotEnvCandidatePaths(configPath string) []string {
	var paths []string
	if configPath != "" {
		if abs, err := filepath.Abs(configPath); err == nil {
			configPath = abs
		}
		paths = append(paths, filepath.Join(filepath.Dir(configPath), ".env"))
	}
	cwd, err := os.Getwd()
	if err != nil {
		return paths
	}
	dir, err := filepath.Abs(cwd)
	if err != nil {
		dir = cwd
	}
	for i := 0; i < 4; i++ {
		paths = append(paths, filepath.Join(dir, ".env"))
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return paths
}

// dotEnvLookup 直接从 .env 文件读取某个 key（不读进程环境变量）。
func dotEnvLookup(configPath, key string) string {
	for _, envPath := range dotEnvCandidatePaths(configPath) {
		if v := dotEnvFileValue(envPath, key); v != "" {
			return v
		}
	}
	return ""
}

func dotEnvFileValue(path, key string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		k, val, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(k) != key {
			continue
		}
		val = strings.TrimSpace(val)
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		return val
	}
	return ""
}

func loadDotEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		val = strings.TrimSpace(val)
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, val)
	}
}
