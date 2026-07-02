//go:build !windows

package exec

import (
	"bytes"
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

func TestDemangleLineReplacesMangledSymbols(t *testing.T) {
	cxxfilt := fakeCxxFilt(t)

	line := "2   app  0x1 _ZN8GTEngine4TickEv + 12\n4   app  0x2 _ZZN3Foo3barEvEN3$_08__invokeEv + 24\n"
	got := demangleLine(cxxfilt, line)
	want := "2   app  0x1 GTEngine::Tick() + 12\n4   app  0x2 Foo::bar()::$_0::__invoke() + 24\n"
	if got != want {
		t.Fatalf("demangleLine() = %q, want %q", got, want)
	}
}

func TestDemangleWriterFlushesPartialLine(t *testing.T) {
	var out bytes.Buffer
	writer := &demangleWriter{dst: &out, cxxfilt: fakeCxxFilt(t)}

	if _, err := writer.Write([]byte("frame _ZN8GTEngine4TickEv")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("Write() wrote partial line %q before flush", out.String())
	}
	if err := writer.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	want := "frame GTEngine::Tick()"
	if got := out.String(); got != want {
		t.Fatalf("writer output = %q, want %q", got, want)
	}
}

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

func fakeCxxFilt(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "c++filt")
	script := strings.Join([]string{
		"#!/bin/sh",
		"for arg do",
		"  case \"$arg\" in",
		"    _ZN8GTEngine4TickEv) echo 'GTEngine::Tick()' ;;",
		"    _ZZN3Foo3barEvEN3\\$_08__invokeEv) echo 'Foo::bar()::$_0::__invoke()' ;;",
		"    *) echo \"$arg\" ;;",
		"  esac",
		"done",
	}, "\n")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake c++filt: %v", err)
	}
	return path
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
