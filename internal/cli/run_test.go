package cli

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/promptctl/macklebox/internal/ui"
)

type result struct {
	code   int
	stdout string
	stderr string
}

func run(t *testing.T, argv ...string) result {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Main(argv, ui.IO{In: strings.NewReader(""), Out: &out, Err: &errOut})
	return result{code: code, stdout: out.String(), stderr: errOut.String()}
}

// --help prints usage to stdout and exits 0, reading no config.
func TestHelpGoesToStdoutAndExitsZero(t *testing.T) {
	for _, argv := range [][]string{{"-h"}, {"--help"}} {
		got := run(t, argv...)
		if got.code != ExitOK {
			t.Errorf("mackup %v exit = %d, want %d", argv, got.code, ExitOK)
		}
		if !strings.Contains(got.stdout, "Usage:") {
			t.Errorf("mackup %v stdout = %q, want a usage block", argv, got.stdout)
		}
		if got.stderr != "" {
			t.Errorf("mackup %v wrote %q to stderr, want nothing", argv, got.stderr)
		}
	}
}

// --version prints "Mackup <version>" to stdout and exits 0.
func TestVersionGoesToStdoutAndExitsZero(t *testing.T) {
	got := run(t, "--version")
	if got.code != ExitOK {
		t.Errorf("mackup --version exit = %d, want %d", got.code, ExitOK)
	}
	if !regexp.MustCompile(`^Mackup \S+\n$`).MatchString(got.stdout) {
		t.Errorf("mackup --version stdout = %q, want a single %q line", got.stdout, "Mackup <version>")
	}
	if got.stderr != "" {
		t.Errorf("mackup --version wrote %q to stderr, want nothing", got.stderr)
	}
}

// Supplying both force flags is rejected before config load, with that exact
// line on stderr and exit 1. The line is a literal contract token.
func TestConflictingForceFlagsAreRejected(t *testing.T) {
	for _, argv := range [][]string{
		{"--force", "--force-no", "backup"},
		{"-f", "--force-no", "list"},
		{"--force-no", "--force"},
	} {
		got := run(t, argv...)
		if got.code != ExitFatal {
			t.Errorf("mackup %v exit = %d, want %d", argv, got.code, ExitFatal)
		}
		if got.stderr != forceConflictMessage+"\n" {
			t.Errorf("mackup %v stderr = %q, want exactly %q", argv, got.stderr, forceConflictMessage+"\n")
		}
		if got.stdout != "" {
			t.Errorf("mackup %v wrote %q to stdout, want nothing", argv, got.stdout)
		}
	}
}

// --help and --version are the only paths that skip the gate, so they win over
// the force-flag conflict rather than being rejected by it.
func TestHelpAndVersionOutrankTheForceConflict(t *testing.T) {
	for _, argv := range [][]string{
		{"--help", "--force", "--force-no"},
		{"--version", "--force", "--force-no"},
	} {
		got := run(t, argv...)
		if got.code != ExitOK {
			t.Errorf("mackup %v exit = %d, want %d", argv, got.code, ExitOK)
		}
		if strings.Contains(got.stderr, forceConflictMessage) {
			t.Errorf("mackup %v rejected the force conflict, want the short-circuit to win", argv)
		}
	}
}

// An unrecognized subcommand produces a warning line identifying the unmatched
// argument, then the usage block — both on stderr.
func TestUnrecognizedSubcommandWarnsThenPrintsUsage(t *testing.T) {
	got := run(t, "frobnicate")
	if got.code != ExitFatal {
		t.Errorf("mackup frobnicate exit = %d, want %d", got.code, ExitFatal)
	}
	if got.stdout != "" {
		t.Errorf("mackup frobnicate wrote %q to stdout, want nothing", got.stdout)
	}
	if !strings.Contains(got.stderr, "frobnicate") {
		t.Errorf("mackup frobnicate stderr = %q, want the unmatched argument named", got.stderr)
	}
	if !strings.Contains(got.stderr, "Usage:") {
		t.Errorf("mackup frobnicate stderr = %q, want the usage block after the warning", got.stderr)
	}
	warning, _, _ := strings.Cut(got.stderr, "\n")
	if strings.Contains(warning, "Usage:") {
		t.Errorf("mackup frobnicate stderr starts with %q, want the warning line first", warning)
	}
}

// `show` requires <application>; the parser reports a usage error.
func TestShowWithoutApplicationIsAUsageError(t *testing.T) {
	got := run(t, "show")
	if got.code != ExitFatal {
		t.Errorf("mackup show exit = %d, want %d", got.code, ExitFatal)
	}
	if !strings.Contains(got.stderr, "<application>") {
		t.Errorf("mackup show stderr = %q, want the missing argument named", got.stderr)
	}
}

// A bare invocation is a usage display, not an error.
func TestBareInvocationShowsUsage(t *testing.T) {
	got := run(t)
	if got.code != ExitOK {
		t.Errorf("bare mackup exit = %d, want %d", got.code, ExitOK)
	}
	if !strings.Contains(got.stdout, "Usage:") {
		t.Errorf("bare mackup stdout = %q, want a usage block", got.stdout)
	}
}

// Every accepted subcommand reaches dispatch: it passes the parse and
// config-load steps and fails only because its behavior is not built yet. When
// the sync commands land, these become their real behavior tests.
func TestAcceptedSubcommandsReachDispatch(t *testing.T) {
	for _, argv := range [][]string{
		{"list"},
		{"show", "vim"},
		{"backup"},
		{"backup", "vim"},
		{"restore"},
		{"restore", "vim"},
		{"link"},
		{"link", "vim"},
		{"link", "install"},
		{"link", "install", "vim"},
		{"link", "uninstall"},
		{"link", "uninstall", "vim"},
	} {
		got := run(t, argv...)
		if strings.Contains(got.stderr, "Usage:") {
			t.Errorf("mackup %v was rejected as a usage error: %q", argv, got.stderr)
		}
		if !strings.Contains(got.stderr, "not implemented yet") {
			t.Errorf("mackup %v stderr = %q, want it to reach dispatch", argv, got.stderr)
		}
	}
}

// A cleanly-handled fatal error is a diagnostic on stderr with exit 1 and
// nothing on stdout (appspec/01 §6, the guarded regime).
func TestFatalErrorsLeaveStdoutEmpty(t *testing.T) {
	got := run(t, "list")
	if got.code != ExitFatal {
		t.Errorf("mackup list exit = %d, want %d", got.code, ExitFatal)
	}
	if got.stdout != "" {
		t.Errorf("mackup list wrote %q to stdout, want nothing on a fatal path", got.stdout)
	}
	if !strings.HasPrefix(got.stderr, "Error: ") {
		t.Errorf("mackup list stderr = %q, want an %q diagnostic", got.stderr, "Error: ")
	}
}
