package invoke

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"debugrun/internal/cli"
	"debugrun/internal/config"
)

type Value struct {
	Scalar string
	List   []string
}

type BoundParam struct {
	Spec  config.ParamSpec
	Value Value
}

type Invocation struct {
	ProfileName string
	Bin         string
	Params      []BoundParam
	LiteralArgs []string
	ExtraArgs   []string
}

type Warning struct {
	Param   string
	Message string
}

func Build(p *config.Profile, parsed *cli.Parsed) (*Invocation, error) {
	index := make(map[string]int, len(p.Params))
	seen := make(map[string]bool, len(p.Params))
	params := make([]BoundParam, 0, len(p.Params))
	for _, spec := range p.Params {
		bound := BoundParam{Spec: spec}
		switch spec.Kind {
		case "string":
			if spec.Default != "" {
				bound.Value = Value{Scalar: spec.Default}
			}
		case "list":
			if len(spec.DefaultList) > 0 {
				bound.Value = Value{List: append([]string{}, spec.DefaultList...)}
			} else if spec.DefaultAllValues && len(spec.Values) > 0 {
				bound.Value = Value{List: append([]string{}, spec.Values...)}
			}
		}
		index[spec.Name] = len(params)
		params = append(params, bound)
	}

	for _, raw := range parsed.RawParams {
		idx, ok := index[raw.Name]
		if !ok {
			return nil, fmt.Errorf("unknown param %q for profile %q", raw.Name, p.Name)
		}
		value, err := parseValue(params[idx].Spec, raw.Value)
		if err != nil {
			return nil, err
		}
		spec := params[idx].Spec
		if raw.Append {
			if spec.Kind != "list" || !spec.Multi {
				return nil, fmt.Errorf("param %q does not support -add", spec.Name)
			}
			params[idx].Value.List = append(params[idx].Value.List, value.List...)
			seen[spec.Name] = true
			continue
		}
		if spec.Kind == "list" && spec.Multi {
			if seen[spec.Name] {
				params[idx].Value.List = append(params[idx].Value.List, value.List...)
			} else {
				params[idx].Value = value
			}
			seen[spec.Name] = true
			continue
		}
		params[idx].Value = value
		seen[spec.Name] = true
	}

	return &Invocation{
		ProfileName: p.Name,
		Bin:         p.Bin,
		Params:      params,
		LiteralArgs: append([]string{}, p.LiteralArgs...),
		ExtraArgs:   append([]string{}, parsed.ExtraArgs...),
	}, nil
}

func Validate(inv *Invocation, p *config.Profile, configDir string, checkBinary bool) ([]Warning, error) {
	bin := inv.Bin
	if checkBinary {
		if !filepath.IsAbs(bin) {
			bin = filepath.Join(configDir, bin)
		}
		info, err := os.Stat(bin)
		if err != nil {
			return nil, fmt.Errorf("binary %q: %w", bin, err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("binary %q is a directory", bin)
		}
	}

	var warnings []Warning
	for _, bound := range inv.Params {
		spec := bound.Spec
		switch spec.Kind {
		case "string":
			if spec.Required && bound.Value.Scalar == "" {
				return nil, fmt.Errorf("missing required param %q", spec.Name)
			}
			if len(spec.Values) > 0 && bound.Value.Scalar != "" && !contains(spec.Values, bound.Value.Scalar) {
				if spec.StrictValues {
					return nil, fmt.Errorf("param %q: value %q is not allowed", spec.Name, bound.Value.Scalar)
				}
				warnings = append(warnings, Warning{
					Param:   spec.Name,
					Message: fmt.Sprintf("value %q is outside configured candidates", bound.Value.Scalar),
				})
			}
		case "list":
			if spec.Required && len(bound.Value.List) == 0 {
				return nil, fmt.Errorf("missing required param %q", spec.Name)
			}
			if !spec.Multi && len(bound.Value.List) > 1 {
				return nil, fmt.Errorf("param %q does not allow multiple values", spec.Name)
			}
			if len(spec.Values) > 0 {
				for _, item := range bound.Value.List {
					if !contains(spec.Values, item) {
						if spec.StrictValues {
							return nil, fmt.Errorf("param %q: value %q is not allowed", spec.Name, item)
						}
						warnings = append(warnings, Warning{
							Param:   spec.Name,
							Message: fmt.Sprintf("value %q is outside configured candidates", item),
						})
					}
				}
			}
		}
	}

	return warnings, nil
}

func ResolvedBin(inv *Invocation, configDir string) string {
	if filepath.IsAbs(inv.Bin) {
		return inv.Bin
	}
	return filepath.Join(configDir, inv.Bin)
}

func parseValue(spec config.ParamSpec, raw string) (Value, error) {
	if raw == "" {
		return Value{}, fmt.Errorf("empty value is not allowed for %q", spec.Name)
	}
	switch spec.Kind {
	case "string":
		return Value{Scalar: raw}, nil
	case "list":
		parts := strings.Split(raw, spec.Delimiter)
		items := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			items = append(items, part)
		}
		if len(items) == 0 {
			return Value{}, fmt.Errorf("empty list is not allowed for %q", spec.Name)
		}
		if !spec.Multi && len(items) > 1 {
			return Value{}, fmt.Errorf("param %q does not allow multiple values", spec.Name)
		}
		return Value{List: items}, nil
	default:
		return Value{}, fmt.Errorf("unsupported param kind %q", spec.Kind)
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
