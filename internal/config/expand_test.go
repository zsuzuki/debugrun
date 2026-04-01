package config

import "testing"

func TestExpandReplacesConfiguredVariables(t *testing.T) {
	cfg := &Config{
		Version: 1,
		Global: GlobalConfig{
			HistoryFile: "${HOME}/.run/history.jsonl",
		},
		Profiles: map[string]Profile{
			"main": {
				Name:        "main",
				Bin:         "${CONFIG_DIR}/build/app",
				Env:         map[string]string{"CACHE_DIR": "${CWD}/cache", "HOME_DIR": "${HOME}"},
				LiteralArgs: []string{"--work=${CWD}"},
				Params: []ParamSpec{
					{
						Name:        "data_dir",
						Kind:        "string",
						Default:     "${HOME}/data",
						DefaultList: []string{"${CWD}/a", "${CONFIG_DIR}/b"},
						Values:      []string{"${HOME}/one", "${CWD}/two"},
					},
				},
			},
		},
	}

	expanded := Expand(cfg, ExpandContext{
		HomeDir:   "/Users/tester",
		Cwd:       "/tmp/project/work",
		ConfigDir: "/tmp/project",
	})

	if expanded.Global.HistoryFile != "/Users/tester/.run/history.jsonl" {
		t.Fatalf("history_file = %q", expanded.Global.HistoryFile)
	}

	profile := expanded.Profiles["main"]
	if profile.Bin != "/tmp/project/build/app" {
		t.Fatalf("bin = %q", profile.Bin)
	}
	if profile.LiteralArgs[0] != "--work=/tmp/project/work" {
		t.Fatalf("literal_args = %#v", profile.LiteralArgs)
	}
	if profile.Env["CACHE_DIR"] != "/tmp/project/work/cache" || profile.Env["HOME_DIR"] != "/Users/tester" {
		t.Fatalf("env = %#v", profile.Env)
	}
	if profile.Params[0].Default != "/Users/tester/data" {
		t.Fatalf("default = %q", profile.Params[0].Default)
	}
	if profile.Params[0].DefaultList[0] != "/tmp/project/work/a" || profile.Params[0].DefaultList[1] != "/tmp/project/b" {
		t.Fatalf("default_list = %#v", profile.Params[0].DefaultList)
	}
	if profile.Params[0].Values[0] != "/Users/tester/one" || profile.Params[0].Values[1] != "/tmp/project/work/two" {
		t.Fatalf("values = %#v", profile.Params[0].Values)
	}
}

func TestExpandLeavesUnknownVariablesUntouched(t *testing.T) {
	got := expandString("${UNKNOWN}/path", ExpandContext{
		HomeDir:   "/Users/tester",
		Cwd:       "/tmp/project/work",
		ConfigDir: "/tmp/project",
	})
	if got != "${UNKNOWN}/path" {
		t.Fatalf("expanded = %q", got)
	}
}
