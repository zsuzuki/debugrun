package profile

import (
	"fmt"

	"debugrun/internal/config"
)

func Resolve(cfg *config.Config, name string) (*config.Profile, error) {
	visiting := map[string]bool{}
	resolved, err := resolve(cfg, name, visiting)
	if err != nil {
		return nil, err
	}
	resolved.Name = name
	if err := config.FinalizeProfile(&resolved); err != nil {
		return nil, fmt.Errorf("profile %q: %w", name, err)
	}
	return &resolved, nil
}

func resolve(cfg *config.Config, name string, visiting map[string]bool) (config.Profile, error) {
	profile, ok := cfg.Profiles[name]
	if !ok {
		return config.Profile{}, fmt.Errorf("unknown profile %q", name)
	}
	if !visiting[name] {
		visiting[name] = true
	} else {
		return config.Profile{}, fmt.Errorf("inheritance cycle detected at %q", name)
	}
	defer delete(visiting, name)

	current := cloneProfile(profile)
	if current.Inherits == "" {
		current.Name = name
		return current, nil
	}

	parent, err := resolve(cfg, current.Inherits, visiting)
	if err != nil {
		return config.Profile{}, err
	}

	merged := parent
	merged.Name = name
	if current.Bin != "" {
		merged.Bin = current.Bin
	}
	merged.Env = mergeEnv(parent.Env, current.Env)
	merged.LiteralArgs = append(append([]string{}, parent.LiteralArgs...), current.LiteralArgs...)
	merged.Params = mergeParams(parent.Params, current.Params)
	return merged, nil
}

func cloneProfile(p config.Profile) config.Profile {
	out := config.Profile{
		Name:        p.Name,
		Bin:         p.Bin,
		Inherits:    p.Inherits,
		Env:         copyEnv(p.Env),
		LiteralArgs: append([]string{}, p.LiteralArgs...),
		Params:      append([]config.ParamSpec{}, p.Params...),
	}
	return out
}

func copyEnv(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func mergeEnv(parent, child map[string]string) map[string]string {
	if len(parent) == 0 && len(child) == 0 {
		return nil
	}

	merged := copyEnv(parent)
	if merged == nil {
		merged = make(map[string]string, len(child))
	}
	for key, value := range child {
		merged[key] = value
	}
	return merged
}

func mergeParams(parent, child []config.ParamSpec) []config.ParamSpec {
	merged := append([]config.ParamSpec{}, parent...)
	index := make(map[string]int, len(parent))
	for i, param := range merged {
		index[param.Name] = i
	}

	for _, param := range child {
		if idx, ok := index[param.Name]; ok {
			merged[idx] = param.MergeOver(merged[idx])
			continue
		}
		index[param.Name] = len(merged)
		merged = append(merged, param)
	}

	return merged
}
