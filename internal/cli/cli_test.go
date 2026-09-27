package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/promptctl/macklebox/internal/version"
)

type capture struct {
	out, err bytes.Buffer
	loaded   bool
	code     int
}

// runCLI runs argv with a recording config loader, so a test can assert both the
// streams and whether the run reached the config-load step.
func runCLI(argv []string, loadErr error) *capture {
	c := &capture{}
	load := func(Options) (config, error) {
		c.loaded = true
		return config{}, loadErr
	}
	c.code = run(argv, Streams{Out: &c.out, Err: &c.err, In: strings.NewReader("")}, load)
	return c
}

func TestHelpPrintsUsageToStdoutAndExitsZero(t *testing.T) {
	for _, argv := range [][]string{{"-h"}, {"--help"}} {
		c := runCLI(argv, nil)
		if c.code != ExitOK {
			t.Errorf("%q: exit %d, want %d", argv, c.code, ExitOK)
		}
		if c.out.String() != Help {
			t.Errorf("%q: stdout = %q, want the help text", argv, c.out.String())
		}
		if c.err.Len() != 0 {
			t.Errorf("%q: stderr = %q, want empty", argv, c.err.String())
		}
		if c.loaded {
			t.Errorf("%q: read config, want no config read", argv)
		}
	}
}

func TestVersionPrintsVersionLineToStdoutAndExitsZero(t *testing.T) {
	c := runCLI([]string{"--version"}, nil)
	if c.code != ExitOK {
		t.Errorf("exit %d, want %d", c.code, ExitOK)
	}
	want := "Mackup " + version.String() + "\n"
	if c.out.String() != want {
		t.Errorf("stdout = %q, want %q", c.out.String(), want)
	}
	if c.err.Len() != 0 {
		t.Errorf("stderr = %q, want empty", c.err.String())
	}
	if c.loaded {
		t.Error("read config, want no config read")
	}
}

// The literal line, on stderr, exit 1, before config load.
func TestConflictingForceFlagsAreRejectedBeforeConfigLoad(t *testing.T) {
	c := runCLI([]string{"--force", "--force-no", "backup"}, nil)
	if c.code != ExitFatal {
		t.Errorf("exit %d, want %d", c.code, ExitFatal)
	}
	if got, want := c.err.String(), ForceConflictMessage+"\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
	if c.out.Len() != 0 {
		t.Errorf("stdout = %q, want empty", c.out.String())
	}
	if c.loaded {
		t.Error("read config, want the conflict rejected before config load")
	}
}

func TestConflictingForceFlagsAcceptShortForm(t *testing.T) {
	c := runCLI([]string{"-f", "--force-no", "list"}, nil)
	if c.code != ExitFatal || c.err.String() != ForceConflictMessage+"\n" {
		t.Errorf("exit %d stderr %q, want %d and the conflict line", c.code, c.err.String(), ExitFatal)
	}
}

// An unrecognized subcommand: a warning line naming the unmatched argument,
// then the usage block, both on stderr.
func TestUnrecognizedCommandWarnsThenPrintsUsage(t *testing.T) {
	c := runCLI([]string{"frobnicate"}, nil)
	if c.code != ExitFatal {
		t.Errorf("exit %d, want %d", c.code, ExitFatal)
	}
	if c.out.Len() != 0 {
		t.Errorf("stdout = %q, want empty", c.out.String())
	}
	stderr := c.err.String()
	if !strings.Contains(stderr, "frobnicate") {
		t.Errorf("stderr = %q, want it to name the unmatched argument", stderr)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("stderr = %q, want it to include the usage block", stderr)
	}
	if strings.Index(stderr, "frobnicate") > strings.Index(stderr, "Usage:") {
		t.Errorf("stderr = %q, want the warning line before the usage block", stderr)
	}
	if c.loaded {
		t.Error("read config, want a usage error rejected at parse time")
	}
}

