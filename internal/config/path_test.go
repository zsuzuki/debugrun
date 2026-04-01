package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindPathWalksUpward(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "run.toml")
	if err := os.WriteFile(path, []byte("version = 1\n[profiles.test]\nbin = \"/bin/echo\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := FindPath(nested)
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("FindPath() = %q, want %q", got, path)
	}
}
