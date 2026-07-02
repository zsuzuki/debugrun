//go:build windows

package exec

import (
	"io"
	"os"
	"os/exec"
)

func Run(argv []string, env map[string]string, stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	stdout = wrapDemangler(stdout)
	stderr = wrapDemangler(stderr)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = mergeEnv(os.Environ(), env)
	return flushDemanglers(cmd.Run(), stdout, stderr)
}
