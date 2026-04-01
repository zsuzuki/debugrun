package config

import "strings"

type ExpandContext struct {
	HomeDir   string
	Cwd       string
	ConfigDir string
}

func Expand(cfg *Config, ctx ExpandContext) *Config {
	if cfg == nil {
		return nil
	}

	out := &Config{
		Version: cfg.Version,
		Global: GlobalConfig{
			HistoryFile:     expandString(cfg.Global.HistoryFile, ctx),
			DefaultExecMode: cfg.Global.DefaultExecMode,
		},
		Profiles: make(map[string]Profile, len(cfg.Profiles)),
	}

	for name, profile := range cfg.Profiles {
		expanded := Profile{
			Name:        profile.Name,
			Bin:         expandString(profile.Bin, ctx),
			Inherits:    profile.Inherits,
			LiteralArgs: expandSlice(profile.LiteralArgs, ctx),
			Params:      make([]ParamSpec, 0, len(profile.Params)),
		}

		for _, param := range profile.Params {
			expanded.Params = append(expanded.Params, ParamSpec{
				Name:             param.Name,
				ArgName:          expandString(param.ArgName, ctx),
				ArgMode:          param.ArgMode,
				Kind:             param.Kind,
				Required:         param.Required,
				Multi:            param.Multi,
				Delimiter:        param.Delimiter,
				Values:           expandSlice(param.Values, ctx),
				StrictValues:     param.StrictValues,
				Help:             param.Help,
				Default:          expandString(param.Default, ctx),
				DefaultList:      expandSlice(param.DefaultList, ctx),
				DefaultAllValues: param.DefaultAllValues,
			})
		}

		out.Profiles[name] = expanded
	}

	return out
}

func expandSlice(items []string, ctx ExpandContext) []string {
	if len(items) == 0 {
		return nil
	}

	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, expandString(item, ctx))
	}
	return out
}

func expandString(s string, ctx ExpandContext) string {
	if s == "" {
		return ""
	}

	vars := map[string]string{
		"HOME":       ctx.HomeDir,
		"CWD":        ctx.Cwd,
		"CONFIG_DIR": ctx.ConfigDir,
	}

	return strings.NewReplacer(
		"${HOME}", vars["HOME"],
		"${CWD}", vars["CWD"],
		"${CONFIG_DIR}", vars["CONFIG_DIR"],
	).Replace(s)
}
