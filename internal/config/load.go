package config

import (
	"fmt"
	"os"
)

func Load(path string) (*Config, error) {
	var cfg Config
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("config file %q: %w", path, err)
	}
	if _, err := tomlDecodeFile(path, &cfg); err != nil {
		return nil, err
	}

	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if cfg.Version != 1 {
		return nil, fmt.Errorf("unsupported config version %d", cfg.Version)
	}
	if cfg.Global.HistoryFile == "" {
		cfg.Global.HistoryFile = ".run/history.jsonl"
	}
	if cfg.Global.DefaultExecMode == "" {
		cfg.Global.DefaultExecMode = "exec"
	}
	if len(cfg.Profiles) == 0 {
		return nil, fmt.Errorf("no profiles configured")
	}

	for name, profile := range cfg.Profiles {
		profile.Name = name
		if err := validateProfile(profile); err != nil {
			return nil, fmt.Errorf("profile %q: %w", name, err)
		}
		cfg.Profiles[name] = profile
	}

	return &cfg, nil
}

func validateProfile(profile Profile) error {
	if profile.Bin == "" && profile.Inherits == "" {
		return fmt.Errorf("bin or inherits is required")
	}

	seen := map[string]struct{}{}
	for i := range profile.Params {
		param := &profile.Params[i]
		if param.Name == "" {
			return fmt.Errorf("param name is required")
		}
		if param.ArgName == "" {
			param.ArgName = param.Name
		}
		switch param.ArgMode {
		case "", "kv", "equals", "split":
			if param.ArgMode == "" {
				if param.ArgName != param.Name {
					param.ArgMode = "equals"
				} else {
					param.ArgMode = "kv"
				}
			}
		default:
			return fmt.Errorf("param %q: unsupported arg_mode %q", param.Name, param.ArgMode)
		}
		switch param.ListMode {
		case "", "join", "repeat":
		default:
			return fmt.Errorf("param %q: unsupported list_mode %q", param.Name, param.ListMode)
		}
		if _, ok := seen[param.Name]; ok {
			return fmt.Errorf("duplicate param %q", param.Name)
		}
		seen[param.Name] = struct{}{}

		switch param.Kind {
		case "string":
			if len(param.DefaultList) > 0 {
				return fmt.Errorf("param %q: default_list is only valid for kind=list", param.Name)
			}
			if param.DefaultAllValues {
				return fmt.Errorf("param %q: default_all_values is only valid for kind=list", param.Name)
			}
			if param.ListMode != "" {
				return fmt.Errorf("param %q: list_mode is only valid for kind=list", param.Name)
			}
			if param.Multi {
				return fmt.Errorf("param %q: multi is only valid for kind=list", param.Name)
			}
		case "list":
			if param.Default != "" {
				return fmt.Errorf("param %q: default is only valid for kind=string", param.Name)
			}
			if param.DefaultAllValues && len(param.DefaultList) > 0 {
				return fmt.Errorf("param %q: default_all_values and default_list cannot both be set", param.Name)
			}
			if param.DefaultAllValues && !param.Multi {
				return fmt.Errorf("param %q: default_all_values requires multi=true", param.Name)
			}
			if param.DefaultAllValues && len(param.Values) == 0 {
				return fmt.Errorf("param %q: default_all_values requires values", param.Name)
			}
			if param.Delimiter == "" {
				param.Delimiter = ","
			}
			if !param.Multi {
				param.Multi = false
			}
			if param.ListMode == "" {
				if param.Multi && param.ArgMode != "kv" {
					param.ListMode = "repeat"
				} else {
					param.ListMode = "join"
				}
			}
			if param.ListMode == "repeat" && !param.Multi {
				return fmt.Errorf("param %q: list_mode=repeat requires multi=true", param.Name)
			}
		default:
			return fmt.Errorf("param %q: unsupported kind %q", param.Name, param.Kind)
		}

		profile.Params[i] = *param
	}

	return nil
}
