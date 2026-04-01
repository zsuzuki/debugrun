package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const envConfigPath = "RUN_CONFIG"

func FindPath(startDir string) (string, error) {
	if override := os.Getenv(envConfigPath); override != "" {
		return override, nil
	}

	dir := startDir
	for {
		candidate := filepath.Join(dir, "run.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("run.toml not found from %q upward; set %s to override", startDir, envConfigPath)
		}
		dir = parent
	}
}
