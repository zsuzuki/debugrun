package config

import "fmt"

type Config struct {
	Version  int                `toml:"version"`
	Global   GlobalConfig       `toml:"global"`
	Profiles map[string]Profile `toml:"profiles"`
}

type GlobalConfig struct {
	HistoryFile     string `toml:"history_file"`
	DefaultExecMode string `toml:"default_exec_mode"`
}

type Profile struct {
	Name        string            `toml:"-"`
	Bin         string            `toml:"bin"`
	Inherits    string            `toml:"inherits"`
	Env         map[string]string `toml:"env"`
	LiteralArgs []string          `toml:"literal_args"`
	Params      []ParamSpec       `toml:"params"`
}

type ParamSpec struct {
	Name             string   `toml:"name"`
	ArgName          string   `toml:"arg_name"`
	ArgMode          string   `toml:"arg_mode"`
	ListMode         string   `toml:"list_mode"`
	Kind             string   `toml:"kind"`
	Required         bool     `toml:"required"`
	Multi            bool     `toml:"multi"`
	Delimiter        string   `toml:"delimiter"`
	Values           []string `toml:"values"`
	StrictValues     bool     `toml:"strict_values"`
	Help             string   `toml:"help"`
	Default          string   `toml:"default"`
	DefaultList      []string `toml:"default_list"`
	DefaultAllValues bool     `toml:"default_all_values"`

	hasArgName          bool
	hasArgMode          bool
	hasListMode         bool
	hasKind             bool
	hasRequired         bool
	hasMulti            bool
	hasDelimiter        bool
	hasValues           bool
	hasStrictValues     bool
	hasHelp             bool
	hasDefault          bool
	hasDefaultList      bool
	hasDefaultAllValues bool
}

func (p *ParamSpec) UnmarshalTOML(data any) error {
	m, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("param spec must be a table")
	}

	var out ParamSpec
	if value, ok := m["name"]; ok {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("name must be a string")
		}
		out.Name = s
	}
	if value, ok := m["arg_name"]; ok {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("arg_name must be a string")
		}
		out.ArgName = s
		out.hasArgName = true
	}
	if value, ok := m["arg_mode"]; ok {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("arg_mode must be a string")
		}
		out.ArgMode = s
		out.hasArgMode = true
	}
	if value, ok := m["list_mode"]; ok {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("list_mode must be a string")
		}
		out.ListMode = s
		out.hasListMode = true
	}
	if value, ok := m["kind"]; ok {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("kind must be a string")
		}
		out.Kind = s
		out.hasKind = true
	}
	if value, ok := m["required"]; ok {
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("required must be a boolean")
		}
		out.Required = b
		out.hasRequired = true
	}
	if value, ok := m["multi"]; ok {
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("multi must be a boolean")
		}
		out.Multi = b
		out.hasMulti = true
	}
	if value, ok := m["delimiter"]; ok {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("delimiter must be a string")
		}
		out.Delimiter = s
		out.hasDelimiter = true
	}
	if value, ok := m["values"]; ok {
		items, err := decodeStringSlice(value)
		if err != nil {
			return fmt.Errorf("values: %w", err)
		}
		out.Values = items
		out.hasValues = true
	}
	if value, ok := m["strict_values"]; ok {
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("strict_values must be a boolean")
		}
		out.StrictValues = b
		out.hasStrictValues = true
	}
	if value, ok := m["help"]; ok {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("help must be a string")
		}
		out.Help = s
		out.hasHelp = true
	}
	if value, ok := m["default"]; ok {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("default must be a string")
		}
		out.Default = s
		out.hasDefault = true
	}
	if value, ok := m["default_list"]; ok {
		items, err := decodeStringSlice(value)
		if err != nil {
			return fmt.Errorf("default_list: %w", err)
		}
		out.DefaultList = items
		out.hasDefaultList = true
	}
	if value, ok := m["default_all_values"]; ok {
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("default_all_values must be a boolean")
		}
		out.DefaultAllValues = b
		out.hasDefaultAllValues = true
	}

	*p = out
	return nil
}

func decodeStringSlice(value any) ([]string, error) {
	raw, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("must be an array")
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("must contain only strings")
		}
		out = append(out, s)
	}
	return out, nil
}

func (child ParamSpec) MergeOver(parent ParamSpec) ParamSpec {
	if !child.hasExplicitOverrides() {
		return child
	}

	merged := parent
	merged.Name = child.Name
	if child.hasArgName {
		merged.ArgName = child.ArgName
	}
	if child.hasArgMode {
		merged.ArgMode = child.ArgMode
	}
	if child.hasListMode {
		merged.ListMode = child.ListMode
	}
	if child.hasKind {
		merged.Kind = child.Kind
	}
	if child.hasRequired {
		merged.Required = child.Required
	}
	if child.hasMulti {
		merged.Multi = child.Multi
	}
	if child.hasDelimiter {
		merged.Delimiter = child.Delimiter
	}
	if child.hasValues {
		merged.Values = append([]string{}, child.Values...)
	}
	if child.hasStrictValues {
		merged.StrictValues = child.StrictValues
	}
	if child.hasHelp {
		merged.Help = child.Help
	}
	if child.hasDefault {
		merged.Default = child.Default
	}
	if child.hasDefaultList {
		merged.DefaultList = append([]string{}, child.DefaultList...)
	}
	if child.hasDefaultAllValues {
		merged.DefaultAllValues = child.DefaultAllValues
	}
	return merged
}

func (p ParamSpec) WithPresenceCopiedFrom(src ParamSpec) ParamSpec {
	p.hasArgName = src.hasArgName
	p.hasArgMode = src.hasArgMode
	p.hasListMode = src.hasListMode
	p.hasKind = src.hasKind
	p.hasRequired = src.hasRequired
	p.hasMulti = src.hasMulti
	p.hasDelimiter = src.hasDelimiter
	p.hasValues = src.hasValues
	p.hasStrictValues = src.hasStrictValues
	p.hasHelp = src.hasHelp
	p.hasDefault = src.hasDefault
	p.hasDefaultList = src.hasDefaultList
	p.hasDefaultAllValues = src.hasDefaultAllValues
	return p
}

func (p ParamSpec) hasExplicitOverrides() bool {
	return p.hasArgName || p.hasArgMode || p.hasListMode || p.hasKind ||
		p.hasRequired || p.hasMulti || p.hasDelimiter || p.hasValues ||
		p.hasStrictValues || p.hasHelp || p.hasDefault || p.hasDefaultList ||
		p.hasDefaultAllValues
}
