package cli

import "testing"

// Every invocation form appspec/02-invocation.md lists must be accepted, and
// options must be honored before or after the subcommand.
func TestParseAcceptsEveryInvocationForm(t *testing.T) {
	tests := []struct {
		argv []string
		want Options
	}{
		{[]string{"list"}, Options{Command: CmdList}},
		{[]string{"show", "vim"}, Options{Command: CmdShow, Application: "vim"}},
		{[]string{"backup"}, Options{Command: CmdBackup}},
		{[]string{"backup", "git"}, Options{Command: CmdBackup, Application: "git"}},
		{[]string{"restore"}, Options{Command: CmdRestore}},
		{[]string{"restore", "sublime-text-3"}, Options{Command: CmdRestore, Application: "sublime-text-3"}},
		{[]string{"link"}, Options{Command: CmdLink}},
		{[]string{"link", "vim"}, Options{Command: CmdLink, Application: "vim"}},
		{[]string{"link", "install"}, Options{Command: CmdLinkInstall}},
		{[]string{"link", "install", "vim"}, Options{Command: CmdLinkInstall, Application: "vim"}},
		{[]string{"link", "uninstall"}, Options{Command: CmdLinkUninstall}},
		{[]string{"link", "uninstall", "vim"}, Options{Command: CmdLinkUninstall, Application: "vim"}},
		{[]string{"-h"}, Options{Help: true}},
		{[]string{"--help"}, Options{Help: true}},
		{[]string{"--version"}, Options{Version: true}},
		{nil, Options{Command: CmdNone}},
		// Options before and after the subcommand.
		{[]string{"--dry-run", "backup", "vim"}, Options{DryRun: true, Command: CmdBackup, Application: "vim"}},
		{[]string{"backup", "vim", "--dry-run"}, Options{DryRun: true, Command: CmdBackup, Application: "vim"}},
	}
	for _, tt := range tests {
		got, err := Parse(tt.argv)
		if err != nil {
			t.Errorf("Parse(%q) returned error %v, want none", tt.argv, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Parse(%q) = %+v, want %+v", tt.argv, got, tt.want)
		}
	}
}

// Short and long forms are interchangeable and produce identical behavior.
func TestParseShortAndLongFormsAreEquivalent(t *testing.T) {
	pairs := [][2][]string{
		{{"-f", "backup"}, {"--force", "backup"}},
		{{"-r", "backup"}, {"--root", "backup"}},
		{{"-n", "backup"}, {"--dry-run", "backup"}},
		{{"-v", "backup"}, {"--verbose", "backup"}},
		{{"-c", "/home/u/.mackup.cfg", "backup"}, {"--config-file", "/home/u/.mackup.cfg", "backup"}},
		{{"-c/home/u/.mackup.cfg", "backup"}, {"--config-file=/home/u/.mackup.cfg", "backup"}},
		{{"-c=/home/u/.mackup.cfg", "backup"}, {"--config-file=/home/u/.mackup.cfg", "backup"}},
		{{"-h"}, {"--help"}},
	}
	for _, pair := range pairs {
		short, err := Parse(pair[0])
		if err != nil {
			t.Errorf("Parse(%q): %v", pair[0], err)
			continue
		}
		long, err := Parse(pair[1])
		if err != nil {
			t.Errorf("Parse(%q): %v", pair[1], err)
			continue
		}
		if short != long {
			t.Errorf("Parse(%q) = %+v, Parse(%q) = %+v; want identical", pair[0], short, pair[1], long)
		}
	}
}

// A token that fails validation must not leave any of its flags set, or the
// help/version short-circuit would swallow the error.
func TestParseAppliesNothingFromAMalformedShortCluster(t *testing.T) {
	for _, argv := range [][]string{{"-h=1"}, {"-fz", "list"}, {"-vc"}} {
		got, err := Parse(argv)
		if err == nil {
			t.Errorf("Parse(%q) = %+v, want a usage error", argv, got)
			continue
		}
		if got != (Options{}) {
			t.Errorf("Parse(%q) set %+v, want nothing applied", argv, got)
		}
	}
}

func TestParseStacksShortOptions(t *testing.T) {
	got, err := Parse([]string{"-fnv", "backup"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := Options{Force: true, DryRun: true, Verbose: true, Command: CmdBackup}
	if got != want {
		t.Errorf("Parse(-fnv backup) = %+v, want %+v", got, want)
	}
}

func TestParseCollectsEveryOption(t *testing.T) {
	argv := []string{"--force", "--root", "--dry-run", "--verbose", "--config-file", "cfg", "restore", "vim"}
	got, err := Parse(argv)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := Options{
		Force: true, Root: true, DryRun: true, Verbose: true,
		ConfigFile: "cfg", Command: CmdRestore, Application: "vim",
	}
	if got != want {
		t.Errorf("Parse(%q) = %+v, want %+v", argv, got, want)
	}
}

// Both force flags parse; rejecting the combination is Run's job, at a defined
// point in the dispatch order.
func TestParseAcceptsBothForceFlags(t *testing.T) {
	got, err := Parse([]string{"--force", "--force-no", "backup"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !got.Force || !got.ForceNo {
		t.Errorf("Parse(--force --force-no backup) = %+v, want both force flags set", got)
	}
}

// Anything matching none of the listed usage lines is a usage error.
func TestParseRejectsNonMatchingArgv(t *testing.T) {
	tests := []struct {
		name string
		argv []string
	}{
		{"unrecognized command", []string{"frobnicate"}},
		{"show without application", []string{"show"}},
		{"show with two applications", []string{"show", "vim", "git"}},
		{"list with an application", []string{"list", "vim"}},
		{"backup with two applications", []string{"backup", "vim", "git"}},
		{"link install with two applications", []string{"link", "install", "vim", "git"}},
		{"unrecognized long option", []string{"--frobnicate", "list"}},
		{"unrecognized short option", []string{"-z", "list"}},
		{"config-file without a value", []string{"list", "--config-file"}},
		{"short config-file without a value", []string{"list", "-c"}},
		{"valueless option given a value", []string{"--verbose=yes", "list"}},
		// A malformed --help/--version token matches no usage line, so it must
		// not short-circuit the run the way a well-formed one does.
		{"help given a value", []string{"--help=1"}},
		{"version given a value", []string{"--version=1"}},
		// An empty config path must not be silently indistinguishable from
		// "no --config-file given", which would fall back to default discovery.
		{"empty long config-file value", []string{"--config-file=", "list"}},
		{"empty short config-file value", []string{"-c", "", "list"}},
		{"empty short config-file = value", []string{"-c=", "list"}},
		// Short and long forms must agree: a malformed short help token is a
		// usage error too, and must not half-apply and then exit 0.
		{"short help given a value", []string{"-h=1"}},
		{"short flag given a value", []string{"-v=yes", "list"}},
		// An empty application key would silently widen a single-app run to the
		// whole configured set.
		{"empty application key", []string{"backup", ""}},
		{"empty application key for show", []string{"show", ""}},
		{"empty escaped application key", []string{"backup", "--", ""}},
	}
	for _, tt := range tests {
		if _, err := Parse(tt.argv); err == nil {
			t.Errorf("%s: Parse(%q) returned no error, want a usage error", tt.name, tt.argv)
		}
	}
}

// --help and --version are specified to print and exit taking "no other
// action", so they short-circuit a non-matching grammar.
func TestParseHelpAndVersionShortCircuitGrammarErrors(t *testing.T) {
	for _, argv := range [][]string{
		{"--help", "frobnicate"},
		{"--version", "frobnicate"},
		{"show", "--help"},
		{"list", "extra", "--version"},
	} {
		got, err := Parse(argv)
		if err != nil {
			t.Errorf("Parse(%q) returned error %v, want none", argv, err)
			continue
		}
		if !got.Help && !got.Version {
			t.Errorf("Parse(%q) = %+v, want Help or Version set", argv, got)
		}
	}
}

// They do not rescue a malformed option, though: a typo'd option matches no
// usage line and must not be silently accepted.
func TestParseHelpAndVersionDoNotRescueOptionErrors(t *testing.T) {
	for _, argv := range [][]string{
		{"--frobnicate", "--help"},
		{"--help", "--frobnicate"},
		{"--version", "--frobnicate"},
		{"-z", "--help"},
	} {
		if got, err := Parse(argv); err == nil {
			t.Errorf("Parse(%q) = %+v, want a usage error", argv, got)
		}
	}
}

// "--" ends option parsing, so an application key may begin with a dash.
func TestParseDoubleDashEndsOptions(t *testing.T) {
	got, err := Parse([]string{"backup", "--", "-weird-app"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := Options{Command: CmdBackup, Application: "-weird-app"}
	if got != want {
		t.Errorf("Parse = %+v, want %+v", got, want)
	}
}

// "--" ends option parsing and nothing else: it does not change how positionals
// bind to the grammar, so a subcommand word after it is still the subcommand.
func TestParseDoubleDashDoesNotChangePositionalBinding(t *testing.T) {
	tests := []struct {
		argv []string
		want Options
	}{
		{[]string{"link", "--", "install"}, Options{Command: CmdLinkInstall}},
		{[]string{"link", "--", "uninstall", "vim"}, Options{Command: CmdLinkUninstall, Application: "vim"}},
		{[]string{"--", "link", "install"}, Options{Command: CmdLinkInstall}},
		{[]string{"--", "list"}, Options{Command: CmdList}},
		{[]string{"--", "backup", "vim"}, Options{Command: CmdBackup, Application: "vim"}},
	}
	for _, tt := range tests {
		got, err := Parse(tt.argv)
		if err != nil {
			t.Errorf("Parse(%q): %v", tt.argv, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Parse(%q) = %+v, want %+v", tt.argv, got, tt.want)
		}
	}
}
