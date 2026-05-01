package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRejectsDuplicateParamAlias(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "run.toml")
	content := `
version = 1

[profiles.app]
bin = "/bin/echo"

[[profiles.app.params]]
name = "value"
alias = "v"
kind = "string"

[[profiles.app.params]]
name = "other"
alias = "v"
kind = "string"
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil || err.Error() != `profile "app": duplicate param alias "v"` {
		t.Fatalf("err = %v", err)
	}
}
