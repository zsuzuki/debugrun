package render

import (
	"fmt"
	"sort"
	"strings"

	"debugrun/internal/config"
	"debugrun/internal/history"
	"debugrun/internal/invoke"
)

func Argv(inv *invoke.Invocation, configDir string) []string {
	argv := []string{invoke.ResolvedBin(inv, configDir)}
	for _, bound := range inv.Params {
		argv = append(argv, renderParamArgv(bound)...)
	}
	argv = append(argv, inv.LiteralArgs...)
	argv = append(argv, inv.ExtraArgs...)
	return argv
}

func renderParamArgv(bound invoke.BoundParam) []string {
	switch bound.Spec.Kind {
	case "string":
		if bound.Value.Scalar == "" {
			return nil
		}
		switch bound.Spec.ArgMode {
		case "split":
			return []string{bound.Spec.ArgName, bound.Value.Scalar}
		case "equals":
			return []string{fmt.Sprintf("%s=%s", bound.Spec.ArgName, bound.Value.Scalar)}
		default:
			return []string{fmt.Sprintf("%s=%s", bound.Spec.Name, bound.Value.Scalar)}
		}
	case "list":
		if len(bound.Value.List) == 0 {
			return nil
		}
		switch bound.Spec.ArgMode {
		case "split":
			argv := make([]string, 0, len(bound.Value.List)*2)
			for _, item := range bound.Value.List {
				argv = append(argv, bound.Spec.ArgName, item)
			}
			return argv
		case "equals":
			return []string{fmt.Sprintf("%s=%s", bound.Spec.ArgName, strings.Join(bound.Value.List, bound.Spec.Delimiter))}
		default:
			return []string{fmt.Sprintf("%s=%s", bound.Spec.Name, strings.Join(bound.Value.List, bound.Spec.Delimiter))}
		}
	default:
		return nil
	}
}

func ShellString(argv []string) string {
	parts := make([]string, 0, len(argv))
	for _, arg := range argv {
		if arg == "" {
			parts = append(parts, "''")
			continue
		}
		if isShellSafe(arg) {
			parts = append(parts, arg)
			continue
		}
		escaped := strings.ReplaceAll(arg, `'`, `'\''`)
		parts = append(parts, "'"+escaped+"'")
	}
	return strings.Join(parts, " ")
}

func FormatProfileList(cfg *config.Config) string {
	return strings.Join(ProfileNames(cfg), "\n")
}

func ProfileNames(cfg *config.Config) []string {
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func ParamNames(profile *config.Profile) []string {
	names := make([]string, 0, len(profile.Params))
	for _, spec := range profile.Params {
		names = append(names, spec.Name)
	}
	return names
}

func ParamValues(profile *config.Profile, paramName string) []string {
	for _, spec := range profile.Params {
		if spec.Name == paramName {
			return append([]string{}, spec.Values...)
		}
	}
	return nil
}

func FormatParams(profile *config.Profile) string {
	lines := []string{
		fmt.Sprintf("Profile: %s", profile.Name),
		fmt.Sprintf("Binary : %s", profile.Bin),
		"",
		"Params:",
	}
	if len(profile.Params) == 0 {
		lines = append(lines, "  (none)")
		return strings.Join(lines, "\n")
	}
	for _, spec := range profile.Params {
		line := fmt.Sprintf("  %-10s %-6s", spec.Name, spec.Kind)
		if spec.ArgName != "" && spec.ArgName != spec.Name {
			line += " arg=" + spec.ArgName
		}
		if spec.ArgMode != "" && spec.ArgMode != "kv" {
			line += " mode=" + spec.ArgMode
		}
		if spec.Kind == "string" && spec.Default != "" {
			line += " default=" + spec.Default
		}
		if spec.Kind == "list" && len(spec.DefaultList) > 0 {
			line += " default=" + strings.Join(spec.DefaultList, spec.Delimiter)
		}
		if spec.Kind == "list" && spec.DefaultAllValues {
			line += " default=all-values"
		}
		if len(spec.Values) > 0 {
			line += " values=" + strings.Join(spec.Values, ",")
		}
		if spec.Help != "" {
			line += "  " + spec.Help
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func FormatHistoryEntry(entry history.Entry) string {
	parts := make([]string, 0, len(entry.Params))
	for _, param := range entry.Params {
		if param.Kind == "string" && param.Scalar != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", param.Name, param.Scalar))
		}
		if param.Kind == "list" && len(param.List) > 0 {
			parts = append(parts, fmt.Sprintf("%s=%s", param.Name, strings.Join(param.List, ",")))
		}
	}
	line := fmt.Sprintf("%s  %s", entry.Timestamp.Format("2006-01-02 15:04:05"), entry.ProfileName)
	if len(parts) > 0 {
		line += "  " + strings.Join(parts, " ")
	}
	if len(entry.ExtraArgs) > 0 {
		line += " -- " + strings.Join(entry.ExtraArgs, " ")
	}
	return line
}

func FormatReplayCommand(commandName string, entry history.Entry) string {
	argv := []string{commandName, entry.ProfileName}
	for _, param := range entry.Params {
		switch param.Kind {
		case "string":
			if param.Scalar != "" {
				argv = append(argv, fmt.Sprintf("%s=%s", param.Name, param.Scalar))
			}
		case "list":
			if len(param.List) > 0 {
				argv = append(argv, fmt.Sprintf("%s=%s", param.Name, strings.Join(param.List, ",")))
			}
		}
	}
	if len(entry.ExtraArgs) > 0 {
		argv = append(argv, "--")
		argv = append(argv, entry.ExtraArgs...)
	}
	return ShellString(argv)
}

func isShellSafe(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '/', '.', '_', '-', '=', ',', ':':
			continue
		default:
			return false
		}
	}
	return true
}
