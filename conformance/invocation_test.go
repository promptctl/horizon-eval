package conformance

import (
	"reflect"
	"strings"
	"testing"
)

// seeded is a small home tree used by cases that must leave home untouched.
var seeded = map[string]string{
	".vimrc":        "set nocompatible\n",
	".mackup.cfg":   "[storage]\nengine = file_system\npath = shared\n",
	".config/":      "",
	"shared/.vimrc": "set nocompatible\n",
}

// appspec/02 "Global options": --help prints the usage/help text to stdout and
// exits 0, with no other action and no config read. The exact wording is
// human-facing, so the checks are on the usage grammar it must contain.
func TestHelp(t *testing.T) {
	usage := Contains(
		"Usage:",
		"mackup [options] list",
		"mackup [options] show <application>",
		"mackup [options] backup [<application>]",
		"mackup [options] restore [<application>]",
		"mackup [options] link install [<application>]",
		"mackup [options] link uninstall [<application>]",
		"mackup [options] link [<application>]",
		"--help",
		"--version",
		"--force",
		"--force-no",
		"--root",
		"--dry-run",
		"--verbose",
		"--config-file",
	)
	RunCases(t, []Case{
		{
			Name: "long form", Args: []string{"--help"},
			Code: Exit(0), Stdout: usage, Stderr: Empty(),
			Home: seeded, HomeUnchanged: true,
		},
		{
			Name: "short form", Args: []string{"-h"},
			Code: Exit(0), Stdout: usage, Stderr: Empty(),
			Home: seeded, HomeUnchanged: true,
		},
		{
			// Ruling, recorded deliberately because appspec/02 can be read both
			// ways. "anything that does not match one of the listed usage lines
			// is a usage error" would make this argv an error, but --help is
			// specified as "Print the usage/help text to stdout and exit 0. No
			// other action", and the reference's parser behaves the same way: a
			// stray positional alongside --help is tolerated, a malformed or
			// unknown *option* is not. The two cases below pin both halves.
			Name: "with an extra positional", Args: []string{"--help", "frobnicate"},
			Code: Exit(0), Stdout: usage, Stderr: Empty(),
			Home: seeded, HomeUnchanged: true,
		},
		{
			Name: "but not alongside an unknown option", Args: []string{"--help", "--frobnicate"},
			Code: Exit(1), Stdout: Empty(), Stderr: InOrder("--frobnicate", "Usage:"),
			Home: seeded, HomeUnchanged: true,
		},
	})
}

// appspec/02: --version prints "Mackup <version>" to stdout and exits 0.
// appspec/00 Provenance fixes the version string for a build with no package
// metadata as the literal token "unknown", which is what a tree build is.
func TestVersion(t *testing.T) {
	RunCases(t, []Case{
		{
			Name: "prints the Mackup version line", Args: []string{"--version"},
			Code: Exit(0), Stdout: Exactly("Mackup unknown\n"), Stderr: Empty(),
			Home: seeded, HomeUnchanged: true,
		},
	})
}

// appspec/02 "Mutually exclusive force flags" and appspec/07's error table: the
// single literal line on stderr, exit 1, with no config read and no action.
func TestForceFlagsAreMutuallyExclusive(t *testing.T) {
	line := Exactly("Options --force and --force-no are mutually exclusive.\n")
	RunCases(t, []Case{
		{
			Name: "long forms", Args: []string{"--force", "--force-no", "backup"},
			Code: Exit(1), Stdout: Empty(), Stderr: line,
			Home: seeded, HomeUnchanged: true,
		},
		{
			Name: "short force with long force-no", Args: []string{"-f", "--force-no", "list"},
			Code: Exit(1), Stdout: Empty(), Stderr: line,
			Home: seeded, HomeUnchanged: true,
		},
		{
			// Rejected before config load, so the command named makes no
			// difference and neither does an unreadable config.
			Name: "before config load", Args: []string{"--force-no", "--force", "restore", "vim"},
			Code: Exit(1), Stdout: Empty(), Stderr: line,
			Home: seeded, HomeUnchanged: true,
		},
	})
}

// appspec/02 "Argument-parser behavior": argv matching none of the usage lines
// is a usage error — a warning naming the unmatched argument, then the usage
// block. appspec/07 routes parser usage and warning text to stderr.
func TestUsageErrors(t *testing.T) {
	RunCases(t, []Case{
		{
			Name: "unrecognized subcommand", Args: []string{"frobnicate"},
			Code: Exit(1), Stdout: Empty(),
			Stderr: InOrder("frobnicate", "Usage:"),
			Home:   seeded, HomeUnchanged: true,
		},
		{
			Name: "show with no application", Args: []string{"show"},
			Code: Exit(1), Stdout: Empty(), Stderr: Contains("Usage:"),
			Home: seeded, HomeUnchanged: true,
		},
		{
			Name: "too many positionals", Args: []string{"show", "vim", "git"},
			Code: Exit(1), Stdout: Empty(), Stderr: InOrder("git", "Usage:"),
			Home: seeded, HomeUnchanged: true,
		},
		{
			Name: "list takes no application", Args: []string{"list", "vim"},
			Code: Exit(1), Stdout: Empty(), Stderr: InOrder("vim", "Usage:"),
			Home: seeded, HomeUnchanged: true,
		},
		{
			Name: "unrecognized option", Args: []string{"--frobnicate", "list"},
			Code: Exit(1), Stdout: Empty(), Stderr: InOrder("--frobnicate", "Usage:"),
			Home: seeded, HomeUnchanged: true,
		},
		{
			// A malformed --help token matches no usage line, so it must not
			// short-circuit to a successful help display.
			Name: "help given a value", Args: []string{"--help=1"},
			Code: Exit(1), Stdout: Empty(), Stderr: InOrder("--help", "Usage:"),
			Home: seeded, HomeUnchanged: true,
		},
		{
			// A typo'd option is reported even alongside --version.
			Name: "unrecognized option with version", Args: []string{"--version", "--frobnicate"},
			Code: Exit(1), Stdout: Empty(), Stderr: InOrder("--frobnicate", "Usage:"),
			Home: seeded, HomeUnchanged: true,
		},
		{
			// An empty application key must not read as "no application given",
			// which would silently widen the run to the whole configured set.
			Name: "empty application key", Args: []string{"backup", ""},
			Code: Exit(1), Stdout: Empty(), Stderr: Contains("Usage:"),
			Home: seeded, HomeUnchanged: true,
		},
		{
			// An empty config path must not read as "no --config-file given",
			// which would fall back to default discovery.
			Name: "empty config path", Args: []string{"--config-file=", "list"},
			Code: Exit(1), Stdout: Empty(), Stderr: Contains("Usage:"),
			Home: seeded, HomeUnchanged: true,
		},
	})
}