func TestShowWithoutApplicationIsAUsageError(t *testing.T) {
	c := runCLI([]string{"show"}, nil)
	if c.code != ExitFatal {
		t.Errorf("exit %d, want %d", c.code, ExitFatal)
	}
	if !strings.Contains(c.err.String(), "Usage:") {
		t.Errorf("stderr = %q, want the usage block", c.err.String())
	}
}

// A bare invocation is a usage display: the usage block, exit 0.
func TestBareInvocationShowsUsage(t *testing.T) {
	c := runCLI(nil, nil)
	if c.code != ExitOK {
		t.Errorf("exit %d, want %d", c.code, ExitOK)
	}
	if c.out.String() != Help {
		t.Errorf("stdout = %q, want the help text", c.out.String())
	}
	if c.loaded {
		t.Error("read config, want no config read for a usage display")
	}
}

// Every command except --help and --version loads config before dispatching.
func TestEveryCommandLoadsConfigBeforeDispatch(t *testing.T) {
	for _, argv := range [][]string{
		{"list"},
		{"show", "vim"},
		{"backup"},
		{"restore", "vim"},
		{"link"},
		{"link", "install"},
		{"link", "uninstall", "vim"},
	} {
		c := runCLI(argv, nil)
		if !c.loaded {
			t.Errorf("%q: did not load config, want config loaded before dispatch", argv)
		}
	}
}

// A fatal config error terminates the run for any command, with a diagnostic on
// stderr, nothing on stdout, and a non-zero exit.
func TestFatalConfigErrorAbortsEveryCommand(t *testing.T) {
	for _, argv := range [][]string{{"list"}, {"show", "vim"}, {"backup", "vim"}} {
		c := runCLI(argv, errors.New("Error: Unable to find the storage folder: /nope"))
		if c.code == ExitOK {
			t.Errorf("%q: exit %d, want non-zero", argv, c.code)
		}
		if c.out.Len() != 0 {
			t.Errorf("%q: stdout = %q, want empty", argv, c.out.String())
		}
		if !strings.Contains(c.err.String(), "Unable to find the storage folder") {
			t.Errorf("%q: stderr = %q, want the config diagnostic", argv, c.err.String())
		}
	}
}

// appspec/07 specifies some fatal diagnostics as a single "Error: ... Aborting."
// line and others as multi-line messages carrying no such prefix, so the run
// must print the error's own text rather than reshaping it.
func TestFatalConfigErrorIsPrintedVerbatim(t *testing.T) {
	multiline := "Unable to find your Dropbox =(\nhttps://example.invalid/doc"
	c := runCLI([]string{"list"}, errors.New(multiline))
	if got, want := c.err.String(), multiline+"\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

// A malformed --help/--version token, or a typo'd option alongside one, matches
// no usage line, so it is a usage error and must not short-circuit to a
// successful help or version display.
func TestMalformedHelpAndVersionTokensAreUsageErrors(t *testing.T) {
	for _, argv := range [][]string{
		{"--help=1"},
		{"--version=1"},
		{"--frobnicate", "--help"},
		{"--version", "--frobnicate"},
	} {
		c := runCLI(argv, nil)
		if c.code != ExitFatal {
			t.Errorf("%q: exit %d, want %d", argv, c.code, ExitFatal)
		}
		if c.out.Len() != 0 {
			t.Errorf("%q: stdout = %q, want empty", argv, c.out.String())
		}
		if !strings.Contains(c.err.String(), "Usage:") {
			t.Errorf("%q: stderr = %q, want the usage block", argv, c.err.String())
		}
	}
}

// Until the commands land, a recognized command must not report success.
func TestUnimplementedCommandDoesNotReportSuccess(t *testing.T) {
	c := runCLI([]string{"list"}, nil)
	if c.code == ExitOK {
		t.Errorf("exit %d, want non-zero while the command is unimplemented", c.code)
	}
	if c.out.Len() != 0 {
		t.Errorf("stdout = %q, want empty", c.out.String())
	}
	if c.err.Len() == 0 {
		t.Error("stderr empty, want a diagnostic")
	}
}
