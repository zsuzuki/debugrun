package exec

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

func Run(argv []string, env map[string]string, stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = mergeEnv(os.Environ(), env)
	return cmd.Run()
}

func mergeEnv(base []string, overrides map[string]string) []string {
	if len(overrides) == 0 {
		return base
	}

	merged := make(map[string]string, len(base)+len(overrides))
	order := make([]string, 0, len(base)+len(overrides))
	for _, item := range base {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}
		if _, exists := merged[key]; !exists {
			order = append(order, key)
		}
		merged[key] = value
	}
	for key, value := range overrides {
		if _, exists := merged[key]; !exists {
			order = append(order, key)
		}
		merged[key] = value
	}
	out := make([]string, 0, len(order))
	for _, key := range order {
		out = append(out, key+"="+merged[key])
	}
	return out
}