// appspec/02 "Argument-parser behavior": a bare invocation is a usage display.
func TestBareInvocation(t *testing.T) {
	RunCases(t, []Case{
		{
			Name: "shows usage", Args: nil,
			Code: Exit(0), Stdout: Contains("Usage:"),
			Home: seeded, HomeUnchanged: true,
		},
	})
}

// appspec/02 "Invocation forms": short and long forms are interchangeable and
// produce identical behavior. Checked at the boundary by running both and
// comparing everything a caller observes.
func TestShortAndLongFormsAreInterchangeable(t *testing.T) {
	pairs := []struct {
		name        string
		short, long []string
	}{
		{"help", []string{"-h"}, []string{"--help"}},
		{"verbose given a value", []string{"-v=yes", "list"}, []string{"--verbose=yes", "list"}},
		{"help given a value", []string{"-h=1"}, []string{"--help=1"}},
		{"config path attached", []string{"-c=cfg", "list"}, []string{"--config-file=cfg", "list"}},
		{"config path separate", []string{"-c", "cfg", "list"}, []string{"--config-file", "cfg", "list"}},
		{"config path missing", []string{"-c"}, []string{"--config-file"}},
	}
	for _, pair := range pairs {
		t.Run(pair.name, func(t *testing.T) {
			short := Run(t, Invocation{Args: pair.short})
			long := Run(t, Invocation{Args: pair.long})
			if short.Code != long.Code {
				t.Errorf("exit: %v = %d, %v = %d", pair.short, short.Code, pair.long, long.Code)
			}
			if short.PlainStdout() != long.PlainStdout() {
				t.Errorf("stdout: %v = %q, %v = %q",
					pair.short, short.PlainStdout(), pair.long, long.PlainStdout())
			}
			if short.PlainStderr() != long.PlainStderr() {
				t.Errorf("stderr: %v = %q, %v = %q",
					pair.short, short.PlainStderr(), pair.long, long.PlainStderr())
			}
		})
	}
}

// appspec/02 "Command dispatch order": only --help and --version bypass the
// config-load gate, so every other command reaches it. Until the resolvers land
// there is no config failure to induce, so what this pins is the split itself:
// the two short-circuit paths succeed, every other command does not.
func TestOnlyHelpAndVersionBypassTheConfigGate(t *testing.T) {
	for _, args := range [][]string{
		{"list"},
		{"show", "vim"},
		{"backup"},
		{"backup", "vim"},
		{"restore"},
		{"link"},
		{"link", "install"},
		{"link", "uninstall"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			r := Run(t, Invocation{Args: args, Home: seeded})
			// Exactly 1, appspec/02's fatal code — not merely non-zero, which a
			// panicking binary would also satisfy.
			if r.Code != 1 {
				t.Errorf("exit = %d, want 1 — no command is implemented yet, so none may report success", r.Code)
			}
			if r.PlainStdout() != "" {
				t.Errorf("stdout = %q, want empty", r.PlainStdout())
			}
			if r.PlainStderr() == "" {
				t.Error("stderr empty, want a diagnostic")
			}
			AssertUnchanged(t, r.Home, r.HomeBefore)
		})
	}
}

// A case may name the spec's other environment variables, and they reach the
// program. HOME is the rig's, and Run refuses a case that tries to take it.
func TestEnvOverridesReachTheProgramButHomeIsTheRigs(t *testing.T) {
	env, err := childEnv("/tmp/home", map[string]string{
		"XDG_CONFIG_HOME": "/tmp/home/.config",
		"MACKUP_CONFIG":   "/tmp/home/other.cfg",
	})
	if err != nil {
		t.Fatalf("childEnv: %v", err)
	}
	want := []string{
		"HOME=/tmp/home",
		"MACKUP_CONFIG=/tmp/home/other.cfg",
		"PATH=" + fixedPATH,
		"XDG_CONFIG_HOME=/tmp/home/.config",
	}
	if !reflect.DeepEqual(env, want) {
		t.Errorf("childEnv = %q, want %q", env, want)
	}

	if _, err := childEnv("/tmp/home", map[string]string{"HOME": "/elsewhere"}); err == nil {
		t.Error("childEnv accepted a HOME override; the tree assertions are anchored to the rig's home")
	}
}
