//go:build !windows

package exec

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const shutdownGrace = 500 * time.Millisecond

func Run(argv []string, env map[string]string, stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = mergeEnv(os.Environ(), env)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return err
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case sig := <-signals:
		signalProcessGroup(cmd.Process.Pid, sig)
	}

	select {
	case err := <-done:
		return err
	case <-time.After(shutdownGrace):
		killProcessGroup(cmd.Process.Pid)
		return <-done
	}
}

func signalProcessGroup(pid int, sig os.Signal) {
	sysSig, ok := sig.(syscall.Signal)
	if !ok {
		sysSig = syscall.SIGTERM
	}
	_ = syscall.Kill(-pid, sysSig)
}

func killProcessGroup(pid int) {
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}
