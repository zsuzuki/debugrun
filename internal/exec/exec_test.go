//go:build !windows

package exec

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestRunKillsProcessGroupOnInterrupt(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	script := strings.Join([]string{
		"trap '' INT TERM",
		"(trap '' INT TERM; while :; do sleep 1; done) &",
		fmt.Sprintf("echo $! > %q", pidFile),
		"wait",
	}, "\n")

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run([]string{"/bin/sh", "-c", script}, nil, nil, io.Discard, io.Discard)
	}()

	childPID := waitForPIDFile(t, pidFile)

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("send SIGINT to self: %v", err)
	}

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("Run returned nil after interrupt")
		}
	case <-time.After(3 * time.Second):
		_ = syscall.Kill(childPID, syscall.SIGKILL)
		t.Fatal("Run did not return after interrupt")
	}

	eventuallyNoProcess(t, childPID)
}

func waitForPIDFile(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
			if parseErr == nil {
				return pid
			}
			lastErr = parseErr
		} else {
			lastErr = err
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("read child pid: %v", lastErr)
	return 0
}

func eventuallyNoProcess(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		err := syscall.Kill(pid, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("process %d is still running", pid)
}
