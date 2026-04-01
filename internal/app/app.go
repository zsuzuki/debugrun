package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"debugrun/internal/cli"
	"debugrun/internal/config"
	runexec "debugrun/internal/exec"
	"debugrun/internal/history"
	"debugrun/internal/invoke"
	"debugrun/internal/profile"
	"debugrun/internal/render"
)

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	parsed, err := cli.Parse(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	startDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	configPath, err := config.FindPath(startDir)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	configDir := filepath.Dir(configPath)

	switch parsed.Action {
	case cli.ActionLast:
		return runLast(configPath, stdout, stderr)
	case cli.ActionEditLast:
		return runEditLast(parsed.RepeatIndex, configPath, stdout, stderr)
	case cli.ActionHistory:
		return runHistory(configPath, stdout, stderr)
	case cli.ActionRepeat:
		return runRepeat(parsed.RepeatIndex, configPath, configDir, stdin, stdout, stderr)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	cfg = config.Expand(cfg, config.ExpandContext{
		HomeDir:   resolveHomeDir(),
		Cwd:       startDir,
		ConfigDir: configDir,
	})

	switch parsed.Action {
	case cli.ActionList:
		fmt.Fprintln(stdout, render.FormatProfileList(cfg))
		return 0
	case cli.ActionComplete:
		return runComplete(parsed, cfg, stdout, stderr)
	case cli.ActionParams:
		prof, err := profile.Resolve(cfg, parsed.ProfileName)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, render.FormatParams(prof))
		return 0
	case cli.ActionShow, cli.ActionExec:
		return runInvocation(parsed, cfg, configDir, stdin, stdout, stderr)
	default:
		fmt.Fprintln(stderr, "unsupported action")
		return 2
	}
}

func runInvocation(parsed *cli.Parsed, cfg *config.Config, configDir string, stdin io.Reader, stdout, stderr io.Writer) int {
	prof, err := profile.Resolve(cfg, parsed.ProfileName)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	inv, err := invoke.Build(prof, parsed)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	checkBinary := parsed.Action == cli.ActionExec
	warnings, err := invoke.Validate(inv, prof, configDir, checkBinary)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	printWarnings(stderr, warnings)

	argv := render.Argv(inv, configDir)
	if parsed.Action == cli.ActionShow {
		fmt.Fprintln(stdout, render.CommandString(inv, configDir))
		return 0
	}

	historyPath := resolveHistoryPath(configDir, cfg.Global.HistoryFile)
	if err := history.Append(historyPath, history.FromInvocation(inv)); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if err := runexec.Run(argv, inv.Env, stdin, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func runLast(configPath string, stdout, stderr io.Writer) int {
	cfg, err := loadRuntimeConfig(configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	entry, err := history.Last(resolveHistoryPath(filepath.Dir(configPath), cfg.Global.HistoryFile))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintln(stdout, render.FormatHistoryEntry(*entry))
	return 0
}

func runHistory(configPath string, stdout, stderr io.Writer) int {
	cfg, err := loadRuntimeConfig(configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	entries, err := history.ReadAll(resolveHistoryPath(filepath.Dir(configPath), cfg.Global.HistoryFile))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	for _, entry := range entries {
		fmt.Fprintln(stdout, render.FormatHistoryEntry(entry))
	}
	return 0
}

func runEditLast(index int, configPath string, stdout, stderr io.Writer) int {
	cfg, err := loadRuntimeConfig(configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	entry, err := history.FromLast(resolveHistoryPath(filepath.Dir(configPath), cfg.Global.HistoryFile), index)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintln(stdout, render.FormatReplayCommand("run", *entry))
	return 0
}

func runRepeat(index int, configPath, configDir string, stdin io.Reader, stdout, stderr io.Writer) int {
	cfg, err := loadRuntimeConfig(configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	entry, err := history.FromLast(resolveHistoryPath(configDir, cfg.Global.HistoryFile), index)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	prof, err := profile.Resolve(cfg, entry.ProfileName)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	parsed := &cli.Parsed{
		Action:      cli.ActionExec,
		ProfileName: entry.ProfileName,
		ExtraArgs:   append([]string{}, entry.ExtraArgs...),
	}
	for _, param := range entry.Params {
		switch param.Kind {
		case "string":
			if param.Scalar != "" {
				parsed.RawParams = append(parsed.RawParams, cli.RawParam{Name: param.Name, Value: param.Scalar})
			}
		case "list":
			if len(param.List) > 0 {
				parsed.RawParams = append(parsed.RawParams, cli.RawParam{Name: param.Name, Value: joinList(param.List)})
			}
		}
	}

	inv, err := invoke.Build(prof, parsed)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	warnings, err := invoke.Validate(inv, prof, configDir, true)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	printWarnings(stderr, warnings)

	argv := render.Argv(inv, configDir)
	if err := history.Append(resolveHistoryPath(configDir, cfg.Global.HistoryFile), history.FromInvocation(inv)); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := runexec.Run(argv, inv.Env, stdin, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func resolveHistoryPath(configDir, historyFile string) string {
	if filepath.IsAbs(historyFile) {
		return historyFile
	}
	return filepath.Join(configDir, historyFile)
}

func joinList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	out := items[0]
	for _, item := range items[1:] {
		out += "," + item
	}
	return out
}

func loadRuntimeConfig(configPath string) (*config.Config, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return config.Expand(cfg, config.ExpandContext{
		HomeDir:   resolveHomeDir(),
		Cwd:       cwd,
		ConfigDir: filepath.Dir(configPath),
	}), nil
}

func resolveHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func printWarnings(w io.Writer, warnings []invoke.Warning) {
	if len(warnings) == 0 {
		return
	}
	for _, warning := range warnings {
		fmt.Fprintf(w, "warning: %s: %s\n", warning.Param, warning.Message)
	}
}

func runComplete(parsed *cli.Parsed, cfg *config.Config, stdout, stderr io.Writer) int {
	switch parsed.Topic {
	case "profiles":
		for _, name := range render.ProfileNames(cfg) {
			fmt.Fprintln(stdout, name)
		}
		return 0
	case "params":
		prof, err := profile.Resolve(cfg, parsed.ProfileName)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		for _, name := range render.ParamNames(prof) {
			fmt.Fprintln(stdout, name)
		}
		return 0
	case "values":
		prof, err := profile.Resolve(cfg, parsed.ProfileName)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		values := render.ParamValues(prof, parsed.ParamName)
		if values == nil {
			fmt.Fprintf(stderr, "unknown param %q for profile %q\n", parsed.ParamName, parsed.ProfileName)
			return 1
		}
		for _, value := range values {
			fmt.Fprintln(stdout, value)
		}
		return 0
	default:
		fmt.Fprintf(stderr, "unsupported complete topic %q\n", parsed.Topic)
		return 2
	}
}
