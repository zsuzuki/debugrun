package profile

import (
	"os"
	"path/filepath"
	"testing"

	"debugrun/internal/config"
)

func TestResolveMergesLiteralArgsAndParams(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"base": {
				Bin:         "/bin/base",
				Env:         map[string]string{"APP_MODE": "debug", "DATA_DIR": "/tmp/base"},
				LiteralArgs: []string{"--base"},
				Params: []config.ParamSpec{
					{Name: "data_dir", Kind: "string", Default: "/tmp/base"},
					{Name: "entities", Kind: "list", Multi: true, Delimiter: ","},
				},
			},
			"child": {
				Inherits:    "base",
				Bin:         "/bin/child",
				Env:         map[string]string{"APP_MODE": "release"},
				LiteralArgs: []string{"--child"},
				Params: []config.ParamSpec{
					{Name: "entities", Kind: "list", Multi: true, Delimiter: ",", Values: []string{"item-1"}},
					{Name: "fields", Kind: "list", Multi: true, Delimiter: ","},
				},
			},
		},
	}

	got, err := Resolve(cfg, "child")
	if err != nil {
		t.Fatal(err)
	}
	if got.Bin != "/bin/child" {
		t.Fatalf("bin = %q", got.Bin)
	}
	if len(got.LiteralArgs) != 2 || got.LiteralArgs[0] != "--base" || got.LiteralArgs[1] != "--child" {
		t.Fatalf("literal args = %#v", got.LiteralArgs)
	}
	if got.Env["APP_MODE"] != "release" || got.Env["DATA_DIR"] != "/tmp/base" {
		t.Fatalf("env = %#v", got.Env)
	}
	if len(got.Params) != 3 {
		t.Fatalf("params len = %d", len(got.Params))
	}
	if got.Params[1].Name != "entities" || len(got.Params[1].Values) != 1 || got.Params[1].Values[0] != "item-1" {
		t.Fatalf("entities override failed: %#v", got.Params[1])
	}
	if got.Params[2].Name != "fields" {
		t.Fatalf("last param = %q", got.Params[2].Name)
	}
}

func TestResolveInheritsUnspecifiedParamFields(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "run.toml")
	content := `
version = 1

[profiles.base]
bin = "/bin/base"

[[profiles.base.params]]
name = "dir"
kind = "list"
multi = true
delimiter = ","
arg_name = "-dir"
arg_mode = "equals"
list_mode = "repeat"
default_list = ["BASE"]
values = ["A", "B"]

[profiles.child]
inherits = "base"
bin = "/bin/child"

[[profiles.child.params]]
name = "dir"
default_list = ["CHILD"]
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(cfg, "child")
	if err != nil {
		t.Fatal(err)
	}
	if got.Params[0].Kind != "list" || !got.Params[0].Multi {
		t.Fatalf("param kind/multi = %#v", got.Params[0])
	}
	if got.Params[0].ArgMode != "equals" || got.Params[0].ListMode != "repeat" {
		t.Fatalf("param modes = %#v", got.Params[0])
	}
	if len(got.Params[0].DefaultList) != 1 || got.Params[0].DefaultList[0] != "CHILD" {
		t.Fatalf("default_list = %#v", got.Params[0].DefaultList)
	}
	if len(got.Params[0].Values) != 2 || got.Params[0].Values[0] != "A" || got.Params[0].Values[1] != "B" {
		t.Fatalf("values = %#v", got.Params[0].Values)
	}
}
