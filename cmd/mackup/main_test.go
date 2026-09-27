package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binary is the compiled mackup command, built once for this package's tests.
var binary string

// TestMain builds the command so the tests below observe it where the spec
// observes it: at the process boundary — stdout, stderr, exit code. The full
// black-box conformance rig (a throwaway home, the whole error table) is its own
// ticket; this is the smoke test that the boundary is wired up at all.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "mackup-bin")
	if err != nil {
		panic(err)
	}

	binary = filepath.Join(dir, "mackup")
	build := exec.Command("go", "build", "-o", binary, ".")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		os.RemoveAll(dir)
		panic(err)
	}

	// os.Exit skips deferred calls, so the cleanup is explicit.
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// invoke runs the built command and returns what a caller observes.
func invoke(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	cmd.Stdin = strings.NewReader("")
	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		code = 0
	case errors.As(err, &exitErr):
		code = exitErr.ExitCode()
	default:
		t.Fatalf("running %v: %v", args, err)
	}
	return out.String(), errb.String(), code
}

func TestHelpGoesToStdoutWithExitZero(t *testing.T) {
	for _, flag := range []string{"-h", "--help"} {
		stdout, stderr, code := invoke(t, flag)
		if code != 0 {
			t.Errorf("mackup %s: exit %d, want 0", flag, code)
		}
		if !strings.Contains(stdout, "Usage:") {
			t.Errorf("mackup %s: stdout = %q, want the usage block", flag, stdout)
		}
		if stderr != "" {
			t.Errorf("mackup %s: stderr = %q, want empty", flag, stderr)
		}
	}
}

func TestVersionGoesToStdoutWithExitZero(t *testing.T) {
	stdout, stderr, code := invoke(t, "--version")
	if code != 0 {
		t.Errorf("exit %d, want 0", code)
	}
	// Built from the tree, so no package version is available: the stable
	// fallback token (appspec/00-overview.md, Provenance).
	if stdout != "Mackup unknown\n" {
		t.Errorf("stdout = %q, want %q", stdout, "Mackup unknown\n")
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestConflictingForceFlagsGoToStderrWithExitOne(t *testing.T) {
	stdout, stderr, code := invoke(t, "--force", "--force-no", "backup")
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
	if stderr != "Options --force and --force-no are mutually exclusive.\n" {
		t.Errorf("stderr = %q, want the literal conflict line", stderr)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
}

func TestUnrecognizedCommandGoesToStderr(t *testing.T) {
	stdout, stderr, code := invoke(t, "frobnicate")
	if code == 0 {
		t.Error("exit 0, want non-zero")
	}
	if !strings.Contains(stderr, "frobnicate") || !strings.Contains(stderr, "Usage:") {
		t.Errorf("stderr = %q, want the unmatched argument then the usage block", stderr)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
}

func TestBareInvocationShowsUsage(t *testing.T) {
	stdout, _, code := invoke(t)
	if code != 0 {
		t.Errorf("exit %d, want 0", code)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("stdout = %q, want the usage block", stdout)
	}
}
