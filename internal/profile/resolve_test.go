package profile

import (
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
